package hub

import (
	"cakeduel-backend/internal/game"
)

// eventMsgsFromFiltered 将已按公开过滤的事件转换为消息(不再做玩家过滤)
func eventMsgsFromFiltered(filtered []game.Event) []EventMsg {
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

// SpectatorPlayerMsg 观战视角的玩家公开信息
type SpectatorPlayerMsg struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Cakes     int    `json:"cakes"`
	HandCount int    `json:"handCount"`
}

// SpectatorViewMsg 观战视角(不含任何手牌信息)
type SpectatorViewMsg struct {
	Frame            int                  `json:"frame"`
	Phase            string               `json:"phase"`
	AttackerIndex    int                  `json:"attackerIndex"`
	RoundNumber      int                  `json:"roundNumber"`
	BoutWinners      []int                `json:"boutWinners"`
	GameEnded        *GameEndedMsg        `json:"gameEnded"`
	Config           GameConfigMsg        `json:"config"`
	LastAttackPassed bool                 `json:"lastAttackPassed"`
	Paused           bool                 `json:"paused"`
	Players          []SpectatorPlayerMsg `json:"players"`
	AttackingClaim   *ClaimMsg            `json:"attackingClaim"`
	BlockingClaim    *ClaimMsg            `json:"blockingClaim"`
	DeckCount        int                  `json:"deckCount"`
	DiscardCount     int                  `json:"discardCount"`
}

// buildSpectatorView 构建观战视角
func buildSpectatorView(s *game.State) *SpectatorViewMsg {
	msg := &SpectatorViewMsg{
		Frame:            s.Frame,
		Phase:            s.Phase,
		AttackerIndex:    s.AttackerIndex,
		RoundNumber:      len(s.BoutWinners) + 1,
		BoutWinners:      append([]int{}, s.BoutWinners...),
		LastAttackPassed: s.LastAttackPassed,
		DeckCount:        len(s.Deck),
		DiscardCount:     len(s.Discard),
		Config: GameConfigMsg{
			RoundsToWin:        s.Config.RoundsToWin,
			SpecialCardsToAdd:  s.Config.SpecialCardsToAdd,
			StartingHandLimit:  s.Config.StartingHandLimit,
			TurnTimeoutSeconds: s.Config.TurnTimeoutSeconds,
		},
	}
	if s.GameEnded != nil {
		msg.GameEnded = &GameEndedMsg{Winner: s.GameEnded.Winner}
	}
	for i := 0; i < 2; i++ {
		msg.Players = append(msg.Players, SpectatorPlayerMsg{
			Index:     i,
			Cakes:     s.Players[i].Cakes,
			HandCount: len(s.Players[i].Hand),
		})
	}
	if s.AttackingClaim != nil {
		msg.AttackingClaim = &ClaimMsg{Claim: s.AttackingClaim.Claim, CardCount: len(s.AttackingClaim.CardIDs)}
	}
	if s.BlockingClaim != nil {
		msg.BlockingClaim = &ClaimMsg{Claim: s.BlockingClaim.Claim, CardCount: len(s.BlockingClaim.CardIDs)}
	}
	return msg
}

// buildSpectatorZones 观战分区(只含公开牌堆实体, 不含任何手牌名称)
func buildSpectatorZones(s *game.State, events []game.Event) *ZonesMsg {
	base := len(s.BoutWinners) * len(s.CardNames)
	z := &ZonesMsg{
		RevealedPileCards: make(map[int]string),
		DeckCount:         len(s.Deck),
		DiscardCount:      len(s.Discard),
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
