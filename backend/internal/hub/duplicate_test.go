package hub

import (
	"cakeduel-backend/internal/game"
	"encoding/json"
	"testing"
)

// TestDuplicateTokenJoinBlocked 同一 token 已有席位时, 新连接不能重复进入同一房间
func TestDuplicateTokenJoinBlocked(t *testing.T) {
	h := NewHub(game.GameConfig{RoundsToWin: 3}, nil)
	defer h.Close()

	c1 := newClient(h, nil)
	c1.Token = "same-token"
	h.CreateRoom(c1, &ClientMessage{Name: "甲", Mode: "private"})

	raw := <-c1.send
	var joined ServerMessage
	if err := json.Unmarshal(raw, &joined); err != nil {
		t.Fatalf("解析建房消息失败: %v", err)
	}
	if joined.Type != "room_joined" {
		t.Fatalf("建房应返回 room_joined, 实际 %s", joined.Type)
	}

	c2 := newClient(h, nil)
	c2.Token = "same-token"
	h.JoinRoom(c2, &ClientMessage{Name: "乙", Code: joined.RoomCode, As: "player"})
	raw2 := <-c2.send
	var denied ServerMessage
	if err := json.Unmarshal(raw2, &denied); err != nil {
		t.Fatalf("解析拒绝消息失败: %v", err)
	}
	if denied.Type != "error" || denied.Message == "" {
		t.Fatalf("同 token 重复加入应被拒绝, 实际 %s %s", denied.Type, denied.Message)
	}
}
