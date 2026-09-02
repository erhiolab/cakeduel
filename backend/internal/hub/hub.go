package hub

import (
	"cakeduel-backend/internal/game"
	"cakeduel-backend/internal/logger"
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	redisOnlinePrefix = "cakeduel:online:"
)

// Hub 房间中心
type Hub struct {
	mu             sync.Mutex
	rooms          map[string]*Room
	matching       []*Client
	clientsByToken map[string]*Client
	gameConfig     game.GameConfig
	roomCodeLen    int
	matchTimeout   time.Duration
	disconnectGrace time.Duration
	rdb            redis.Cmdable
	closed         bool
}

// NewHub 创建房间中心
func NewHub(cfg game.GameConfig, rdb redis.Cmdable) *Hub {
	return &Hub{
		rooms:           make(map[string]*Room),
		clientsByToken:  make(map[string]*Client),
		gameConfig:      cfg,
		roomCodeLen:     6,
		matchTimeout:    60 * time.Minute,
		disconnectGrace: 45 * time.Second,
		rdb:             rdb,
	}
}

// SetRoomCodeLen 设置房间码长度
func (h *Hub) SetRoomCodeLen(n int) {
	if n >= 4 && n <= 12 {
		h.roomCodeLen = n
	}
}

// SetMatchTimeout 设置匹配超时
func (h *Hub) SetMatchTimeout(d time.Duration) {
	h.matchTimeout = d
}

// SetDisconnectGrace 设置断线重连宽容期
func (h *Hub) SetDisconnectGrace(d time.Duration) {
	h.disconnectGrace = d
}

// registerClient 注册连接; 若 token 对应断线席位则恢复
func (h *Hub) registerClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clientsByToken[c.Token]; ok && old.disconnected && old.room != nil {
		room := old.room
		idx := old.playerIndex
		if room.Clients[idx] == old {
			// 恢复席位
			room.Clients[idx] = c
			c.room = room
			c.playerIndex = idx
			c.Name = old.Name
			c.matching = false
			old.room = nil
			old.disconnected = false
			room.stopDisconnectTimer()
			h.clientsByToken[c.Token] = c
			h.markOnline(c)
			// 通知对手
			if other := room.otherClient(c); other != nil {
				other.Send(ServerMessage{Type: "player_reconnected", PlayerIndex: idx, Players: room.playersInfo()})
			}
			// 恢复状态: 对局中下发最新状态, 否则回到大厅
			if room.GameStarted && room.Game != nil {
				room.broadcastGameState(nil)
			} else {
				c.Send(ServerMessage{Type: "room_joined", RoomCode: room.Code, PlayerIndex: idx, Players: room.playersInfo()})
			}
			logger.Log.Info("玩家重连成功", zap.String("room", room.Code), zap.Int("player", idx))
			return
		}
	}
	h.clientsByToken[c.Token] = c
}

// markOnline 异步刷新 Redis 在线标记(不阻塞主流程, Redis 不可用时忽略)
func (h *Hub) markOnline(c *Client) {
	if h.rdb == nil || c.Token == "" {
		return
	}
	token, name := c.Token, c.Name
	go func() {
		defer func() {
			// 防御: Redis 异常时不应拖垮整个进程
			if r := recover(); r != nil {
				logger.Log.Warn("Redis 在线标记写入异常", zap.Any("panic", r))
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = h.rdb.Set(ctx, redisOnlinePrefix+token, name, 2*time.Minute).Err()
	}()
}

// unmarkOnline 异步清除在线标记
func (h *Hub) unmarkOnline(c *Client) {
	if h.rdb == nil || c.Token == "" {
		return
	}
	token := c.Token
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Warn("Redis 在线标记清除异常", zap.Any("panic", r))
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = h.rdb.Del(ctx, redisOnlinePrefix+token).Err()
	}()
}

// HandleMessage 处理客户端消息
func (h *Hub) HandleMessage(c *Client, msg *ClientMessage) {
	switch msg.Type {
	case "create_room":
		h.CreateRoom(c, msg)
	case "join_room":
		h.JoinRoom(c, msg)
	case "start_game":
		h.StartGame(c)
	case "action":
		h.HandleAction(c, msg.Action)
	case "rematch":
		h.Rematch(c)
	case "chat":
		h.Chat(c, msg.Text)
	case "leave":
		h.Leave(c, true)
	case "ping":
		c.Send(ServerMessage{Type: "pong"})
	}
}

// Chat 房间内聊天
func (h *Hub) Chat(c *Client, text string) {
	room := c.room
	if room == nil {
		c.sendError("你不在任何房间中")
		return
	}
	text = sanitizeName(text)
	if text == "" {
		return
	}
	runes := []rune(text)
	if len(runes) > 200 {
		runes = runes[:200]
		text = string(runes)
	}
	msg := ServerMessage{Type: "chat", From: c.playerIndex, Name: c.Name, Text: text}
	for _, cl := range room.Clients {
		if cl != nil && cl.room == room {
			cl.Send(msg)
		}
	}
}

// CreateRoom 创建房间
func (h *Hub) CreateRoom(c *Client, msg *ClientMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		c.sendError("服务器正在关闭")
		return
	}
	if c.room != nil || c.matching {
		c.sendError("你已在房间或匹配队列中")
		return
	}
	name := sanitizeName(msg.Name)
	c.Name = name
	h.markOnline(c)
	mode := msg.Mode
	if mode != "private" && mode != "random" {
		mode = "private"
	}

	if mode == "random" {
		// 尝试与等待中的随机玩家配对(内存队列, 单实例稳定)
		waiter := h.popMemoryMatch(c)
		if waiter != nil {
			h.pairRandom(c, waiter)
			return
		}
		// 进入等待队列
		c.matching = true
		h.matching = append(h.matching, c)
		c.Send(ServerMessage{Type: "matching", Message: "正在匹配对手^"})
		go h.matchTimeoutChecker(c)
		return
	}

	// 私有房间
	room := h.newRoom("private")
	room.addClient(c, 0)
	h.rooms[room.Code] = room
	c.Send(ServerMessage{
		Type:        "room_joined",
		RoomCode:    room.Code,
		PlayerIndex: 0,
		Players:     room.playersInfo(),
	})
}

// popMemoryMatch 从内存队列取等待者
func (h *Hub) popMemoryMatch(self *Client) *Client {
	for i := len(h.matching) - 1; i >= 0; i-- {
		waiter := h.matching[i]
		h.matching = append(h.matching[:i], h.matching[i+1:]...)
		if waiter == self || waiter.room != nil || waiter.disconnected {
			continue
		}
		waiter.matching = false
		return waiter
	}
	return nil
}

// pairRandom 随机匹配配对(需持有 hub 锁)
func (h *Hub) pairRandom(c, waiter *Client) {
	room := h.newRoom("random")
	room.addClient(c, 0)
	room.addClient(waiter, 1)
	h.rooms[room.Code] = room
	h.markOnline(c)
	h.markOnline(waiter)
	h.announceRoom(room)
	room.announceJoined()
	go room.waitForStart()
	logger.Log.Info("随机匹配成功", zap.String("room", room.Code), zap.String("p1", c.Name), zap.String("p2", waiter.Name))
}

// matchTimeoutChecker 随机匹配超时检查
func (h *Hub) matchTimeoutChecker(c *Client) {
	timer := time.NewTimer(h.matchTimeout)
	defer timer.Stop()
	<-timer.C
	h.mu.Lock()
	defer h.mu.Unlock()
	if !c.matching || c.room != nil {
		return
	}
	// 内存队列中移除
	for i, waiter := range h.matching {
		if waiter == c {
			h.matching = append(h.matching[:i], h.matching[i+1:]...)
			break
		}
	}
	c.matching = false
	h.unmarkOnline(c)
	c.Send(ServerMessage{Type: "match_timeout", Message: "匹配超时, 请重试"})
}

// JoinRoom 加入房间
func (h *Hub) JoinRoom(c *Client, msg *ClientMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		c.sendError("服务器正在关闭")
		return
	}
	if c.room != nil || c.matching {
		c.sendError("你已在房间或匹配队列中")
		return
	}
	code := sanitizeCode(msg.Code)
	room, ok := h.rooms[code]
	if !ok {
		c.sendError("房间不存在或已关闭")
		return
	}
	if room.GameStarted {
		c.sendError("对局已经开始, 无法加入")
		return
	}
	if room.Clients[0] == nil || room.Clients[1] == nil {
		idx := 0
		if room.Clients[0] != nil {
			idx = 1
		}
		room.addClient(c, idx)
		c.Name = sanitizeName(msg.Name)
		h.markOnline(c)
		c.Send(ServerMessage{
			Type:        "room_joined",
			RoomCode:    room.Code,
			PlayerIndex: idx,
			Players:     room.playersInfo(),
		})
		h.announceRoom(room)
		go room.waitForStart()
		return
	}
	c.sendError("房间已满")
}

// StartGame 开始游戏
func (h *Hub) StartGame(c *Client) {
	h.mu.Lock()
	room := c.room
	if room == nil {
		h.mu.Unlock()
		c.sendError("你不在任何房间中")
		return
	}
	if room.Clients[0] == nil || room.Clients[1] == nil {
		h.mu.Unlock()
		c.sendError("等待对手加入…")
		return
	}
	if room.GameStarted {
		h.mu.Unlock()
		c.sendError("对局已经开始")
		return
	}
	room.startGame()
	h.mu.Unlock()
}

// HandleAction 处理游戏动作
func (h *Hub) HandleAction(c *Client, action *ActionMsg) {
	h.mu.Lock()
	room := c.room
	if room == nil || !room.GameStarted || room.Game == nil {
		h.mu.Unlock()
		c.sendError("当前没有进行中的对局")
		return
	}
	room.mu.Lock()
	if room.paused() {
		room.mu.Unlock()
		h.mu.Unlock()
		c.sendError("对方已掉线, 对局已暂停, 等待重连…")
		return
	}
	room.applyAction(c, action)
	room.mu.Unlock()
	h.mu.Unlock()
}

// Rematch 再来一局
func (h *Hub) Rematch(c *Client) {
	h.mu.Lock()
	room := c.room
	if room == nil {
		h.mu.Unlock()
		c.sendError("你不在任何房间中")
		return
	}
	if room.Game == nil || room.Game.GameEnded == nil {
		h.mu.Unlock()
		c.sendError("对局尚未结束")
		return
	}
	// 记录本方同意; 双方都同意才开新局
	room.rematchVotes[c.playerIndex] = true
	if room.rematchVotes[0] && room.rematchVotes[1] {
		logger.Log.Info("双方同意再来一局", zap.String("room", room.Code))
		room.startGame()
		h.mu.Unlock()
		return
	}
	// 广播当前投票状态
	votes := room.rematchVotes
	for _, cl := range room.Clients {
		if cl != nil && cl.room == room {
			cl.Send(ServerMessage{Type: "rematch_status", RematchVotes: votes})
		}
	}
	h.mu.Unlock()
}

// Leave 离开房间/取消匹配/断线处理
// graceful=true 表示玩家主动离开(立即释放席位); false 表示连接断开(进入重连宽容期)
func (h *Hub) Leave(c *Client, graceful bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c.matching {
		for i, waiter := range h.matching {
			if waiter == c {
				h.matching = append(h.matching[:i], h.matching[i+1:]...)
				break
			}
		}
		c.matching = false
		h.unmarkOnline(c)
		c.Close()
		return
	}
	room := c.room
	if room == nil {
		h.unmarkOnline(c)
		c.Close()
		return
	}
	if graceful {
		room.removeClient(c)
		h.unmarkOnline(c)
		return
	}
	// 断线: 保留席位, 启动重连宽容期
	if room.Clients[c.playerIndex] == c {
		c.disconnected = true
		h.unmarkOnline(c)
		room.stopTurnTimer()
		if other := room.otherClient(c); other != nil && !other.disconnected {
			other.Send(ServerMessage{Type: "player_disconnected", PlayerIndex: c.playerIndex, Players: room.playersInfo(), Reason: "对方已掉线, 等待重连…"})
		}
		grace := h.disconnectGrace
		if grace <= 0 {
			grace = 45 * time.Second
		}
		room.disconnectTimer = time.AfterFunc(grace, func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			// 若已重连则跳过
			if room.Clients[c.playerIndex] != c || !c.disconnected {
				return
			}
			logger.Log.Info("重连超时, 释放席位", zap.String("room", room.Code), zap.Int("player", c.playerIndex))
			room.removeClient(c)
		})
	}
	c.Close()
}

// Close 关闭中心
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for _, room := range h.rooms {
		room.closeAll("服务器正在关闭")
	}
	for _, c := range h.matching {
		c.Send(ServerMessage{Type: "room_closed", Reason: "服务器正在关闭"})
		c.Close()
	}
	h.matching = nil
	h.rooms = make(map[string]*Room)
}

// newRoom 创建房间
func (h *Hub) newRoom(mode string) *Room {
	code := h.generateCode()
	return &Room{
		Code: code,
		Mode: mode,
		Hub:  h,
	}
}

// generateCode 生成唯一房间码
func (h *Hub) generateCode() string {
	for {
		code := randomRoomCode(h.roomCodeLen)
		if _, ok := h.rooms[code]; !ok {
			return code
		}
	}
}

// announceRoom 广播房间成员变化
func (h *Hub) announceRoom(room *Room) {
	msg := ServerMessage{Type: "player_joined", Players: room.playersInfo()}
	for _, c := range room.Clients {
		if c != nil && c.room == room {
			c.Send(msg)
		}
	}
}
