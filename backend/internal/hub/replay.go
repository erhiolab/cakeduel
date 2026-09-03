package hub

import (
	"cakeduel-backend/internal/game"
	"cakeduel-backend/internal/utils"
	"context"
	"errors"
	"time"
)

// 分享回放的 KV key 前缀与有效期
const (
	sharedReplayPrefix = "cakeduel:shared:replay:"
	sharedReplayTTL    = 24 * time.Hour
	maxSharedReplayLen = 4 * 1024 * 1024
)

// ShareReplay 保存一份可分享的回放, 返回访问 ID(Redis 存 1 天)
func (h *Hub) ShareReplay(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("回放数据为空")
	}
	if len(raw) > maxSharedReplayLen {
		return "", errors.New("回放数据过大")
	}
	id := utils.RandomString(10)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := h.kvSet(ctx, sharedReplayPrefix+id, string(raw), sharedReplayTTL); err != nil {
		return "", err
	}
	return id, nil
}

// FetchSharedReplay 按 ID 读取分享的回放
func (h *Hub) FetchSharedReplay(id string) (string, bool) {
	if id == "" {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return h.kvGet(ctx, sharedReplayPrefix+id)
}

// MAX_REPLAY_FRAMES 单局回放最多记录的帧数(防止异常局撑爆内存)
const MAX_REPLAY_FRAMES = 600

// ReplayFrameMsg 回放的一帧(与观战视角同构, 不含任何私有手牌信息)
type ReplayFrameMsg struct {
	View        *SpectatorViewMsg `json:"view"`
	Zones       *ZonesMsg         `json:"zones"`
	Events      []EventMsg        `json:"events"`
	Reveal      *RevealMsg        `json:"reveal,omitempty"`
	PlayerHands [2][]string       `json:"playerHands"`
	AttackPile  []string          `json:"attackPile"`
	BlockPile   []string          `json:"blockPile"`
}

// ReplayDataMsg 对局回放数据
type ReplayDataMsg struct {
	RoomCode    string           `json:"roomCode"`
	Mode        string           `json:"mode"`
	PlayerNames [2]string        `json:"playerNames"`
	StartedAt   int64            `json:"startedAt"`
	DurationMs  int64            `json:"durationMs"`
	Winner      int              `json:"winner"`
	DeckConfig  map[string]int   `json:"deckConfig"`
	RoundsToWin int              `json:"roundsToWin"`
	Frames      []ReplayFrameMsg `json:"frames"`
	Chats       []ChatMsgData    `json:"chats"`
}

// recordReplayFrame 在每次状态广播时记录一帧公开状态(需持有 room 锁)
func (r *Room) recordReplayFrame(events []game.Event, reveal *RevealMsg) {
	s := r.Game
	if s == nil || r.replayStarted.IsZero() || len(r.replayFrames) >= MAX_REPLAY_FRAMES {
		return
	}
	// 重连补发状态等无新事件的广播不产生新帧
	if len(events) == 0 && reveal == nil {
		return
	}
	frame := ReplayFrameMsg{
		View:   buildSpectatorView(s),
		Zones:  buildSpectatorZones(s, events),
		Events: eventMsgsFromFiltered(game.FilterPublicEvents(events, s.CardNames)),
		Reveal: reveal,
	}
	// 回放记录双方手牌与出牌堆的真实牌面(观战看不到, 回放可见)
	for p := 0; p < 2; p++ {
		for _, id := range s.Players[p].Hand {
			if id >= 0 && id < len(s.CardNames) {
				frame.PlayerHands[p] = append(frame.PlayerHands[p], s.CardNames[id])
			}
		}
	}
	if s.AttackingClaim != nil {
		for _, id := range s.AttackingClaim.CardIDs {
			if id >= 0 && id < len(s.CardNames) {
				frame.AttackPile = append(frame.AttackPile, s.CardNames[id])
			}
		}
	}
	if s.BlockingClaim != nil {
		for _, id := range s.BlockingClaim.CardIDs {
			if id >= 0 && id < len(s.CardNames) {
				frame.BlockPile = append(frame.BlockPile, s.CardNames[id])
			}
		}
	}
	r.replayFrames = append(r.replayFrames, frame)
}

// buildReplayData 汇总当前回放数据(需持有 room 锁)
func (r *Room) buildReplayData() *ReplayDataMsg {
	s := r.Game
	if s == nil || s.GameEnded == nil || r.replayStarted.IsZero() {
		return nil
	}
	data := &ReplayDataMsg{
		RoomCode:    r.Code,
		Mode:        r.Mode,
		StartedAt:   r.replayStarted.UnixMilli(),
		DurationMs:  time.Since(r.replayStarted).Milliseconds(),
		Winner:      s.GameEnded.Winner,
		DeckConfig:  map[string]int{},
		RoundsToWin: s.Config.RoundsToWin,
		Frames:      append([]ReplayFrameMsg{}, r.replayFrames...),
		Chats:       r.chatSnapshot(),
	}
	for i, c := range r.Clients {
		if c != nil {
			data.PlayerNames[i] = c.Name
		}
	}
	if r.DeckConfig != nil {
		for k, v := range r.DeckConfig {
			data.DeckConfig[k] = v
		}
	}
	return data
}

// sendReplayData 对局结束后把回放数据推送给双方玩家
func (r *Room) sendReplayData() {
	if r.replaySent {
		return
	}
	data := r.buildReplayData()
	if data == nil {
		return
	}
	r.replaySent = true
	msg := ServerMessage{Type: "replay_data", Replay: data}
	for _, c := range r.Clients {
		if c != nil {
			c.Send(msg)
		}
	}
}
