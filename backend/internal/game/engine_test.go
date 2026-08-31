package game

import (
	"testing"
)

func testConfig() GameConfig {
	return GameConfig{
		RoundsToWin:       3,
		SpecialCardsToAdd: 11,
		StartingHandLimit: 4,
	}
}

// TestNewGame 初始状态
func TestNewGame(t *testing.T) {
	s, events, err := NewGame(testConfig(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if s.Phase != PhaseAttack {
		t.Fatalf("初始阶段应为 attack, 得到 %s", s.Phase)
	}
	if s.AttackerIndex != 0 {
		t.Fatalf("初始进攻方应为 0, 得到 %d", s.AttackerIndex)
	}
	if s.Players[0].Cakes != 3 || s.Players[1].Cakes != 4 {
		t.Fatalf("初始蛋糕应为 3/4, 得到 %d/%d", s.Players[0].Cakes, s.Players[1].Cakes)
	}
	if len(s.Players[0].Hand) != 4 || len(s.Players[1].Hand) != 4 {
		t.Fatalf("初始手牌应为 4/4, 得到 %d/%d", len(s.Players[0].Hand), len(s.Players[1].Hand))
	}
	if len(s.CardNames) != 40 {
		t.Fatalf("总牌数应为 40(基础29+特殊11), 得到 %d", len(s.CardNames))
	}
	if len(s.Deck)+len(s.Players[0].Hand)+len(s.Players[1].Hand) != 40 {
		t.Fatalf("牌堆+手牌应为 40, 得到 %d", len(s.Deck)+len(s.Players[0].Hand)+len(s.Players[1].Hand))
	}
	hasGameStarted := false
	hasBoutStarted := false
	for _, evt := range events {
		if evt.Type == EventGameStarted {
			hasGameStarted = true
		}
		if evt.Type == EventBoutStarted {
			hasBoutStarted = true
		}
	}
	if !hasGameStarted || !hasBoutStarted {
		t.Fatal("缺少 game_started / bout_started 事件")
	}
}

// TestClaimAndAccept 出牌并接受, 结算伤害
func TestClaimAndAccept(t *testing.T) {
	s, _, err := NewGame(testConfig(), 7)
	if err != nil {
		t.Fatal(err)
	}
	// 进攻方声明 soldier(可以撒谎), 出一张牌
	if _, err := Step(s, 0, &Action{Type: ActionClaim, HandIndices: []int{0}, Claim: "soldier"}); err != nil {
		t.Fatal(err)
	}
	if s.Phase != PhaseBlock {
		t.Fatalf("声明后应进入 block, 得到 %s", s.Phase)
	}
	before0, before1 := s.Players[0].Cakes, s.Players[1].Cakes
	if _, err := Step(s, 1, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	// 接受后, 进攻方抢走 1 块蛋糕
	if s.Players[0].Cakes != before0+1 || s.Players[1].Cakes != before1-1 {
		t.Fatalf("蛋糕结算错误: %d/%d -> %d/%d", before0, before1, s.Players[0].Cakes, s.Players[1].Cakes)
	}
	if s.Phase != PhaseAttack {
		t.Fatalf("结算后应回到 attack, 得到 %s", s.Phase)
	}
	if s.AttackerIndex != 1 {
		t.Fatalf("结算后进攻方应换人, 得到 %d", s.AttackerIndex)
	}
}

// TestChallengeMismatch 质疑抓到说谎 -> 挑战者获胜
func TestChallengeMismatch(t *testing.T) {
	// 手动构造状态: 声明 wizard 但实际打出的牌是 soldier
	s := &State{
		Phase:         PhaseReview,
		AttackerIndex: 0,
		CardNames:     []string{"soldier", "wizard"},
		Players: [2]Player{
			{Hand: []int{1}, HandLimit: 4, Cakes: 3},
			{Hand: []int{0}, HandLimit: 4, Cakes: 4},
		},
		AttackingClaim: &Claim{Claim: "wizard", CardIDs: []int{0}},
		BlockingClaim:  &Claim{Claim: "defender", CardIDs: []int{0}},
		Config:         testConfig(),
	}
	result := checkClaim(s)
	if !result.success {
		t.Fatal("实际是 soldier 却声明 wizard, 质疑应成功")
	}
}

// TestChallengeMatch 声明属实 -> 质疑者输
func TestChallengeMatch(t *testing.T) {
	s := &State{
		Phase:         PhaseReview,
		AttackerIndex: 0,
		CardNames:     []string{"soldier", "wizard"},
		Players: [2]Player{
			{Hand: []int{1}, HandLimit: 4, Cakes: 3},
			{Hand: []int{0}, HandLimit: 4, Cakes: 4},
		},
		AttackingClaim: &Claim{Claim: "soldier", CardIDs: []int{0}},
		BlockingClaim:  &Claim{Claim: "wizard", CardIDs: []int{1}},
		Config:         testConfig(),
	}
	result := checkClaim(s)
	if result.success {
		t.Fatal("声明属实, 质疑应失败")
	}
}

// TestBaacratesTransform 咩格拉底+狼爵士被质疑时变为科学家
func TestBaacratesTransform(t *testing.T) {
	s := &State{
		Phase:         PhaseReview,
		AttackerIndex: 0,
		CardNames:     []string{"baacrates", "wolfy", "scientist"},
		Players: [2]Player{
			{Hand: []int{2}, HandLimit: 4, Cakes: 3},
			{Hand: nil, HandLimit: 4, Cakes: 4},
		},
		AttackingClaim: &Claim{Claim: "scientist", CardIDs: []int{0, 1}},
		BlockingClaim:  &Claim{Claim: "scientist", CardIDs: []int{2}},
		Config:         testConfig(),
	}
	result := checkClaim(s)
	if result.success {
		t.Fatal("咩格拉底与狼爵士应变身为科学家, 声明应属实")
	}
}

// TestAgentUTransform 特工U变为对手手牌中的声明牌
func TestAgentUTransform(t *testing.T) {
	s := &State{
		Phase:         PhaseReview,
		AttackerIndex: 0,
		CardNames:     []string{"agent_u", "soldier"},
		Players: [2]Player{
			{Hand: []int{1}, HandLimit: 4, Cakes: 3},
			{Hand: nil, HandLimit: 4, Cakes: 4},
		},
		AttackingClaim: &Claim{Claim: "soldier", CardIDs: []int{1}},
		BlockingClaim:  &Claim{Claim: "soldier", CardIDs: []int{0}},
		Config:         testConfig(),
	}
	result := checkClaim(s)
	if result.success {
		t.Fatal("对手手牌中有 soldier, 特工U应变身, 声明应属实")
	}
	if result.extraEvent != EffectTransformToOppHand {
		t.Fatalf("应触发 reveal_challenger_hand, 得到 %s", result.extraEvent)
	}
}

// TestPierrotTransform 绵顿变为上回合声明的攻击牌
func TestPierrotTransform(t *testing.T) {
	s := &State{
		Phase:         PhaseBlock,
		AttackerIndex: 0,
		CardNames:     []string{"pierrot", "wizard"},
		Players: [2]Player{
			{Hand: nil, HandLimit: 4, Cakes: 3, LastAttackingClaim: ""},
			{Hand: []int{1}, HandLimit: 4, Cakes: 4, LastAttackingClaim: "wizard"},
		},
		AttackingClaim: &Claim{Claim: "wizard", CardIDs: []int{1}},
		Config:         testConfig(),
	}
	result := checkClaim(s)
	if result.success {
		t.Fatal("绵顿应变为上回合声明的 wizard, 声明应属实")
	}
}

// TestSummonerPick 召唤师查看对手手牌并按同名牌数量抢蛋糕
func TestSummonerPick(t *testing.T) {
	s, _, err := NewGame(testConfig(), 3)
	if err != nil {
		t.Fatal(err)
	}
	// 找到一张 summoner 或直接声明 summoner(说谎)
	if _, err := Step(s, 0, &Action{Type: ActionClaim, HandIndices: []int{0}, Claim: "summoner"}); err != nil {
		t.Fatal(err)
	}
	// 防守方接受
	if _, err := Step(s, 1, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	if s.Phase != PhasePick {
		t.Fatalf("召唤师应进入 pick 阶段, 得到 %s", s.Phase)
	}
	specs := LegalActions(s, 0)
	var pickSpec *ActionSpec
	for i := range specs {
		if specs[i].Type == ActionPick {
			pickSpec = &specs[i]
			break
		}
	}
	if pickSpec == nil || len(pickSpec.PickFrom) == 0 {
		t.Fatal("pick 阶段应提供选牌选项")
	}
	before := s.Players[0].Cakes
	if _, err := Step(s, 0, &Action{Type: ActionPick, PickIndices: []int{0}}); err != nil {
		t.Fatal(err)
	}
	if s.Players[0].Cakes < before {
		t.Fatal("召唤师不应减少自己的蛋糕")
	}
}

// TestConcede 认输结束游戏
func TestConcede(t *testing.T) {
	s, _, err := NewGame(testConfig(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Step(s, 0, &Action{Type: ActionConcede}); err != nil {
		t.Fatal(err)
	}
	if s.GameEnded == nil || s.GameEnded.Winner != 1 {
		t.Fatalf("认输后对手应获胜, 得到 %+v", s.GameEnded)
	}
}

// TestFullMatch 完整对局直至分出胜负
func TestFullMatch(t *testing.T) {
	s, _, err := NewGame(testConfig(), 99)
	if err != nil {
		t.Fatal(err)
	}
	rounds := 0
	for s.GameEnded == nil && rounds < 500 {
		rounds++
		player := phasingPlayer(s)
		specs := LegalActions(s, player)
		if len(specs) == 0 {
			t.Fatalf("第 %d 步没有合法动作", rounds)
		}
		// 简单策略: 有 claim 就出第一张牌声明第一个可声明的牌名, 否则 pass
		var action *Action
		used := false
		for i := range specs {
			if specs[i].Type == ActionClaim && !used {
				handLen := len(s.Players[player].Hand)
				if handLen > 0 {
					action = &Action{Type: ActionClaim, HandIndices: []int{0}, Claim: specs[i].ClaimFrom[0]}
					used = true
				}
				break
			}
		}
		if action == nil {
			for i := range specs {
				if specs[i].Type == ActionPass {
					action = &Action{Type: ActionPass}
					break
				}
				if specs[i].Type == ActionChallenge {
					action = &Action{Type: ActionChallenge}
					break
				}
				if specs[i].Type == ActionPick {
					action = &Action{Type: ActionPick, PickIndices: []int{0}}
					break
				}
			}
		}
		if action == nil {
			action = &Action{Type: ActionConcede}
		}
		if _, err := Step(s, player, action); err != nil {
			t.Fatalf("第 %d 步动作 %+v 失败: %v", rounds, action, err)
		}
	}
	if s.GameEnded == nil {
		t.Fatal("对局未能在 500 步内结束")
	}
	if len(s.BoutWinners) < s.Config.RoundsToWin {
		t.Fatalf("胜场数 %d 少于 %d", len(s.BoutWinners), s.Config.RoundsToWin)
	}
}

// cardIndex 返回第 occurrence 张同名卡在牌池中的索引
func cardIndex(names []string, name string, occurrence int) int {
	count := 0
	for i, n := range names {
		if n == name {
			if count == occurrence {
				return i
			}
			count++
		}
	}
	return -1
}

// newManualState 构造指定攻防声明的对局状态(block 阶段, 攻击方 0)
func newManualState() *State {
	names := append(append([]string{}, BaseCardList...), SpecialCardList...)
	return &State{
		Phase:         PhaseBlock,
		AttackerIndex: 0,
		CardNames:     names,
		Config:        testConfig(),
		RNG:           NewRNG(7),
		Deck:          rangeInt(len(names)),
		Players: [2]Player{
			{HandLimit: 4, Cakes: 3},
			{HandLimit: 4, Cakes: 4},
		},
		AttackingClaim: &Claim{Claim: "soldier", CardIDs: []int{0}},
	}
}

// TestScoutAttacksAgain 斥候抢1蛋糕且攻击方再行动
func TestScoutAttacksAgain(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "scout", CardIDs: []int{cardIndex(s.CardNames, "scout", 0)}}
	s.Players[0].Hand = []int{13, 5}
	s.Players[1].Hand = []int{9, 16}
	events, err := Step(s, 1, &Action{Type: ActionPass})
	if err != nil {
		t.Fatal(err)
	}
	if s.Players[0].Cakes != 4 || s.Players[1].Cakes != 3 {
		t.Fatalf("斥候应抢 1 块蛋糕: %d/%d", s.Players[0].Cakes, s.Players[1].Cakes)
	}
	if s.AttackerIndex != 0 {
		t.Fatalf("斥候应让攻击方再行动, 攻击方变为 %d", s.AttackerIndex)
	}
	if s.Phase != PhaseAttack {
		t.Fatalf("结算后应回到 attack, 得到 %s", s.Phase)
	}
	_ = events
}

// TestOracleRevealAndAgain 神谕师查看对手手牌且再行动
func TestOracleRevealAndAgain(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "oracle", CardIDs: []int{cardIndex(s.CardNames, "oracle", 0)}}
	s.Players[0].Hand = []int{13, 5}
	s.Players[1].Hand = []int{9, 16}
	events, err := Step(s, 1, &Action{Type: ActionPass})
	if err != nil {
		t.Fatal(err)
	}
	if s.AttackerIndex != 0 {
		t.Fatalf("神谕师应让攻击方再行动, 攻击方变为 %d", s.AttackerIndex)
	}
	hasReveal := false
	for _, evt := range events {
		if evt.Type == EventHandRevealed {
			hasReveal = true
		}
	}
	if !hasReveal {
		t.Fatal("神谕师应触发查看手牌事件")
	}
}

// TestOracleRevealVisibility 神谕师查看手牌只对查看者可见
func TestOracleRevealVisibility(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "oracle", CardIDs: []int{cardIndex(s.CardNames, "oracle", 0)}}
	s.Players[0].Hand = []int{cardIndex(s.CardNames, "soldier", 0)}
	s.Players[1].Hand = []int{cardIndex(s.CardNames, "defender", 0)}
	events, err := Step(s, 1, &Action{Type: ActionPass})
	if err != nil {
		t.Fatal(err)
	}
	// 攻击方(0, 查看者)应看到 hand_revealed 及牌名
	f0 := FilterEvents(events, 0, s.CardNames)
	f1 := FilterEvents(events, 1, s.CardNames)
	hasReveal0 := false
	for _, evt := range f0 {
		if evt.Type == EventHandRevealed {
			hasReveal0 = true
			if len(evt.CardNames) == 0 {
				t.Fatal("查看者应看到被查看手牌的牌名")
			}
		}
	}
	if !hasReveal0 {
		t.Fatal("查看者应收到 hand_revealed 事件")
	}
	for _, evt := range f1 {
		if evt.Type == EventHandRevealed {
			t.Fatal("被查看者不应收到 hand_revealed 事件(不应看到自己的牌被展示)")
		}
	}
}

// TestSummonerRevealVisibility 召唤师查看手牌只对选牌者可见
func TestSummonerRevealVisibility(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "summoner", CardIDs: []int{cardIndex(s.CardNames, "summoner", 0)}}
	s.Players[1].Cakes = 10
	s.Players[0].Hand = []int{cardIndex(s.CardNames, "soldier", 0)}
	s.Players[1].Hand = []int{cardIndex(s.CardNames, "soldier", 0)}
	if _, err := Step(s, 1, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	events, err := Step(s, 0, &Action{Type: ActionPick, PickIndices: []int{0}})
	if err != nil {
		t.Fatal(err)
	}
	f0 := FilterEvents(events, 0, s.CardNames)
	f1 := FilterEvents(events, 1, s.CardNames)
	hasReveal0 := false
	for _, evt := range f0 {
		if evt.Type == EventHandRevealed {
			hasReveal0 = true
		}
	}
	if !hasReveal0 {
		t.Fatal("召唤师应看到对手手牌")
	}
	for _, evt := range f1 {
		if evt.Type == EventHandRevealed {
			t.Fatal("被查看者不应看到 hand_revealed 事件")
		}
	}
}

// TestQuartermasterHandLimit 军需官手牌上限+1
func TestQuartermasterHandLimit(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "quartermaster", CardIDs: []int{cardIndex(s.CardNames, "quartermaster", 0)}}
	s.Players[0].Hand = []int{13, 5}
	s.Players[1].Hand = []int{9, 16}
	if _, err := Step(s, 1, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	if s.Players[0].HandLimit != 5 {
		t.Fatalf("军需官应让手牌上限+1, 得到 %d", s.Players[0].HandLimit)
	}
}

// TestPriestBlocksAndTakesCard 牧师挡住攻击, 取消效果并收入一张攻击牌
func TestPriestBlocksAndTakesCard(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "wizard", CardIDs: []int{cardIndex(s.CardNames, "wizard", 0)}}
	s.BlockingClaim = &Claim{Claim: "priest", CardIDs: []int{cardIndex(s.CardNames, "priest", 0)}}
	s.Phase = PhaseReview
	s.Players[0].Hand = []int{5, 9}
	s.Players[1].Hand = []int{16, 16}
	if _, err := Step(s, 0, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	if s.Players[0].Cakes != 3 || s.Players[1].Cakes != 4 {
		t.Fatalf("牧师应完全挡住法师伤害: %d/%d", s.Players[0].Cakes, s.Players[1].Cakes)
	}
	if len(s.Players[1].Hand) < 3 {
		t.Fatalf("牧师应收入一张攻击牌, 手牌为 %d", len(s.Players[1].Hand))
	}
}

// TestAngelBlocksAll 天使挡住所有攻击并让对手再行动
func TestAngelBlocksAll(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "wizard", CardIDs: []int{cardIndex(s.CardNames, "wizard", 0), cardIndex(s.CardNames, "wizard", 1)}}
	s.BlockingClaim = &Claim{Claim: "angel", CardIDs: []int{cardIndex(s.CardNames, "angel", 0)}}
	s.Phase = PhaseReview
	s.Players[0].Hand = []int{5, 9}
	s.Players[1].Hand = []int{16, 16}
	if _, err := Step(s, 0, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	if s.Players[0].Cakes != 3 || s.Players[1].Cakes != 4 {
		t.Fatalf("天使应挡下所有伤害: %d/%d", s.Players[0].Cakes, s.Players[1].Cakes)
	}
	if s.AttackerIndex != 0 {
		t.Fatalf("天使应让对手(攻击方)再行动, 攻击方变为 %d", s.AttackerIndex)
	}
}

// TestSummonerStealsTwo 召唤师按对手同名牌数量抢蛋糕
func TestSummonerStealsTwo(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "summoner", CardIDs: []int{cardIndex(s.CardNames, "summoner", 0)}}
	s.Players[1].Cakes = 10
	s.Players[0].Hand = []int{13, 5}
	s.Players[1].Hand = []int{0, 1, 9} // 两张 soldier
	if _, err := Step(s, 1, &Action{Type: ActionPass}); err != nil {
		t.Fatal(err)
	}
	if s.Phase != PhasePick {
		t.Fatalf("召唤师应进入 pick 阶段, 得到 %s", s.Phase)
	}
	if _, err := Step(s, 0, &Action{Type: ActionPick, PickIndices: []int{0}}); err != nil {
		t.Fatal(err)
	}
	// pickFrom 按卡牌顺序, 0=soldier; 对手两张 soldier, 每张抢 2 块
	if s.Players[0].Cakes != 7 || s.Players[1].Cakes != 6 {
		t.Fatalf("召唤师应抢 4 块蛋糕: %d/%d", s.Players[0].Cakes, s.Players[1].Cakes)
	}
}

// TestSpecialCardClaimBlacklist 特殊卡效果不影响基础规则
func TestSpecialCardClaimBlacklist(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "soldier", CardIDs: []int{0}}
	s.Players[0].Hand = []int{13, 5}
	s.Players[1].Hand = []int{9, 16}
	// 防守方可以声明盾卫挡士兵
	legal := LegalActions(s, 1)
	foundDefender := false
	for _, spec := range legal {
		if spec.Type == ActionClaim {
			for _, c := range spec.ClaimFrom {
				if c == "defender" {
					foundDefender = true
				}
			}
		}
	}
	if !foundDefender {
		t.Fatal("防守方应可声明盾卫")
	}
}
