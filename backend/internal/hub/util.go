package hub

import (
	"cakeduel-backend/internal/game"
	"math/rand"
	"strings"
)

// randomRoomCode 生成房间码
func randomRoomCode(n int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// sanitizeName 清理玩家名
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "神秘玩家"
	}
	runes := []rune(name)
	if len(runes) > 12 {
		runes = runes[:12]
	}
	return strings.TrimSpace(string(runes))
}

// sanitizeCode 清理房间码
func sanitizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

var _ = game.Event{}
