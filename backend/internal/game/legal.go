package game

import "errors"

// ActionSpec 合法动作描述
type ActionSpec struct {
	Type                  string   `json:"type"`
	ClaimFrom             []string `json:"claimFrom,omitempty"`
	AvailableHandIndices  []int    `json:"availableHandIndices,omitempty"`
	PickType              string   `json:"pickType,omitempty"`
	PickFrom              []string `json:"pickFrom,omitempty"`
	MinPicks              int      `json:"minPicks,omitempty"`
	MaxPicks              int      `json:"maxPicks,omitempty"`
}

// LegalActions 获取玩家的合法动作
func LegalActions(s *State, playerIndex int) []ActionSpec {
	if s.GameEnded != nil || phasingPlayer(s) != playerIndex {
		return nil
	}
	return legalActions(s, playerIndex)
}

// legalActions 获取玩家的合法动作(内部)
func legalActions(s *State, playerIndex int) []ActionSpec {
	hand := s.Players[playerIndex].Hand
	concede := ActionSpec{Type: ActionConcede}
	switch s.Phase {
	case PhaseAttack:
		out := make([]ActionSpec, 0, 3)
		if len(hand) > 0 {
			out = append(out, ActionSpec{
				Type:                 ActionClaim,
				ClaimFrom:            claimableForAttack(s),
				AvailableHandIndices: rangeInt(len(hand)),
			})
		}
		out = append(out, ActionSpec{Type: ActionPass})
		out = append(out, concede)
		return out
	case PhaseBlock:
		out := make([]ActionSpec, 0, 4)
		blockers := claimableForBlock(s)
		if len(blockers) > 0 {
			out = append(out, ActionSpec{
				Type:                 ActionClaim,
				ClaimFrom:            blockers,
				AvailableHandIndices: rangeInt(len(hand)),
			})
		}
		out = append(out, ActionSpec{Type: ActionPass})
		out = append(out, ActionSpec{Type: ActionChallenge})
		out = append(out, concede)
		return out
	case PhaseReview:
		return []ActionSpec{
			{Type: ActionPass},
			{Type: ActionChallenge},
			concede,
		}
	case PhasePick:
		out := make([]ActionSpec, 0, 2)
		if spec := pickSpec(s); spec != nil {
			out = append(out, *spec)
		}
		out = append(out, concede)
		return out
	}
	return nil
}

// pickSpec 选牌动作描述
func pickSpec(s *State) *ActionSpec {
	if len(s.PickPhaseEffects) == 0 {
		return nil
	}
	effect := s.PickPhaseEffects[0]
	if effect.Effect.Constraint != "single_card_name" {
		return nil
	}
	return &ActionSpec{
		Type:     ActionPick,
		PickType: effect.Effect.Type,
		PickFrom: effect.PickFrom,
		MinPicks: 1,
		MaxPicks: 1,
	}
}

// claimableAll 所有可声明的牌名(用于选牌)
func claimableAll(s *State) []string {
	names := make([]string, 0, len(s.CardNames))
	for _, name := range s.CardNames {
		if _, ok := CardDefs[name]; ok {
			names = append(names, name)
		}
	}
	return unique(names)
}

// claimableForAttack 进攻方可声明的牌名
func claimableForAttack(s *State) []string {
	blacklist := s.Players[s.AttackerIndex].ClaimBlacklist
	var out []string
	for _, name := range unique(s.CardNames) {
		def, ok := CardDefs[name]
		if !ok || def.Attack == nil || containsString(blacklist, name) {
			continue
		}
		out = append(out, name)
	}
	return out
}

// claimableForBlock 防守方可声明的牌名
func claimableForBlock(s *State) []string {
	if s.AttackingClaim == nil {
		return nil
	}
	attackType := CardDefs[s.AttackingClaim.Claim].Type
	blacklist := s.Players[opponent(s.AttackerIndex)].ClaimBlacklist
	var out []string
	for _, name := range unique(s.CardNames) {
		def, ok := CardDefs[name]
		if !ok || def.Block == nil || containsString(blacklist, name) {
			continue
		}
		blockType := def.Block.BlockType
		if blockType == "all_types" || blockType == attackType {
			out = append(out, name)
		}
	}
	return out
}

// validateAction 校验动作参数
func validateAction(s *State, spec *ActionSpec, action *Action) error {
	switch action.Type {
	case ActionClaim:
		if len(action.HandIndices) == 0 {
			return errors.New("没有选择任何牌")
		}
		if !containsString(spec.ClaimFrom, action.Claim) {
			return errors.New("不能声明该牌名")
		}
		seen := make(map[int]bool)
		handLen := len(s.Players[phasingPlayer(s)].Hand)
		for _, idx := range action.HandIndices {
			if idx < 0 || idx >= handLen {
				return errors.New("手牌索引越界")
			}
			if seen[idx] {
				return errors.New("同一张牌被选择了多次")
			}
			seen[idx] = true
		}
	case ActionPick:
		n := len(action.PickIndices)
		if n < spec.MinPicks || n > spec.MaxPicks {
			return errors.New("选牌数量不正确")
		}
		seen := make(map[int]bool)
		for _, idx := range action.PickIndices {
			if idx < 0 || idx >= len(spec.PickFrom) {
				return errors.New("选牌索引越界")
			}
			if seen[idx] {
				return errors.New("同一选项被选择了多次")
			}
			seen[idx] = true
		}
	}
	return nil
}

// claimResult 质疑检查结果
type claimResult struct {
	claimed     string
	success     bool
	ids         []int
	transformed map[int]string
	extraEvent  string
}

// checkClaim 翻开声明并检查是否符合
func checkClaim(s *State) claimResult {
	var claim *Claim
	if s.BlockingClaim != nil {
		claim = s.BlockingClaim
	} else {
		claim = s.AttackingClaim
	}
	transformed := make(map[int]string)
	extraEvent := ""
	challenger := phasingPlayer(s)

	for _, cardID := range claim.CardIDs {
		name := s.CardNames[cardID]
		def := CardDefs[name]
		switch def.TransformOnChallenge {
		case EffectTransformToOppHand:
			oppHand := s.Players[challenger].Hand
			hasClaim := false
			for _, id := range oppHand {
				if s.CardNames[id] == claim.Claim {
					hasClaim = true
					break
				}
			}
			if hasClaim {
				transformed[cardID] = claim.Claim
				extraEvent = EffectTransformToOppHand
			}
		case EffectTransformToLastClaim:
			if s.BlockingClaim == nil {
				lastClaim := s.Players[challenger].LastAttackingClaim
				if lastClaim != "" {
					transformed[cardID] = lastClaim
				}
			}
		case EffectTransformSelfWolfySci:
			transformed[cardID] = "scientist"
			// 同一牌堆中的狼爵士也变为科学家
			for _, otherID := range claim.CardIDs {
				if otherID != cardID && s.CardNames[otherID] == "wolfy" {
					transformed[otherID] = "scientist"
				}
			}
		}
	}

	allMatch := true
	for _, cardID := range claim.CardIDs {
		name := s.CardNames[cardID]
		actual := name
		if t, ok := transformed[cardID]; ok {
			actual = t
		}
		if actual != claim.Claim {
			allMatch = false
			break
		}
	}
	return claimResult{
		claimed:     claim.Claim,
		success:     !allMatch,
		ids:         append([]int{}, claim.CardIDs...),
		transformed: transformed,
		extraEvent:  extraEvent,
	}
}
