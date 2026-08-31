package game

import "testing"

// TestAngelNoHandReveal 天使防守结算不应产生手牌查看事件
func TestAngelNoHandReveal(t *testing.T) {
	s := newManualState()
	// B(防守方) 用天使挡 A 的法师攻击
	s.AttackingClaim = &Claim{Claim: "wizard", CardIDs: []int{cardIndex(s.CardNames, "wizard", 0)}}
	s.BlockingClaim = &Claim{Claim: "angel", CardIDs: []int{cardIndex(s.CardNames, "angel", 0)}}
	s.Phase = PhaseReview
	s.Players[0].Hand = []int{cardIndex(s.CardNames, "soldier", 0), cardIndex(s.CardNames, "archer", 0)}
	s.Players[1].Hand = []int{cardIndex(s.CardNames, "defender", 0), cardIndex(s.CardNames, "scientist", 0)}

	events, err := Step(s, 0, &Action{Type: ActionPass})
	if err != nil {
		t.Fatal(err)
	}
	for _, evt := range events {
		if evt.Type == EventHandRevealed {
			t.Fatalf("天使结算不应触发手牌查看事件: %+v", evt)
		}
		if evt.Type == EventClaimMade && evt.Pile == "block_pile" {
			// 防守声明只对防守方可见, 攻击方不应看到牌名
		}
	}
	// 天使: 攻击方(0)继续行动
	if s.AttackerIndex != 0 {
		t.Fatalf("天使应让攻击方再行动, 得到 %d", s.AttackerIndex)
	}
}

// TestBlockClaimVisibility 防守声明事件只对防守方可见
func TestBlockClaimVisibility(t *testing.T) {
	s := newManualState()
	s.AttackingClaim = &Claim{Claim: "wizard", CardIDs: []int{cardIndex(s.CardNames, "wizard", 0)}}
	s.Players[0].Hand = []int{cardIndex(s.CardNames, "soldier", 0)}
	s.Players[1].Hand = []int{cardIndex(s.CardNames, "angel", 0)}

	events, err := Step(s, 1, &Action{Type: ActionClaim, HandIndices: []int{0}, Claim: "angel"})
	if err != nil {
		t.Fatal(err)
	}
	var claimEvent *Event
	for i := range events {
		if events[i].Type == EventClaimMade {
			claimEvent = &events[i]
			break
		}
	}
	if claimEvent == nil {
		t.Fatal("缺少 claim_made 事件")
	}
	// 防守方(1)能看到牌名, 攻击方(0)不能
	filtered0 := FilterEvents([]Event{*claimEvent}, 0, s.CardNames)
	filtered1 := FilterEvents([]Event{*claimEvent}, 1, s.CardNames)
	if len(filtered0) != 0 {
		t.Fatalf("攻击方不应看到防守声明事件: %+v", filtered0)
	}
	if len(filtered1) != 1 || len(filtered1[0].CardNames) == 0 {
		t.Fatalf("防守方应看到自己的声明: %+v", filtered1)
	}
}
