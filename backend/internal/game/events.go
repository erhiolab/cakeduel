package game

// RevealedCard 质疑翻开的一张牌
type RevealedCard struct {
	CardName      string `json:"cardName"`
	TransformedTo string `json:"transformedTo,omitempty"`
}

// Event 游戏事件
type Event struct {
	ID            int            `json:"id"`
	Type          string         `json:"type"`
	Player        *int           `json:"player,omitempty"`
	Phase         string         `json:"phase,omitempty"`
	Pile          string         `json:"pile,omitempty"`
	Claim         string         `json:"claim,omitempty"`
	CardNames     []string       `json:"cardNames,omitempty"`
	RevealedCards []RevealedCard `json:"revealedCards,omitempty"`
	Challenger    int            `json:"challenger,omitempty"`
	ClaimedCard   string         `json:"claimedCard,omitempty"`
	Success       bool           `json:"success"`
	From          int            `json:"from,omitempty"`
	To            int            `json:"to,omitempty"`
	Amount        int            `json:"amount,omitempty"`
	CakesAfter    [2]int         `json:"cakesAfter,omitempty"`
	Winner        int            `json:"winner,omitempty"`
	AttackerIndex int            `json:"attackerIndex,omitempty"`
	BoutNumber    int            `json:"boutNumber,omitempty"`
	PickType      string         `json:"pickType,omitempty"`
	Picks         []string       `json:"picks,omitempty"`
	Zone          string         `json:"zone,omitempty"`
	Config        *GameConfig    `json:"config,omitempty"`

	// 内部字段
	cardIDs     []int
	revealedIDs []int
	transformed map[int]string
	visibility  [2]bool
}

// Event types
const (
	EventGameStarted   = "game_started"
	EventBoutStarted   = "bout_started"
	EventPhaseChanged  = "phase_changed"
	EventClaimMade     = "claim_made"
	EventPassMade      = "pass_made"
	EventChallengeMade = "challenge_made"
	EventPickMade      = "pick_made"
	EventCakesTransfer = "cakes_transferred"
	EventCardDrawn     = "card_drawn"
	EventCardDiscarded = "card_discarded"
	EventDeckShuffled  = "deck_shuffled"
	EventHandRevealed  = "hand_revealed"
	EventWolfyTaunt    = "wolfy_taunt"
	EventBoutEnded     = "bout_ended"
	EventGameEnded     = "game_ended"
	EventConcedeMade   = "concede_made"
)

// EventWriter 事件收集器
type EventWriter struct {
	State  *State
	Events []Event
}

// push 添加事件
func (w *EventWriter) push(visibility [2]bool, evt Event) int {
	w.State.LastEventID++
	evt.ID = w.State.LastEventID
	evt.visibility = visibility
	w.Events = append(w.Events, evt)
	return evt.ID
}

// public 全员可见事件
func (w *EventWriter) public(evt Event) int {
	return w.push([2]bool{true, true}, evt)
}

// private 仅指定玩家可见事件
func (w *EventWriter) private(player int, evt Event) int {
	vis := [2]bool{false, false}
	vis[player] = true
	return w.push(vis, evt)
}

// cardNames 将卡牌实例ID转换为牌名
func (w *EventWriter) cardNames(cardIDs []int) []string {
	out := make([]string, 0, len(cardIDs))
	for _, id := range cardIDs {
		if id >= 0 && id < len(w.State.CardNames) {
			out = append(out, w.State.CardNames[id])
		} else {
			out = append(out, "unknown")
		}
	}
	return out
}

// filterForPlayer 按玩家过滤事件(可见性), 并将卡牌ID替换为牌名
func filterForPlayer(events []Event, player int, cardNames []string) []Event {
	out := make([]Event, 0, len(events))
	for _, evt := range events {
		if !evt.visibility[player] {
			continue
		}
		copied := evt
		copied.visibility = [2]bool{false, false}
		switch copied.Type {
		case EventClaimMade, EventCardDrawn, EventCardDiscarded, EventDeckShuffled, EventHandRevealed:
			copied.CardNames = nil
			for _, id := range evt.cardIDs {
				if id >= 0 && id < len(cardNames) {
					copied.CardNames = append(copied.CardNames, cardNames[id])
				} else {
					copied.CardNames = append(copied.CardNames, "unknown")
				}
			}
		case EventChallengeMade:
			copied.RevealedCards = nil
			for _, id := range evt.revealedIDs {
				name := "unknown"
				if id >= 0 && id < len(cardNames) {
					name = cardNames[id]
				}
				copied.RevealedCards = append(copied.RevealedCards, RevealedCard{
					CardName:      name,
					TransformedTo: evt.transformed[id],
				})
			}
		}
		out = append(out, copied)
	}
	return out
}

// FilterEvents 按玩家过滤事件(导出)
func FilterEvents(events []Event, player int, cardNames []string) []Event {
	return filterForPlayer(events, player, cardNames)
}

// RevealedIDs 质疑翻开的卡牌实例ID(导出)
func (e Event) RevealedIDs() []int {
	return e.revealedIDs
}

// Transformed 卡牌变换映射(导出)
func (e Event) Transformed() map[int]string {
	return e.transformed
}
