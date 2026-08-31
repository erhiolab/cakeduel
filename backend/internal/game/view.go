package game

// MeView 己方视角
type MeView struct {
	Index     int      `json:"index"`
	Cakes     int      `json:"cakes"`
	Hand      []string `json:"hand"`
	HandLimit int      `json:"handLimit"`
}

// OpponentView 对方视角
type OpponentView struct {
	Index     int    `json:"index"`
	Cakes     int    `json:"cakes"`
	HandCount int    `json:"handCount"`
}

// ClaimView 声明视角
type ClaimView struct {
	Claim     string `json:"claim"`
	CardCount int    `json:"cardCount"`
}

// PlayerView 玩家视角
type PlayerView struct {
	Frame             int         `json:"frame"`
	Me                MeView      `json:"me"`
	Opponent          OpponentView `json:"opponent"`
	DeckCount         int         `json:"deckCount"`
	DiscardCount      int         `json:"discardCount"`
	AttackingClaim    *ClaimView  `json:"attackingClaim"`
	BlockingClaim     *ClaimView  `json:"blockingClaim"`
	Phase             string      `json:"phase"`
	AttackerIndex     int         `json:"attackerIndex"`
	BoutWinners       []int       `json:"boutWinners"`
	GameEnded         *GameEnded  `json:"gameEnded"`
	Config            GameConfig  `json:"config"`
	LastAttackPassed  bool        `json:"lastAttackPassed"`
}

// GetView 获取玩家视角
func GetView(s *State, playerIndex int) *PlayerView {
	opp := opponent(playerIndex)
	view := &PlayerView{
		Frame:         s.Frame,
		Me: MeView{
			Index:     playerIndex,
			Cakes:     s.Players[playerIndex].Cakes,
			Hand:      namesOf(s, s.Players[playerIndex].Hand),
			HandLimit: s.Players[playerIndex].HandLimit,
		},
		Opponent: OpponentView{
			Index:     opp,
			Cakes:     s.Players[opp].Cakes,
			HandCount: len(s.Players[opp].Hand),
		},
		DeckCount:        len(s.Deck),
		DiscardCount:     len(s.Discard),
		Phase:            s.Phase,
		AttackerIndex:    s.AttackerIndex,
		BoutWinners:      append([]int{}, s.BoutWinners...),
		Config:           s.Config,
		LastAttackPassed: s.LastAttackPassed,
	}
	if s.GameEnded != nil {
		ended := *s.GameEnded
		view.GameEnded = &ended
	}
	if s.AttackingClaim != nil {
		view.AttackingClaim = &ClaimView{
			Claim:     s.AttackingClaim.Claim,
			CardCount: len(s.AttackingClaim.CardIDs),
		}
	}
	if s.BlockingClaim != nil {
		view.BlockingClaim = &ClaimView{
			Claim:     s.BlockingClaim.Claim,
			CardCount: len(s.BlockingClaim.CardIDs),
		}
	}
	return view
}

// namesOf 卡牌实例ID转牌名
func namesOf(s *State, ids []int) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id >= 0 && id < len(s.CardNames) {
			out = append(out, s.CardNames[id])
		} else {
			out = append(out, "unknown")
		}
	}
	return out
}
