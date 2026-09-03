package hub

import (
	"cakeduel-backend/internal/game"
	"cakeduel-backend/internal/logger"
	"cakeduel-backend/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// 管理员与房间记录相关的 Redis key / TTL
const (
	adminPasswordKey = "cakeduel:admin:password"
	adminAuthPrefix  = "cakeduel:admin:auth:"
	roomRecordPrefix = "cakeduel:admin:room:"

	adminChallengeTTL = time.Minute
	adminAuthTTL      = 30 * time.Minute
	// 房间实时记录只服务“进行中”对局: 结束即删, 异常残留也应在 10 分钟内过期
	roomRecordTTL = 10 * time.Minute
)

// AdminClientView 在线客户端视图
type AdminClientView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Token    string `json:"token"`
	RoomCode string `json:"roomCode,omitempty"`
	Role     string `json:"role"`
	Matching bool   `json:"matching"`
}

// AdminPlayerView 房间内玩家视图(管理员可见双方手牌)
type AdminPlayerView struct {
	Index     int      `json:"index"`
	Name      string   `json:"name"`
	Connected bool     `json:"connected"`
	Cakes     int      `json:"cakes"`
	HandCount int      `json:"handCount"`
	Hand      []string `json:"hand"`
}

// AdminRoomView 房间视图(含双方手牌/聊天/观战人数)
type AdminRoomView struct {
	Code           string            `json:"code"`
	Mode           string            `json:"mode"`
	Status         string            `json:"status"`
	Phase          string            `json:"phase"`
	Paused         bool              `json:"paused"`
	Players        []AdminPlayerView `json:"players"`
	SpectatorNames []string          `json:"spectatorNames"`
	SpectatorCount int               `json:"spectatorCount"`
	AttackerIndex  int               `json:"attackerIndex"`
	RoundNumber    int               `json:"roundNumber"`
	BoutWinners    []int             `json:"boutWinners"`
	Claims         []string          `json:"claims"`
	GameOver       bool              `json:"gameOver"`
	Winner         int               `json:"winner"`
	DeckConfig     map[string]int    `json:"deckConfig"`
	ChatHistory    []ChatMsgData     `json:"chatHistory"`
}

// AdminCreateChallenge 创建一分钟有效的管理员访问密码并写入 KV(Redis)
// 密码不回显给页面, 由管理员自行从 Redis 读取
func (h *Hub) AdminCreateChallenge() error {
	password := utils.RandomString(6)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.kvSet(ctx, adminPasswordKey, password, adminChallengeTTL); err != nil {
		return err
	}
	return nil
}

// AdminVerify 校验密码: 读取 Redis 中一次性密码比对, 成功后删除并发放令牌
func (h *Hub) AdminVerify(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", errors.New("请输入密码")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	stored, ok := h.kvGet(ctx, adminPasswordKey)
	if !ok {
		return "", errors.New("未生成密码或已过期")
	}
	if stored != password {
		return "", errors.New("密码错误")
	}
	h.kvDel(ctx, adminPasswordKey)
	token := utils.RandomString(24)
	if err := h.kvSet(ctx, adminAuthPrefix+token, "1", adminAuthTTL); err != nil {
		return "", err
	}
	return token, nil
}

// AdminAuthorize 校验管理员令牌
func (h *Hub) AdminAuthorize(token string) bool {
	if strings.TrimSpace(token) == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	val, ok := h.kvGet(ctx, adminAuthPrefix+token)
	return ok && val == "1"
}

// AdminClients 当前在线客户端列表
func (h *Hub) AdminClients() []AdminClientView {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]AdminClientView, 0, len(h.clients))
	for c := range h.clients {
		role := "online"
		code := ""
		switch {
		case c.waiting:
			role = "waiting_spectator"
		case c.spectating:
			role = "spectator"
		case c.matching:
			role = "matching"
		case c.room != nil:
			role = "player"
		}
		if c.room != nil {
			code = c.room.Code
		}
		tail := c.Token
		if len(tail) > 6 {
			tail = tail[len(tail)-6:]
		}
		out = append(out, AdminClientView{
			ID:       c.ID,
			Name:     c.Name,
			Token:    tail,
			RoomCode: code,
			Role:     role,
			Matching: c.matching,
		})
	}
	return out
}

// AdminRooms 当前房间列表(含双方手牌/聊天)
func (h *Hub) AdminRooms() []AdminRoomView {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]AdminRoomView, 0, len(h.rooms))
	for _, r := range h.rooms {
		out = append(out, h.snapshotRoomView(r))
	}
	return out
}

// snapshotRoomView 房间完整快照(调用方需持有 hub 锁)
func (h *Hub) snapshotRoomView(r *Room) AdminRoomView {
	r.mu.Lock()
	defer r.mu.Unlock()
	view := AdminRoomView{
		Code:           r.Code,
		Mode:           r.Mode,
		Status:         "waiting",
		DeckConfig:     map[string]int{},
		ChatHistory:    r.chatSnapshot(),
		SpectatorNames: make([]string, 0, len(r.Spectators)),
	}
	for i, c := range r.Clients {
		pv := AdminPlayerView{
			Index:     i,
			Name:      "",
			Connected: false,
			Hand:      []string{},
		}
		if c != nil {
			pv.Name = c.Name
			pv.Connected = !c.disconnected
		}
		view.Players = append(view.Players, pv)
	}
	for _, sp := range r.Spectators {
		if sp != nil {
			view.SpectatorNames = append(view.SpectatorNames, sp.Name)
		}
	}
	view.SpectatorCount = len(view.SpectatorNames)
	if r.DeckConfig != nil {
		for k, v := range r.DeckConfig {
			view.DeckConfig[k] = v
		}
	}
	if !r.GameStarted {
		view.Status = "waiting"
		return view
	}
	s := r.Game
	if s == nil {
		view.Status = "waiting"
		return view
	}
	view.Status = "playing"
	view.Phase = s.Phase
	view.Paused = r.paused()
	view.AttackerIndex = s.AttackerIndex
	view.RoundNumber = len(s.BoutWinners) + 1
	view.BoutWinners = append([]int{}, s.BoutWinners...)
	if s.GameEnded != nil {
		view.Status = "finished"
		view.GameOver = true
		view.Winner = s.GameEnded.Winner
	}
	for i, p := range s.Players {
		view.Players[i].Cakes = p.Cakes
		view.Players[i].HandCount = len(p.Hand)
		for _, id := range p.Hand {
			if id >= 0 && id < len(s.CardNames) {
				view.Players[i].Hand = append(view.Players[i].Hand, s.CardNames[id])
			}
		}
	}
	if s.AttackingClaim != nil {
		view.Claims = append(view.Claims, claimViewText(s, s.AttackingClaim))
	}
	if s.BlockingClaim != nil {
		view.Claims = append(view.Claims, claimViewText(s, s.BlockingClaim))
	}
	return view
}

// claimViewText 声明文本
func claimViewText(s *game.State, claim *game.Claim) string {
	return claim.Claim + "×" + strconv.Itoa(len(claim.CardIDs))
}

// AdminDismissRoom 强制解散房间
func (h *Hub) AdminDismissRoom(code string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.rooms[code]
	if !ok {
		return errors.New("房间不存在")
	}
	r.closeAll("房间已被管理员解散")
	delete(h.rooms, code)
	h.removeRoomRecord(code)
	logger.Log.Info("管理员解散房间", zap.String("room", code))
	return nil
}

// persistRoomRecord 把房间完整信息写入 KV(Redis), 异步执行避免阻塞对局
func (h *Hub) persistRoomRecord(r *Room) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Log.Warn("房间记录写入异常", zap.Any("panic", rec))
			}
		}()
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.rooms[r.Code] != r {
			return
		}
		view := h.snapshotRoomView(r)
		data, err := json.Marshal(view)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = h.kvSet(ctx, roomRecordPrefix+r.Code, string(data), roomRecordTTL)
	}()
}

// removeRoomRecord 删除房间 KV 记录
func (h *Hub) removeRoomRecord(code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	h.kvDel(ctx, roomRecordPrefix+code)
}
