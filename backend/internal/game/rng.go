package game

// RNG xoshiro128** 随机数生成器
type RNG struct {
	s0, s1, s2, s3 uint32
}

// NewRNG 根据种子创建随机数生成器
func NewRNG(seed uint32) *RNG {
	n := seed
	next := func() uint32 {
		n = n*1664525 + 1013904223
		return n
	}
	return &RNG{
		s0: next() | 1,
		s1: next() | 1,
		s2: next() | 1,
		s3: next() | 1,
	}
}

// rotl 循环左移
func rotl(x uint32, k uint) uint32 {
	return (x << k) | (x >> (32 - k))
}

// next 下一个 [0,1) 随机数
func (r *RNG) next() float64 {
	result := rotl(r.s1*5, 7) * 9
	t := r.s1 << 9
	r.s2 ^= r.s0
	r.s3 ^= r.s1
	r.s1 ^= r.s2
	r.s0 ^= r.s3
	r.s2 ^= t
	r.s3 = rotl(r.s3, 11)
	return float64(result) / 4294967296
}

// NextInt 返回 [0,n) 的随机整数
func (r *RNG) NextInt(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() * float64(n))
}

// Shuffle 原地洗牌并返回
func (r *RNG) Shuffle(items []int) []int {
	for i := len(items) - 1; i > 0; i-- {
		j := r.NextInt(i + 1)
		items[i], items[j] = items[j], items[i]
	}
	return items
}

// Pick 随机取一个元素
func (r *RNG) Pick(items []int) int {
	if len(items) == 0 {
		return -1
	}
	return items[r.NextInt(len(items))]
}

// Sample 随机取 n 个不重复元素
func (r *RNG) Sample(items []int, n int) []int {
	if n < 0 {
		n = 0
	}
	if n > len(items) {
		n = len(items)
	}
	if n == 0 {
		return []int{}
	}
	if n == len(items) {
		out := make([]int, len(items))
		copy(out, items)
		return r.Shuffle(out)
	}
	// 拒绝采样
	selected := make(map[int]bool)
	for len(selected) < n {
		idx := r.NextInt(len(items))
		selected[idx] = true
	}
	out := make([]int, 0, n)
	for idx := range selected {
		out = append(out, items[idx])
	}
	return out
}
