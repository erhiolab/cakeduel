package hub

import (
	"context"
	"time"
)

// kvItem 内存兜底 KV 的条目(Redis 不可用时使用)
type kvItem struct {
	data    []byte
	expires time.Time
}

// kvSet 写入 KV(Redis 可用走 Redis, 否则内存兜底, 懒过期)
func (h *Hub) kvSet(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Minute
	}
	if h.rdb != nil {
		return h.rdb.Set(ctx, key, value, ttl).Err()
	}
	h.kvMu.Lock()
	defer h.kvMu.Unlock()
	h.kvMem[key] = kvItem{
		data:    []byte(value),
		expires: time.Now().Add(ttl),
	}
	return nil
}

// kvGet 读取 KV(不存在/已过期返回空)
func (h *Hub) kvGet(ctx context.Context, key string) (string, bool) {
	if h.rdb != nil {
		val, err := h.rdb.Get(ctx, key).Result()
		if err != nil {
			return "", false
		}
		return val, true
	}
	h.kvMu.Lock()
	defer h.kvMu.Unlock()
	item, ok := h.kvMem[key]
	if !ok {
		return "", false
	}
	if time.Now().After(item.expires) {
		delete(h.kvMem, key)
		return "", false
	}
	return string(item.data), true
}

// kvDel 删除 KV
func (h *Hub) kvDel(ctx context.Context, key string) {
	if h.rdb != nil {
		_ = h.rdb.Del(ctx, key).Err()
		return
	}
	h.kvMu.Lock()
	defer h.kvMu.Unlock()
	delete(h.kvMem, key)
}

// kvKeys 列出指定前缀的 key
func (h *Hub) kvKeys(ctx context.Context, prefix string) []string {
	if h.rdb != nil {
		keys, err := h.rdb.Keys(ctx, prefix+"*").Result()
		if err == nil {
			return keys
		}
		return nil
	}
	h.kvMu.Lock()
	defer h.kvMu.Unlock()
	now := time.Now()
	out := make([]string, 0, len(h.kvMem))
	for key, item := range h.kvMem {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			if !now.After(item.expires) {
				out = append(out, key)
			} else {
				delete(h.kvMem, key)
			}
		}
	}
	return out
}
