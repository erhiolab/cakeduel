package hub

import (
	"cakeduel-backend/internal/logger"
	"cakeduel-backend/internal/utils"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client WebSocket 客户端
type Client struct {
	ID           string
	Token        string
	Name         string
	conn         *websocket.Conn
	send         chan []byte
	hub          *Hub
	room         *Room
	playerIndex  int
	matching     bool
	disconnected bool
	closeOnce    sync.Once
}

// newClient 创建客户端
func newClient(h *Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:   utils.RandomString(12),
		conn: conn,
		send: make(chan []byte, 64),
		hub:  h,
	}
}

// Send 发送消息(非阻塞)
func (c *Client) Send(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		logger.Log.Warn("消息序列化失败", zap.Error(err))
		return
	}
	select {
	case c.send <- data:
	default:
		logger.Log.Warn("客户端发送队列已满, 丢弃消息", zap.String("client", c.ID))
	}
}

// writeLoop 写循环
func (c *Client) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer c.conn.Close()
	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readLoop 读循环
func (c *Client) readLoop() {
	defer func() {
		c.hub.Leave(c, false)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	})
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.Send(ServerMessage{Type: "error", Message: "消息格式错误"})
			continue
		}
		c.hub.HandleMessage(c, &msg)
	}
}

// ServeWS 处理一个 WebSocket 连接(阻塞直到断开)
func (h *Hub) ServeWS(conn *websocket.Conn, token string) {
	c := newClient(h, conn)
	if token == "" {
		token = utils.RandomString(24)
	}
	c.Token = token
	// 尝试恢复断线席位; 新连接则直接注册
	h.registerClient(c)
	go c.writeLoop()
	c.readLoop()
}

// Close 关闭连接
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
	})
}

// sendError 发送错误
func (c *Client) sendError(message string) {
	c.Send(ServerMessage{Type: "error", Message: message})
}

// ServerMessage 服务端消息
type ServerMessage struct {
	Type        string         `json:"type"`
	RoomCode    string         `json:"roomCode,omitempty"`
	PlayerIndex int            `json:"playerIndex"`
	From        int            `json:"from,omitempty"`
	Name        string         `json:"name,omitempty"`
	Text        string         `json:"text,omitempty"`
	Players     []PlayerInfo   `json:"players,omitempty"`
	LeftIndex   int            `json:"leftIndex,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	Seed        uint32         `json:"seed,omitempty"`
	View        *PlayerViewMsg `json:"view,omitempty"`
	Zones       *ZonesMsg      `json:"zones,omitempty"`
	Legal       []LegalMsg     `json:"legal"`
	Events      []EventMsg     `json:"events"`
	Reveal      *RevealMsg     `json:"reveal,omitempty"`
	YourTurn    bool           `json:"yourTurn,omitempty"`
	GameOver    bool           `json:"gameOver,omitempty"`
	Paused      bool           `json:"paused,omitempty"`
	RematchVotes [2]bool       `json:"rematchVotes,omitempty"`
	Message     string         `json:"message,omitempty"`
	Connected   []bool         `json:"connected,omitempty"`
}

// ClientMessage 客户端消息
type ClientMessage struct {
	Type   string     `json:"type"`
	Name   string     `json:"name,omitempty"`
	Mode   string     `json:"mode,omitempty"`
	Code   string     `json:"code,omitempty"`
	Action *ActionMsg `json:"action,omitempty"`
	Text   string     `json:"text,omitempty"`
}

// ActionMsg 动作消息
type ActionMsg struct {
	Type        string   `json:"type"`
	HandIndices []int    `json:"handIndices,omitempty"`
	Claim       string   `json:"claim,omitempty"`
	PickIndices []int    `json:"pickIndices,omitempty"`
}

// PlayerInfo 玩家信息
type PlayerInfo struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
}

// PlayerViewMsg 玩家视角
type PlayerViewMsg struct {
	Frame            int          `json:"frame"`
	Me               MeMsg        `json:"me"`
	Opponent         OpponentMsg  `json:"opponent"`
	DeckCount        int          `json:"deckCount"`
	DiscardCount     int          `json:"discardCount"`
	AttackingClaim   *ClaimMsg    `json:"attackingClaim"`
	BlockingClaim    *ClaimMsg    `json:"blockingClaim"`
	Phase            string       `json:"phase"`
	AttackerIndex    int          `json:"attackerIndex"`
	BoutWinners      []int        `json:"boutWinners"`
	GameEnded        *GameEndedMsg `json:"gameEnded"`
	Config           GameConfigMsg `json:"config"`
	LastAttackPassed bool         `json:"lastAttackPassed"`
	RoundNumber      int          `json:"roundNumber"`
}

// MeMsg 己方信息
type MeMsg struct {
	Index     int      `json:"index"`
	Cakes     int      `json:"cakes"`
	Hand      []string `json:"hand"`
	HandLimit int      `json:"handLimit"`
}

// OpponentMsg 对方信息
type OpponentMsg struct {
	Index     int    `json:"index"`
	Cakes     int    `json:"cakes"`
	HandCount int    `json:"handCount"`
}

// ClaimMsg 声明信息
type ClaimMsg struct {
	Claim     string `json:"claim"`
	CardCount int    `json:"cardCount"`
}

// GameEndedMsg 游戏结束信息
type GameEndedMsg struct {
	Winner int `json:"winner"`
}

// GameConfigMsg 游戏配置
type GameConfigMsg struct {
	RoundsToWin        int `json:"roundsToWin"`
	SpecialCardsToAdd  int `json:"specialCardsToAdd"`
	StartingHandLimit  int `json:"startingHandLimit"`
	TurnTimeoutSeconds int `json:"turnTimeoutSeconds"`
}

// ZonesMsg 分区卡牌
type ZonesMsg struct {
	PlayerHand       []CardEntityMsg   `json:"playerHand"`
	OpponentHand     []CardEntityMsg   `json:"opponentHand"`
	AttackPile       []CardEntityMsg   `json:"attackPile"`
	BlockPile        []CardEntityMsg   `json:"blockPile"`
	DeckTop          []CardEntityMsg   `json:"deckTop"`
	RevealedPileCards map[int]string   `json:"revealedPileCards"`
	DeckCount        int               `json:"deckCount"`
	DiscardCount     int               `json:"discardCount"`
}

// CardEntityMsg 卡牌实体
type CardEntityMsg struct {
	EntityID int    `json:"entityId"`
	Name     string `json:"name,omitempty"`
}

// LegalMsg 合法动作
type LegalMsg struct {
	Type                 string   `json:"type"`
	ClaimFrom            []string `json:"claimFrom,omitempty"`
	AvailableHandIndices []int    `json:"availableHandIndices,omitempty"`
	PickType             string   `json:"pickType,omitempty"`
	PickFrom             []string `json:"pickFrom,omitempty"`
	MinPicks             int      `json:"minPicks,omitempty"`
	MaxPicks             int      `json:"maxPicks,omitempty"`
}

// EventMsg 事件
type EventMsg struct {
	ID            int             `json:"id"`
	Type          string          `json:"type"`
	Player        *int            `json:"player,omitempty"`
	Phase         string          `json:"phase,omitempty"`
	Pile          string          `json:"pile,omitempty"`
	Claim         string          `json:"claim,omitempty"`
	CardNames     []string        `json:"cardNames,omitempty"`
	RevealedCards []RevealedMsg   `json:"revealedCards,omitempty"`
	Challenger    int             `json:"challenger,omitempty"`
	ClaimedCard   string          `json:"claimedCard,omitempty"`
	Success       bool            `json:"success"`
	From          int             `json:"from,omitempty"`
	To            int             `json:"to,omitempty"`
	Amount        int             `json:"amount,omitempty"`
	CakesAfter    [2]int          `json:"cakesAfter,omitempty"`
	Winner        int             `json:"winner,omitempty"`
	AttackerIndex int             `json:"attackerIndex,omitempty"`
	BoutNumber    int             `json:"boutNumber,omitempty"`
	PickType      string          `json:"pickType,omitempty"`
	Picks         []string        `json:"picks,omitempty"`
	Zone          string          `json:"zone,omitempty"`
}

// RevealedMsg 翻开牌信息
type RevealedMsg struct {
	CardName      string `json:"cardName"`
	TransformedTo string `json:"transformedTo,omitempty"`
}

// RevealMsg 质疑翻开动画数据
type RevealMsg struct {
	Pile  string           `json:"pile"`
	Cards []CardEntityMsg  `json:"cards"`
}
