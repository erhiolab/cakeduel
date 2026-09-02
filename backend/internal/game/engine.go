package game

import (
	"errors"
	"fmt"
	"sort"
)

// NewGame 创建新游戏, 返回初始状态与初始事件
func NewGame(cfg GameConfig, seed uint32) (*State, []Event, error) {
	if cfg.RoundsToWin <= 0 {
		cfg.RoundsToWin = 3
	}
	if cfg.StartingHandLimit <= 0 {
		cfg.StartingHandLimit = StartingHandLimit
	}
	if cfg.SpecialCardsToAdd > len(SpecialCardList) {
		cfg.SpecialCardsToAdd = len(SpecialCardList)
	}
	if cfg.SpecialCardsToAdd < 0 {
		cfg.SpecialCardsToAdd = 0
	}

	rng := NewRNG(seed)
	cardList := append([]string{}, BaseCardList...)
	if cfg.DeckConfig != nil {
		// 自定义卡组: 特殊卡按配置数量加入(可全部为 0 = 纯基础卡)
		cardList = append(cardList, expandDeckConfig(cfg.DeckConfig)...)
	} else if cfg.SpecialCardsToAdd > 0 {
		indices := rangeInt(len(SpecialCardList))
		sampled := rng.Sample(indices, cfg.SpecialCardsToAdd)
		for _, idx := range sampled {
			cardList = append(cardList, SpecialCardList[idx])
		}
	}

	s := &State{
		Phase:         PhaseAttack,
		AttackerIndex: 0,
		Players: [2]Player{
			{HandLimit: cfg.StartingHandLimit, Cakes: 3},
			{HandLimit: cfg.StartingHandLimit, Cakes: 4},
		},
		RNG:       rng,
		CardNames: cardList,
		Config:    cfg,
	}
	w := &EventWriter{State: s}
	w.public(Event{Type: EventGameStarted, Config: &cfg})
	startBout(w)
	return s, w.Events, nil
}

// Step 执行玩家动作, 原地更新状态并返回事件
func Step(s *State, playerIndex int, action *Action) ([]Event, error) {
	if s.GameEnded != nil {
		return nil, errors.New("游戏已结束")
	}
	phasing := phasingPlayer(s)
	if phasing != playerIndex {
		return nil, errors.New("还没轮到你的回合")
	}

	legal := legalActions(s, playerIndex)
	var spec *ActionSpec
	for i := range legal {
		if legal[i].Type == action.Type {
			spec = &legal[i]
			break
		}
	}
	if spec == nil {
		return nil, fmt.Errorf("当前阶段不能执行: %s", action.Type)
	}
	if err := validateAction(s, spec, action); err != nil {
		return nil, err
	}

	w := &EventWriter{State: s}
	var err error
	switch action.Type {
	case ActionClaim:
		err = doClaim(w, action)
	case ActionPass:
		err = doPass(w)
	case ActionChallenge:
		err = doChallenge(w)
	case ActionPick:
		err = doPick(w, action)
	case ActionConcede:
		err = doConcede(w)
	}
	if err != nil {
		return nil, err
	}
	s.Frame++
	return w.Events, nil
}

// doClaim 出牌声明
func doClaim(w *EventWriter, action *Action) error {
	s := w.State
	player := phasingPlayer(s)
	newHand, removed, err := removeIndices(s.Players[player].Hand, action.HandIndices)
	if err != nil {
		return err
	}
	s.Players[player].Hand = newHand
	claim := &Claim{Claim: action.Claim, CardIDs: removed}
	pile := ""
	if s.AttackingClaim == nil {
		s.AttackingClaim = claim
		s.Players[player].LastAttackingClaim = action.Claim
		pile = "attack_pile"
	} else if s.BlockingClaim == nil {
		s.BlockingClaim = claim
		pile = "block_pile"
	} else {
		return errors.New("非法状态: 两个声明都已存在")
	}
	w.private(player, Event{
		Type:    EventClaimMade,
		Player:  intPtr(player),
		Pile:    pile,
		Claim:   action.Claim,
		cardIDs: removed,
	})
	if pile == "attack_pile" {
		setPhase(w, PhaseBlock)
	} else {
		setPhase(w, PhaseReview)
	}
	return nil
}

// doPass 过牌/接受
func doPass(w *EventWriter) error {
	s := w.State
	player := phasingPlayer(s)
	w.public(Event{Type: EventPassMade, Player: intPtr(player)})

	if s.AttackingClaim == nil {
		// 进攻方连续过牌 -> 本局结束, 蛋糕多者获胜
		if s.LastAttackPassed {
			if s.Players[0].Cakes > s.Players[1].Cakes {
				endBout(w, 0)
			} else {
				endBout(w, 1)
			}
			return nil
		}
		s.LastAttackPassed = true
		nextAttacker(w)
		return nil
	}

	// 接受声明, 结算攻防
	s.LastAttackPassed = false
	attackerDef := CardDefs[s.AttackingClaim.Claim]
	if attackerDef.Attack == nil {
		return errors.New("进攻声明不是攻击牌")
	}
	attackerCount := len(s.AttackingClaim.CardIDs)
	blockerCount := 0
	if s.BlockingClaim != nil {
		blockerCount = len(s.BlockingClaim.CardIDs)
	}
	unblocked := attackerCount
	effectsToResolve := attackerCount

	if s.BlockingClaim != nil {
		blockerDef := CardDefs[s.BlockingClaim.Claim]
		if blockerDef.Block == nil {
			return errors.New("防守声明不是防守牌")
		}
		if blockerDef.Block.BlockMultipleCards {
			unblocked = 0
			if blockerDef.Block.BlockExtraEffects {
				effectsToResolve = 0
			}
		} else {
			u := attackerCount - blockerCount
			if u < 0 {
				u = 0
			}
			unblocked = u
			if blockerDef.Block.BlockExtraEffects {
				effectsToResolve = u
			}
		}
		// 触发防守牌额外效果
		for i := 0; i < blockerCount; i++ {
			for _, effect := range blockerDef.Block.ExtraEffects {
				applyExtraEffect(w, effect)
			}
		}
	}

	// 每张未被挡住的攻击牌造成伤害
	for i := 0; i < unblocked; i++ {
		damage := attackerDef.Attack.Damage
		if damage > 0 {
			transferCakes(w, s.AttackerIndex, opponent(s.AttackerIndex), damage)
		}
	}
	// 每张未被挡住的攻击牌触发效果
	for i := 0; i < effectsToResolve; i++ {
		for _, effect := range attackerDef.Attack.ExtraEffects {
			applyExtraEffect(w, effect)
		}
		if attackerDef.Attack.TriggerPick != nil {
			s.PickPhaseEffects = append(s.PickPhaseEffects, PickEffect{
				Player:   s.AttackerIndex,
				Effect:   *attackerDef.Attack.TriggerPick,
				PickFrom: claimableAll(s),
			})
		}
	}

	// 结算: 弃置牌堆, 检查蛋糕, 进入选牌或下一回合
	discardPiles(w)
	for _, p := range []int{0, 1} {
		if s.Players[p].Cakes <= 0 {
			endBout(w, opponent(p))
			return nil
		}
	}
	if len(s.PickPhaseEffects) > 0 {
		setPhase(w, PhasePick)
	} else {
		nextAttacker(w)
	}
	return nil
}

// doChallenge 质疑
func doChallenge(w *EventWriter) error {
	s := w.State
	result := checkClaim(s)
	challenger := phasingPlayer(s)
	evt := Event{
		Type:        EventChallengeMade,
		Challenger:  challenger,
		ClaimedCard: result.claimed,
		Success:     result.success,
		revealedIDs: result.ids,
		transformed: result.transformed,
	}
	w.public(evt)
	if result.extraEvent == EffectTransformToOppHand {
		// 只有质疑者能看到被翻开的手牌
		w.private(challenger, Event{
			Type:    EventHandRevealed,
			Player:  intPtr(challenger),
			cardIDs: append([]int{}, s.Players[challenger].Hand...),
		})
	}
	if result.success {
		// 质疑成功: 挑战者赢下本局
		endBout(w, challenger)
	} else {
		// 声明全部属实: 质疑者输掉本局
		endBout(w, opponent(challenger))
	}
	return nil
}

// doPick 选牌效果
func doPick(w *EventWriter, action *Action) error {
	s := w.State
	if len(s.PickPhaseEffects) == 0 {
		return errors.New("当前没有选牌效果")
	}
	effect := s.PickPhaseEffects[0]
	s.PickPhaseEffects = s.PickPhaseEffects[1:]
	picks := make([]string, 0, len(action.PickIndices))
	for _, idx := range action.PickIndices {
		if idx < 0 || idx >= len(effect.PickFrom) {
			return errors.New("选牌索引越界")
		}
		picks = append(picks, effect.PickFrom[idx])
	}
	w.public(Event{Type: EventPickMade, Player: intPtr(effect.Player), PickType: effect.Effect.Type, Picks: picks})

	switch effect.Effect.Type {
	case "name_peek_steal_two":
		opp := opponent(effect.Player)
		// 只有召唤师(选牌者)能看到对方手牌
		w.private(effect.Player, Event{
			Type:    EventHandRevealed,
			Player:  intPtr(opp),
			cardIDs: append([]int{}, s.Players[opp].Hand...),
		})
		count := 0
		for _, cardID := range s.Players[opp].Hand {
			if cardID >= 0 && cardID < len(s.CardNames) && s.CardNames[cardID] == picks[0] {
				count += 2
			}
		}
		if count > 0 {
			transferCakes(w, effect.Player, opp, count)
		}
	default:
		return fmt.Errorf("未知的选牌效果: %s", effect.Effect.Type)
	}

	for _, p := range []int{0, 1} {
		if s.Players[p].Cakes <= 0 {
			endBout(w, opponent(p))
			return nil
		}
	}
	if len(s.PickPhaseEffects) > 0 {
		setPhase(w, PhasePick)
	} else {
		nextAttacker(w)
	}
	return nil
}

// doConcede 认输
func doConcede(w *EventWriter) error {
	s := w.State
	player := phasingPlayer(s)
	w.public(Event{Type: EventConcedeMade, Player: intPtr(player)})
	endGame(w, opponent(player))
	return nil
}

// applyExtraEffect 应用额外效果
func applyExtraEffect(w *EventWriter, effect string) {
	s := w.State
	switch effect {
	case EffectAttacksAgain:
		s.NextAttackerIndexOverride = append(s.NextAttackerIndexOverride, s.AttackerIndex)
	case EffectIncreaseHandSize:
		s.Players[s.AttackerIndex].HandLimit++
	case EffectTakeAttackPileCard:
		if s.AttackingClaim != nil && len(s.AttackingClaim.CardIDs) > 0 {
			idx := s.RNG.NextInt(len(s.AttackingClaim.CardIDs))
			taken := []int{s.AttackingClaim.CardIDs[idx]}
			s.AttackingClaim.CardIDs = append(s.AttackingClaim.CardIDs[:idx], s.AttackingClaim.CardIDs[idx+1:]...)
			s.Players[opponent(s.AttackerIndex)].Hand = append(s.Players[opponent(s.AttackerIndex)].Hand, taken...)
		}
	case EffectRevealOpponentHand:
		opp := opponent(s.AttackerIndex)
		// 只有查看者(攻击方)能看到对方手牌
		w.private(s.AttackerIndex, Event{
			Type:    EventHandRevealed,
			Player:  intPtr(opp),
			cardIDs: append([]int{}, s.Players[opp].Hand...),
		})
	default:
		// 忽略未知效果(例如 promo 卡)
	}
}

// discardPiles 弃置攻防牌堆并触发狼爵士嘲讽
func discardPiles(w *EventWriter) {
	s := w.State
	if s.AttackingClaim != nil {
		if containsWolfy(s, s.AttackingClaim.CardIDs) {
			w.public(Event{Type: EventWolfyTaunt, Player: intPtr(s.AttackerIndex)})
		}
		w.private(s.AttackerIndex, Event{
			Type:    EventCardDiscarded,
			Zone:    "attack_pile",
			cardIDs: append([]int{}, s.AttackingClaim.CardIDs...),
		})
		s.Discard = append(s.Discard, s.AttackingClaim.CardIDs...)
		s.AttackingClaim = nil
	}
	if s.BlockingClaim != nil {
		if containsWolfy(s, s.BlockingClaim.CardIDs) {
			w.public(Event{Type: EventWolfyTaunt, Player: intPtr(opponent(s.AttackerIndex))})
		}
		w.private(opponent(s.AttackerIndex), Event{
			Type:    EventCardDiscarded,
			Zone:    "block_pile",
			cardIDs: append([]int{}, s.BlockingClaim.CardIDs...),
		})
		s.Discard = append(s.Discard, s.BlockingClaim.CardIDs...)
		s.BlockingClaim = nil
	}
}

// containsWolfy 牌堆中是否包含狼爵士
func containsWolfy(s *State, cardIDs []int) bool {
	for _, id := range cardIDs {
		if id >= 0 && id < len(s.CardNames) && s.CardNames[id] == "wolfy" {
			return true
		}
	}
	return false
}

// transferCakes 蛋糕转移 from -> to
func transferCakes(w *EventWriter, to, from, amount int) {
	s := w.State
	amount = clamp(amount, 0, s.Players[from].Cakes)
	if amount <= 0 {
		return
	}
	s.Players[to].Cakes += amount
	s.Players[from].Cakes -= amount
	w.public(Event{
		Type:       EventCakesTransfer,
		From:       from,
		To:         to,
		Amount:     amount,
		CakesAfter: [2]int{s.Players[0].Cakes, s.Players[1].Cakes},
	})
}

// setPhase 切换阶段
func setPhase(w *EventWriter, phase string) {
	s := w.State
	s.Phase = phase
	w.public(Event{Type: EventPhaseChanged, Player: intPtr(phasingPlayer(s)), Phase: phase})
}

// nextAttacker 轮到下一位进攻方
func nextAttacker(w *EventWriter) {
	s := w.State
	drawToLimits(w)
	if len(s.NextAttackerIndexOverride) > 0 {
		s.AttackerIndex = s.NextAttackerIndexOverride[0]
		s.NextAttackerIndexOverride = s.NextAttackerIndexOverride[1:]
	} else {
		s.AttackerIndex = opponent(s.AttackerIndex)
	}
	setPhase(w, PhaseAttack)
}

// drawToLimits 双方补满手牌
func drawToLimits(w *EventWriter) {
	s := w.State
	for _, p := range []int{s.AttackerIndex, opponent(s.AttackerIndex)} {
		r := s.Players[p].HandLimit - len(s.Players[p].Hand)
		if r > 0 {
			drawCards(w, p, r)
		}
	}
}

// drawCards 从牌堆抽牌
func drawCards(w *EventWriter, player, n int) {
	s := w.State
	if n <= 0 {
		return
	}
	if n > len(s.Deck) {
		n = len(s.Deck)
	}
	drawn := append([]int{}, s.Deck[:n]...)
	s.Deck = s.Deck[n:]
	w.private(player, Event{Type: EventCardDrawn, Zone: "deck", Player: intPtr(player), cardIDs: drawn})
	s.Players[player].Hand = append(s.Players[player].Hand, drawn...)
}

// endBout 结束本局
func endBout(w *EventWriter, winner int) {
	s := w.State
	s.BoutWinners = append(s.BoutWinners, winner)
	w.public(Event{Type: EventBoutEnded, Winner: winner})
	count := 0
	for _, bw := range s.BoutWinners {
		if bw == winner {
			count++
		}
	}
	if count >= s.Config.RoundsToWin {
		endGame(w, winner)
	} else {
		startBout(w)
	}
}

// endGame 结束整场游戏
func endGame(w *EventWriter, winner int) {
	s := w.State
	s.GameEnded = &GameEnded{Winner: winner}
	w.public(Event{Type: EventGameEnded, Winner: winner})
}

// startBout 开始新一局
func startBout(w *EventWriter) {
	s := w.State
	deck := rangeInt(len(s.CardNames))
	s.RNG.Shuffle(deck)
	s.Deck = deck
	w.push([2]bool{false, false}, Event{Type: EventDeckShuffled})

	s.LastAttackPassed = false
	s.Discard = nil
	s.AttackingClaim = nil
	s.BlockingClaim = nil
	s.NextAttackerIndexOverride = nil
	s.PickPhaseEffects = nil
	s.Players[0].Hand = nil
	s.Players[1].Hand = nil
	s.Players[0].LastAttackingClaim = ""
	s.Players[1].LastAttackingClaim = ""
	if len(s.BoutWinners) > 0 {
		s.AttackerIndex = opponent(last(s.BoutWinners))
	} else {
		s.AttackerIndex = 0
	}
	attacker := s.AttackerIndex
	defender := opponent(attacker)
	s.Players[attacker].Cakes = 3
	s.Players[attacker].HandLimit = s.Config.StartingHandLimit
	s.Players[attacker].ClaimBlacklist = nil
	s.Players[defender].Cakes = 4
	s.Players[defender].HandLimit = s.Config.StartingHandLimit
	s.Players[defender].ClaimBlacklist = nil
	w.public(Event{
		Type:          EventBoutStarted,
		AttackerIndex: attacker,
		CakesAfter:    [2]int{s.Players[0].Cakes, s.Players[1].Cakes},
	})
	setPhase(w, PhaseAttack)
	drawToLimits(w)
}

// removeIndices 按索引(升序)移除元素, 返回新切片与被移除的值
func removeIndices(items []int, indices []int) ([]int, []int, error) {
	if len(indices) == 0 {
		return items, nil, errors.New("没有选择任何牌")
	}
	sorted := append([]int{}, indices...)
	sort.Ints(sorted)
	// 校验重复
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			return items, nil, errors.New("同一张牌被选择了多次")
		}
	}
	out := append([]int{}, items...)
	removed := make([]int, 0, len(sorted))
	for i := len(sorted) - 1; i >= 0; i-- {
		idx := sorted[i]
		if idx < 0 || idx >= len(out) {
			return items, nil, errors.New("手牌索引越界")
		}
		removed = append(removed, out[idx])
		out = append(out[:idx], out[idx+1:]...)
	}
	// 反转保持出牌顺序
	for i, j := 0, len(removed)-1; i < j; i, j = i+1, j-1 {
		removed[i], removed[j] = removed[j], removed[i]
	}
	return out, removed, nil
}

func intPtr(v int) *int {
	return &v
}
