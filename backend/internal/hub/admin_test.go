package hub

import (
	"cakeduel-backend/internal/game"
	"context"
	"testing"
)

// TestAdminChallengeVerify 管理员一次性密码: 生成后只能从 KV(Redis)读取, 校验后删除
func TestAdminChallengeVerify(t *testing.T) {
	h := NewHub(game.GameConfig{RoundsToWin: 3}, nil)
	defer h.Close()

	if err := h.AdminCreateChallenge(); err != nil {
		t.Fatalf("生成密码失败: %v", err)
	}
	password, ok := h.kvGet(context.Background(), adminPasswordKey)
	if !ok || password == "" {
		t.Fatal("密码未写入 KV(Redis)")
	}

	// 错误密码应被拒绝
	if _, err := h.AdminVerify("wrong-password"); err == nil {
		t.Fatal("错误密码不应通过")
	}

	// 正确密码: 删除一次性密码并发放令牌
	token, err := h.AdminVerify(password)
	if err != nil {
		t.Fatalf("密码校验失败: %v", err)
	}
	if token == "" || !h.AdminAuthorize(token) {
		t.Fatal("校验成功后令牌无效")
	}
	if _, ok := h.kvGet(context.Background(), adminPasswordKey); ok {
		t.Fatal("一次性密码未在成功后删除")
	}
	// 已删除的密码不能再使用
	if _, err := h.AdminVerify(password); err == nil {
		t.Fatal("已删除的密码不应再次通过")
	}
}
