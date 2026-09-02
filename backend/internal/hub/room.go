package hub

import (
	"cakeduel-backend/internal/game"
	"cakeduel-backend/internal/logger"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Room 房间
type Room struct {
	Code            string
	Mode            string
	Hub             *Hub
	Clients         [2]*Client
	Spectators      []*Client
	DeckConfig      map[string]int
	Game            *game.State
	GameStarted     bool
	mu              sync.Mutex
	turnTimer       *time.Timer
	disconnectTimer *time.Timer
	rematchVotes    [2]bool
	replayStarted   time.Time
	replayFrames    []ReplayFrameMsg
	replaySent      bool
}

// paused 对局是否暂停(任一方掉线等待重连)
func (r *Room) paused() bool {
	if !r.GameStarted {
		return false
	}
	for _, c := range r.Clients {
		if c != nil && c.disconnected {
			return true
		}
	}
	return false
}

// otherClient 返回另一个客户端
func (r *Room) otherClient(c *Client) *Client {
	for _, cl := range r.Clients {
		if cl != nil && cl != c {
			return cl
		}
	}
	return nil
}

// addSpectator 添加观战(需持有 hub 锁)
func (r *Room) addSpectator(c *Client) {
	r.Spectators = append(r.Spectators, c)
	c.room = r
	c.spectating = true
}

// removeSpectator 移除观战(需持有 hub 锁)
func (r *Room) removeSpectator(c *Client) {
	for i, sp := range r.Spectators {
		if sp == c {
			r.Spectators = append(r.Spectators[:i], r.Spectators[i+1:]...)
			break
		}
	}
	c.room = nil
	c.spectating = false
	c.Close()
}

// notifySpectatorsClose 通知观战房间关闭(需持有 hub 锁)
func (r *Room) notifySpectatorsClose(reason string) {
	for _, sp := range r.Spectators {
		sp.Send(ServerMessage{Type: "room_closed", Reason: reason})
		sp.room = nil
		sp.spectating = false
		sp.Close()
	}
	r.Spectators = nil
}

// broadcastSpectatorState 向观战广播公开状态(需持有 hub/room 锁)
func (r *Room) broadcastSpectatorState(events []game.Event, reveal *RevealMsg) {
	if len(r.Spectators) == 0 || r.Game == nil {
		return
	}
	s := r.Game
	evts := eventMsgsFromFiltered(game.FilterPublicEvents(events, s.CardNames))
	view := buildSpectatorView(s)
	view.Paused = r.paused()
	zones := buildSpectatorZones(s, events)
	msg := ServerMessage{
		Type:          "spectator_state",
		SpectatorView: view,
		Zones:         zones,
		Events:        evts,
		Reveal:        reveal,
	}
	for _, sp := range r.Spectators {
		sp.Send(msg)
	}
}

// addClient 添加客户端(需持有 hub 锁)
func (r *Room) addClient(c *Client, index int) {
	r.Clients[index] = c
	c.room = r
	c.playerIndex = index
	c.matching = false
}

// removeClient 移除客户端(需持有 hub 锁)
func (r *Room) removeClient(c *Client) {
	r.stopTurnTimer()
	r.stopDisconnectTimer()
	r.rematchVotes = [2]bool{}
	index := c.playerIndex
	if r.Clients[index] == c {
		r.Clients[index] = nil
	}
	c.room = nil
	c.Close()

	// 通知剩余玩家
	var other *Client
	for _, cl := range r.Clients {
		if cl != nil && cl != c {
			other = cl
			break
		}
	}
	if other == nil {
		// 房间清空
		r.notifySpectatorsClose("对局已结束")
		delete(r.Hub.rooms, r.Code)
		return
	}

	if !r.GameStarted {
		// 对局未开始, 房间关闭
		r.notifySpectatorsClose("房间已关闭")
		other.Send(ServerMessage{Type: "room_closed", Reason: "对手已离开房间"})
		other.room = nil
		other.Close()
		delete(r.Hub.rooms, r.Code)
		return
	}

	if r.Game != nil && r.Game.GameEnded != nil {
		// 对局已自然结束, 有人离开只是关闭房间, 不再误报"对手离开你获胜"
		r.notifySpectatorsClose("对局已结束")
		other.Send(ServerMessage{Type: "room_closed", Reason: "对局结束, 对方已离开"})
		other.room = nil
		other.Close()
		delete(r.Hub.rooms, r.Code)
		return
	}

	// 对局中, 直接判定对手获胜
	r.notifySpectatorsClose("对局已结束")
	other.Send(ServerMessage{Type: "opponent_left", Reason: "对手已离开, 你获胜", LeftIndex: index})
	other.room = nil
	other.Close()
	delete(r.Hub.rooms, r.Code)
}

// closeAll 关闭房间内所有连接
func (r *Room) closeAll(reason string) {
	r.stopTurnTimer()
	r.stopDisconnectTimer()
	r.rematchVotes = [2]bool{}
	for _, c := range r.Clients {
		if c != nil {
			c.Send(ServerMessage{Type: "room_closed", Reason: reason})
			c.room = nil
			c.Close()
		}
	}
	r.Clients = [2]*Client{}
	r.notifySpectatorsClose(reason)
}

// playersInfo 玩家信息列表
func (r *Room) playersInfo() []PlayerInfo {
	out := make([]PlayerInfo, 0, 2)
	for i, c := range r.Clients {
		if c != nil {
			out = append(out, PlayerInfo{Index: i, Name: c.Name, Connected: !c.disconnected})
		} else {
			out = append(out, PlayerInfo{Index: i, Name: "", Connected: false})
		}
	}
	return out
}

// announceJoined 通知房间成员房间信息
func (r *Room) announceJoined() {
	for i, c := range r.Clients {
		if c != nil {
			c.Send(ServerMessage{
				Type:        "room_joined",
				RoomCode:    r.Code,
				Mode:        r.Mode,
				PlayerIndex: i,
				Players:     r.playersInfo(),
				DeckConfig:  r.DeckConfig,
			})
		}
	}
}

// waitForStart 等待房间开局(房主离开时清理)
func (r *Room) waitForStart() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		r.Hub.mu.Lock()
		if _, ok := r.Hub.rooms[r.Code]; !ok {
			r.Hub.mu.Unlock()
			return
		}
		if r.GameStarted {
			r.Hub.mu.Unlock()
			return
		}
		r.Hub.mu.Unlock()
	}
}

// startGame 开始新对局(需持有 hub 锁)
func (r *Room) startGame() {
	r.rematchVotes = [2]bool{}
	r.replayStarted = time.Now()
	r.replayFrames = nil
	r.replaySent = false
	seed := rand.Uint32()
	cfg := r.Hub.gameConfig
	if r.DeckConfig != nil {
		cfg.DeckConfig = r.DeckConfig
	}
	state, events, err := game.NewGame(cfg, seed)
	if err != nil {
		logger.Log.Error("创建游戏失败", zap.Error(err))
		return
	}
	r.Game = state
	r.GameStarted = true
	r.broadcastGameState(events)
}

// applyAction 应用玩家动作(需持有 room 锁)
func (r *Room) applyAction(c *Client, action *ActionMsg) {
	if action == nil {
		c.sendError("缺少动作参数")
		return
	}
	gAction := &game.Action{
		Type:        action.Type,
		HandIndices: action.HandIndices,
		Claim:       action.Claim,
		PickIndices: action.PickIndices,
	}
	var reveal *RevealMsg
	if gAction.Type == game.ActionChallenge {
		reveal = r.buildReveal()
	}
	events, err := game.Step(r.Game, c.playerIndex, gAction)
	if err != nil {
		c.sendError(err.Error())
		return
	}
	r.broadcastGameState(events, reveal)
}

// buildReveal 构造质疑翻开数据(需在动作执行前调用)
func (r *Room) buildReveal() *RevealMsg {
	s := r.Game
	pile := "attack_pile"
	var claim *game.Claim
	if s.BlockingClaim != nil {
		pile = "block_pile"
		claim = s.BlockingClaim
	} else if s.AttackingClaim != nil {
		claim = s.AttackingClaim
	}
	if claim == nil {
		return nil
	}
	base := len(s.BoutWinners) * len(s.CardNames)
	cards := make([]CardEntityMsg, 0, len(claim.CardIDs))
	for _, id := range claim.CardIDs {
		name := s.CardNames[id]
		cards = append(cards, CardEntityMsg{EntityID: base + id, Name: name})
	}
	return &RevealMsg{Pile: pile, Cards: cards}
}

// broadcastGameState 向双方广播最新状态
func (r *Room) broadcastGameState(events []game.Event, reveal ...*RevealMsg) {
	s := r.Game
	if s == nil {
		return
	}
	var revealMsg *RevealMsg
	if len(reveal) > 0 {
		revealMsg = reveal[0]
	}
	for i := 0; i < 2; i++ {
		c := r.Clients[i]
		if c == nil {
			continue
		}
		view := buildViewMsg(s, i)
		zones := buildZonesMsg(s, i, events)
		legal := buildLegalMsg(s, i)
		evts := buildEventsMsg(events, i, s.CardNames)
		yourTurn := phasingIndex(s) == i
		msg := ServerMessage{
			Type:     "game_state",
			View:     view,
			Zones:    zones,
			Legal:    legal,
			Events:   evts,
			Reveal:   revealMsg,
			YourTurn: yourTurn,
			GameOver: s.GameEnded != nil,
			Paused:   r.paused(),
		}
		c.Send(msg)
	}
	r.recordReplayFrame(events, revealMsg)
	r.broadcastSpectatorState(events, revealMsg)
	if s.GameEnded != nil {
		r.sendReplayData()
	}
	r.scheduleTurnTimer()
}

// scheduleTurnTimer 为当前行动玩家启动倒计时, 超时自动过牌(防挂机)
func (r *Room) scheduleTurnTimer() {
	r.stopTurnTimer()
	s := r.Game
	if s == nil || s.GameEnded != nil || r.paused() {
		return
	}
	player := phasingIndex(s)
	if player < 0 {
		return
	}
	timeout := r.Hub.gameConfig.TurnTimeoutSeconds
	if timeout <= 0 {
		return
	}
	frame := s.Frame
	r.turnTimer = time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.Game == nil || r.Game.GameEnded != nil || r.Game.Frame != frame {
			return
		}
		playerNow := phasingIndex(r.Game)
		if playerNow != player {
			return
		}
		legal := game.LegalActions(r.Game, player)
		var action *game.Action
		for _, l := range legal {
			if l.Type == game.ActionPass {
				action = &game.Action{Type: game.ActionPass}
				break
			}
		}
		if action == nil {
			for _, l := range legal {
				if l.Type == game.ActionPick && len(l.PickFrom) > 0 {
					action = &game.Action{Type: game.ActionPick, PickIndices: []int{0}}
					break
				}
			}
		}
		if action == nil {
			return
		}
		events, err := game.Step(r.Game, player, action)
		if err != nil {
			return
		}
		r.broadcastGameState(events)
	})
}

// stopDisconnectTimer 停止重连宽容计时器
func (r *Room) stopDisconnectTimer() {
	if r.disconnectTimer != nil {
		r.disconnectTimer.Stop()
		r.disconnectTimer = nil
	}
}

// stopTurnTimer 停止倒计时
func (r *Room) stopTurnTimer() {
	if r.turnTimer != nil {
		r.turnTimer.Stop()
		r.turnTimer = nil
	}
}

// phasingIndex 当前行动玩家
func phasingIndex(s *game.State) int {
	switch s.Phase {
	case game.PhaseAttack:
		return s.AttackerIndex
	case game.PhaseBlock:
		return 1 - s.AttackerIndex
	case game.PhaseReview:
		return s.AttackerIndex
	case game.PhasePick:
		if len(s.PickPhaseEffects) > 0 {
			return s.PickPhaseEffects[0].Player
		}
	}
	return -1
}

// buildViewMsg 构建玩家视角消息
func buildViewMsg(s *game.State, player int) *PlayerViewMsg {
	view := game.GetView(s, player)
	msg := &PlayerViewMsg{
		Frame: view.Frame,
		Me: MeMsg{
			Index:     view.Me.Index,
			Cakes:     view.Me.Cakes,
			Hand:      view.Me.Hand,
			HandLimit: view.Me.HandLimit,
		},
		Opponent: OpponentMsg{
			Index:     view.Opponent.Index,
			Cakes:     view.Opponent.Cakes,
			HandCount: view.Opponent.HandCount,
		},
		DeckCount:        view.DeckCount,
		DiscardCount:     view.DiscardCount,
		AttackingClaim:   toClaimMsg(view.AttackingClaim),
		BlockingClaim:    toClaimMsg(view.BlockingClaim),
		Phase:            view.Phase,
		AttackerIndex:    view.AttackerIndex,
		BoutWinners:      view.BoutWinners,
		LastAttackPassed: view.LastAttackPassed,
		Config: GameConfigMsg{
			RoundsToWin:        view.Config.RoundsToWin,
			SpecialCardsToAdd:  view.Config.SpecialCardsToAdd,
			StartingHandLimit:  view.Config.StartingHandLimit,
			TurnTimeoutSeconds: view.Config.TurnTimeoutSeconds,
		},
		RoundNumber: len(view.BoutWinners) + 1,
	}
	if view.GameEnded != nil {
		msg.GameEnded = &GameEndedMsg{Winner: view.GameEnded.Winner}
	}
	return msg
}

func toClaimMsg(c *game.ClaimView) *ClaimMsg {
	if c == nil {
		return nil
	}
	return &ClaimMsg{Claim: c.Claim, CardCount: c.CardCount}
}

// buildZonesMsg 构建分区卡牌消息
func buildZonesMsg(s *game.State, player int, events []game.Event) *ZonesMsg {
	base := len(s.BoutWinners) * len(s.CardNames)
	z := &ZonesMsg{
		RevealedPileCards: make(map[int]string),
		DeckCount:         len(s.Deck),
		DiscardCount:      len(s.Discard),
	}
	for _, id := range s.Players[player].Hand {
		z.PlayerHand = append(z.PlayerHand, CardEntityMsg{EntityID: base + id, Name: s.CardNames[id]})
	}
	for _, id := range s.Players[1-player].Hand {
		z.OpponentHand = append(z.OpponentHand, CardEntityMsg{EntityID: base + id})
	}
	if s.AttackingClaim != nil {
		for _, id := range s.AttackingClaim.CardIDs {
			z.AttackPile = append(z.AttackPile, CardEntityMsg{EntityID: base + id})
		}
	}
	if s.BlockingClaim != nil {
		for _, id := range s.BlockingClaim.CardIDs {
			z.BlockPile = append(z.BlockPile, CardEntityMsg{EntityID: base + id})
		}
	}
	top := len(s.Deck)
	if top > 4 {
		top = 4
	}
	for i := 0; i < top; i++ {
		z.DeckTop = append(z.DeckTop, CardEntityMsg{EntityID: base + s.Deck[i]})
	}
	// 质疑翻开信息
	for _, evt := range events {
		if evt.Type == game.EventChallengeMade {
			for _, id := range evt.RevealedIDs() {
				if id >= 0 && id < len(s.CardNames) {
					z.RevealedPileCards[base+id] = s.CardNames[id]
				}
			}
		}
	}
	return z
}

// buildLegalMsg 构建合法动作消息
func buildLegalMsg(s *game.State, player int) []LegalMsg {
	specs := game.LegalActions(s, player)
	out := make([]LegalMsg, 0, len(specs))
	for _, spec := range specs {
		out = append(out, LegalMsg{
			Type:                 spec.Type,
			ClaimFrom:            spec.ClaimFrom,
			AvailableHandIndices: spec.AvailableHandIndices,
			PickType:             spec.PickType,
			PickFrom:             spec.PickFrom,
			MinPicks:             spec.MinPicks,
			MaxPicks:             spec.MaxPicks,
		})
	}
	return out
}

// buildEventsMsg 构建事件消息
func buildEventsMsg(events []game.Event, player int, cardNames []string) []EventMsg {
	filtered := game.FilterEvents(events, player, cardNames)
	out := make([]EventMsg, 0, len(filtered))
	for _, evt := range filtered {
		msg := EventMsg{
			ID:            evt.ID,
			Type:          evt.Type,
			Player:        evt.Player,
			Phase:         evt.Phase,
			Pile:          evt.Pile,
			Claim:         evt.Claim,
			CardNames:     evt.CardNames,
			RevealedCards: nil,
			Challenger:    evt.Challenger,
			ClaimedCard:   evt.ClaimedCard,
			Success:       evt.Success,
			From:          evt.From,
			To:            evt.To,
			Amount:        evt.Amount,
			CakesAfter:    evt.CakesAfter,
			Winner:        evt.Winner,
			AttackerIndex: evt.AttackerIndex,
			BoutNumber:    evt.BoutNumber,
			PickType:      evt.PickType,
			Picks:         evt.Picks,
			Zone:          evt.Zone,
		}
		for _, rc := range evt.RevealedCards {
			msg.RevealedCards = append(msg.RevealedCards, RevealedMsg{CardName: rc.CardName, TransformedTo: rc.TransformedTo})
		}
		out = append(out, msg)
	}
	return out
}
