package game

// AttackDef 攻击定义
type AttackDef struct {
	Damage       int       `json:"damage"`
	ExtraEffects []string  `json:"extraEffects,omitempty"`
	TriggerPick  *PickDef  `json:"triggerPick,omitempty"`
}

// PickDef 选牌效果定义
type PickDef struct {
	Type       string `json:"type"`
	Constraint string `json:"constraint"`
}

// BlockDef 防守定义
type BlockDef struct {
	BlockType          string   `json:"blockType"`
	BlockMultipleCards bool     `json:"blockMultipleCards,omitempty"`
	BlockExtraEffects  bool     `json:"blockExtraEffects,omitempty"`
	ExtraEffects       []string `json:"extraEffects,omitempty"`
}

// CardDef 卡牌定义
type CardDef struct {
	Type                 string    `json:"type"`
	Set                  string    `json:"set"`
	Notation             string    `json:"notation"`
	Attack               *AttackDef `json:"attack,omitempty"`
	Block                *BlockDef  `json:"block,omitempty"`
	Taunt                bool      `json:"taunt,omitempty"`
	TransformOnChallenge string    `json:"transformOnChallenge,omitempty"`
}

// 效果常量
const (
	EffectAttacksAgain          = "current_attacker_attacks_again"
	EffectIncreaseHandSize      = "increase_hand_size_by_one"
	EffectTakeAttackPileCard    = "take_attack_pile_card"
	EffectRevealOpponentHand    = "reveal_opponent_hand"
	EffectTransformToOppHand    = "transform_to_opponents_revealed_hand"
	EffectTransformToLastClaim  = "transform_to_last_claimed"
	EffectTransformSelfWolfySci = "transform_self_and_wolfy_to_scientist"
)

// CardDefs 卡牌定义表
var CardDefs = map[string]CardDef{
	"soldier": {Type: "physical", Set: "base", Notation: "s", Attack: &AttackDef{Damage: 1}},
	"archer":  {Type: "physical", Set: "base", Notation: "a", Attack: &AttackDef{Damage: 1}},
	"wizard":  {Type: "magical", Set: "base", Notation: "w", Attack: &AttackDef{Damage: 2}},
	"defender": {Type: "blocker", Set: "base", Notation: "d", Block: &BlockDef{BlockType: "physical"}},
	"scientist": {Type: "blocker", Set: "base", Notation: "c", Block: &BlockDef{BlockType: "magical"}},
	"wolfy":    {Type: "unclaimable", Set: "base", Notation: "y", Taunt: true},
	"assassin": {Type: "physical", Set: "special", Notation: "A", Attack: &AttackDef{Damage: 5}},
	"scout": {Type: "physical", Set: "special", Notation: "S", Attack: &AttackDef{
		Damage:       1,
		ExtraEffects: []string{EffectAttacksAgain},
	}},
	"summoner": {Type: "magical", Set: "special", Notation: "M", Attack: &AttackDef{
		TriggerPick: &PickDef{Type: "name_peek_steal_two", Constraint: "single_card_name"},
	}},
	"quartermaster": {Type: "special", Set: "special", Notation: "Q", Attack: &AttackDef{
		ExtraEffects: []string{EffectIncreaseHandSize},
	}},
	"oracle": {Type: "special", Set: "special", Notation: "O", Attack: &AttackDef{
		ExtraEffects: []string{EffectRevealOpponentHand, EffectAttacksAgain},
	}},
	"priest": {Type: "blocker", Set: "special", Notation: "P", Block: &BlockDef{
		BlockType:         "all_types",
		BlockExtraEffects: true,
		ExtraEffects:      []string{EffectTakeAttackPileCard},
	}},
	"angel": {Type: "blocker", Set: "special", Notation: "L", Block: &BlockDef{
		BlockType:          "all_types",
		BlockMultipleCards: true,
		BlockExtraEffects:  true,
		ExtraEffects:       []string{EffectAttacksAgain},
	}},
	"baacrates": {Type: "unclaimable", Set: "special", Notation: "B", TransformOnChallenge: EffectTransformSelfWolfySci},
	"agent_u":   {Type: "unclaimable", Set: "special", Notation: "U", TransformOnChallenge: EffectTransformToOppHand},
	"pierrot":   {Type: "unclaimable", Set: "special", Notation: "R", TransformOnChallenge: EffectTransformToLastClaim},
}

// BaseCardList 基础卡池(29 张, 防守卡占比较高)
var BaseCardList = []string{
	"soldier", "soldier", "soldier", "soldier", "soldier", "soldier", "soldier",
	"archer", "archer", "archer", "archer", "archer", "archer",
	"wizard", "wizard", "wizard", "wizard", "wizard",
	"defender", "defender", "defender", "defender", "defender", "defender",
	"scientist", "scientist", "scientist", "scientist",
	"wolfy",
}

// SpecialCardList 特殊卡池(11 张, 强卡稀有, 牧师 2 张补防守)
var SpecialCardList = []string{
	"assassin",
	"scout",
	"summoner",
	"quartermaster",
	"oracle",
	"priest", "priest",
	"angel",
	"baacrates",
	"agent_u",
	"pierrot",
}

// GameConfig 游戏配置
type GameConfig struct {
	RoundsToWin       int    `json:"roundsToWin"`
	SpecialCardsToAdd int    `json:"specialCardsToAdd"`
	StartingHandLimit int    `json:"startingHandLimit"`
	TurnTimeoutSeconds int   `json:"turnTimeoutSeconds"`
	// DeckConfig 自定义特殊卡数量(按卡名), 为空时使用 SpecialCardsToAdd 逻辑
	DeckConfig map[string]int `json:"deckConfig,omitempty"`
}

// SpecialCardNames 可自定义数量的特殊卡(每种数量 0-3)
var SpecialCardNames = []string{
	"assassin", "scout", "summoner", "quartermaster", "oracle",
	"priest", "angel", "baacrates", "agent_u", "pierrot",
}

// NormalizeDeckConfig 规整自定义卡组(只保留合法特殊卡, 数量限制 0-3)
func NormalizeDeckConfig(dc map[string]int) map[string]int {
	if dc == nil {
		return nil
	}
	out := make(map[string]int)
	valid := make(map[string]bool, len(SpecialCardNames))
	for _, n := range SpecialCardNames {
		valid[n] = true
	}
	for name, count := range dc {
		if !valid[name] {
			continue
		}
		if count < 0 {
			count = 0
		}
		if count > 3 {
			count = 3
		}
		out[name] = count
	}
	return out
}

// expandDeckConfig 按自定义数量展开特殊卡
func expandDeckConfig(dc map[string]int) []string {
	var out []string
	for _, name := range SpecialCardNames {
		for i := 0; i < dc[name]; i++ {
			out = append(out, name)
		}
	}
	return out
}

// StartingHandLimit 初始手牌上限
const StartingHandLimit = 4
