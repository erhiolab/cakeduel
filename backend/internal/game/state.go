package game

// Player 玩家状态
type Player struct {
	Hand              []int    `json:"hand"`
	HandLimit         int      `json:"handLimit"`
	ClaimBlacklist    []string `json:"claimBlacklist"`
	Cakes             int      `json:"cakes"`
	LastAttackingClaim string   `json:"lastAttackingClaim"`
}

// Claim 声明(打出的牌)
type Claim struct {
	Claim   string `json:"claim"`
	CardIDs []int  `json:"cardIds"`
}

// PickEffect 选牌阶段效果
type PickEffect struct {
	Player   int
	Effect   PickDef
	PickFrom []string
}

// GameEnded 游戏结束信息
type GameEnded struct {
	Winner int `json:"winner"`
}

// State 游戏状态
type State struct {
	Frame                      int
	LastEventID                int
	Phase                      string // attack | block | review | pick
	LastAttackPassed           bool
	GameEnded                  *GameEnded
	Deck                       []int
	Discard                    []int
	AttackingClaim             *Claim
	BlockingClaim              *Claim
	Players                    [2]Player
	BoutWinners                []int
	AttackerIndex              int
	NextAttackerIndexOverride  []int
	PickPhaseEffects           []PickEffect
	RNG                        *RNG
	CardNames                  []string
	Config                     GameConfig
}

// Phases
const (
	PhaseAttack  = "attack"
	PhaseBlock   = "block"
	PhaseReview  = "review"
	PhasePick    = "pick"
)

// Action 客户端动作
type Action struct {
	Type        string `json:"type"`
	HandIndices []int  `json:"handIndices,omitempty"`
	Claim       string `json:"claim,omitempty"`
	PickIndices []int  `json:"pickIndices,omitempty"`
}

// Action types
const (
	ActionClaim     = "claim"
	ActionPass      = "pass"
	ActionChallenge = "challenge"
	ActionPick      = "pick"
	ActionConcede   = "concede"
)

// opponent 返回对方玩家索引
func opponent(index int) int {
	if index == 0 {
		return 1
	}
	return 0
}

// phasingPlayer 当前阶段行动玩家
func phasingPlayer(s *State) int {
	switch s.Phase {
	case PhaseAttack:
		return s.AttackerIndex
	case PhaseBlock:
		return opponent(s.AttackerIndex)
	case PhaseReview:
		return s.AttackerIndex
	case PhasePick:
		if len(s.PickPhaseEffects) > 0 {
			return s.PickPhaseEffects[0].Player
		}
	}
	return -1
}

// rangeInt 返回 [0,n)
func rangeInt(n int) []int {
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, i)
	}
	return out
}

// clamp 限制范围
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// last 取最后一个元素
func last(items []int) int {
	if len(items) == 0 {
		return -1
	}
	return items[len(items)-1]
}

// containsString 判断字符串切片是否包含
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// containsInt 判断整数切片是否包含
func containsInt(items []int, target int) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// unique 字符串去重
func unique(items []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}
