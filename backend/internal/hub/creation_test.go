package hub

import (
	"cakeduel-backend/internal/game"
	"encoding/json"
	"testing"
)

// TestCreationSwitch 创建房间开关: 关闭后拒绝建房, 重新开启后恢复(重启默认开启)
func TestCreationSwitch(t *testing.T) {
	h := NewHub(game.GameConfig{RoundsToWin: 3}, nil)
	defer h.Close()

	if !h.CreationEnabled() {
		t.Fatal("默认应允许创建房间")
	}
	h.SetCreationEnabled(false)
	if h.CreationEnabled() {
		t.Fatal("关闭失败")
	}

	c := newClient(h, nil)
	h.CreateRoom(c, &ClientMessage{Name: "测试", Mode: "private"})
	select {
	case raw := <-c.send:
		var msg ServerMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("解析消息失败: %v", err)
		}
		if msg.Type != "error" {
			t.Fatalf("关闭创建后应返回 error, 实际 %s", msg.Type)
		}
		if msg.Message == "" {
			t.Fatal("缺少拒绝原因")
		}
	default:
		t.Fatal("关闭创建后客户端没有收到任何消息")
	}

	h.SetCreationEnabled(true)
	if !h.CreationEnabled() {
		t.Fatal("重新开启失败")
	}
}
