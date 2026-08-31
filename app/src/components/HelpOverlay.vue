<script setup lang="ts">
import {onMounted, onUnmounted} from "vue"
import {CARD_KEYS, CARDS, cardSmallImage} from "../game/cards"

// 描述开头的类型 emoji 前缀(带可选变体选择符)
const TYPE_PREFIX_RE = /^[⚔🔥🛡🐑⭐]\uFE0F?\s*·\s*/

defineProps<{
	open: boolean
}>()

const emit = defineEmits<{
	close: []
}>()

// 基础卡列表
const baseCards = CARD_KEYS.filter((k) => ["soldier", "archer", "wizard", "defender", "scientist", "wolfy"].includes(k))

// 特殊卡列表
const specialCards = CARD_KEYS.filter((k) => !baseCards.includes(k))

// 基础卡数量
const baseCounts: Record<string, number> = {soldier: 7, archer: 6, wizard: 5, defender: 6, scientist: 4, wolfy: 1}

// 特殊卡数量
const specialCounts: Record<string, number> = {
	assassin: 1,
	scout: 1,
	summoner: 1,
	quartermaster: 1,
	oracle: 1,
	priest: 2,
	angel: 1,
	baacrates: 1,
	agent_u: 1,
	pierrot: 1,
}

// 精简描述: 去掉类型前缀, 只保留第一句
const shortDesc = (key: string): string => {
	const DESC = CARDS[key]?.desc ?? ""
	const CLEANED = DESC.replace(TYPE_PREFIX_RE, "")
	const DOT = CLEANED.indexOf("。")
	return DOT >= 0 ? CLEANED.slice(0, DOT + 1) : CLEANED
}

// 按 Esc 关闭
const onKeydown = (e: KeyboardEvent) => {
	if (e.key === "Escape") emit("close")
}

// 取卡牌主题色
const colorOf = (key: string) => {
	return CARDS[key]?.color ?? "#B7C0C3"
}

// 十六进制颜色转 rgba
const rgba = (hex: string, a: number) => {
	const R = parseInt(hex.slice(1, 3), 16)
	const G = parseInt(hex.slice(3, 5), 16)
	const B = parseInt(hex.slice(5, 7), 16)
	return `rgba(${R},${G},${B},${a})`
}

onMounted(() => {
	window.addEventListener("keydown", onKeydown)
})

onUnmounted(() => {
	window.removeEventListener("keydown", onKeydown)
})
</script>

<template>
	<div v-if="open" class="overlay" @click.self="emit('close')">
		<div class="sheet">
			<div class="head">
				<h2 class="title-font">玩法</h2>
				<button class="close" aria-label="关闭" @click="emit('close')">✕</button>
			</div>
			<div class="body">
				<section>
					<h3>目标</h3>
					<p>先赢 3 局获胜；每局进攻方 3 个蛋糕、防守方 4 个，各 4 张手牌。</p>
				</section>
				<section>
					<h3>攻防</h3>
					<p>暗牌出牌并声明牌名（可以说谎）。防守方可暗牌克制、放行或质疑。</p>
				</section>
				<section>
					<h3>质疑</h3>
					<p>翻牌核对：有假则说谎者输本局，全真则质疑者输本局。</p>
				</section>
				<section>
					<h3>一局如何结束</h3>
					<p>蛋糕归零 / 质疑出结果 / 连续过牌（蛋糕多者胜）。每轮攻防后补满手牌。</p>
				</section>

				<div class="deck-title">
					<span>基础卡</span>
					<span class="deck-sub">手牌上限 4</span>
				</div>
				<div class="deck-list">
					<div v-for="key in baseCards" :key="key" class="card-row">
						<img :src="cardSmallImage(key)" :alt="CARDS[key].name" :style="{ filter: `drop-shadow(0 1px 2px ${rgba(colorOf(key), 0.45)})` }"/>
						<span class="name">{{ CARDS[key].name }}</span>
						<span class="count" :style="{ background: rgba(colorOf(key), 0.18) }">×{{ baseCounts[key] }}</span>
						<span class="desc">{{ shortDesc(key) }}</span>
					</div>
				</div>
				<div class="deck-title">
					<span>特殊卡</span>
					<span class="deck-sub">每局全部加入</span>
				</div>
				<div class="deck-list">
					<div v-for="key in specialCards" :key="key" class="card-row">
						<img :src="cardSmallImage(key)" :alt="CARDS[key].name" :style="{ filter: `drop-shadow(0 1px 2px ${rgba(colorOf(key), 0.45)})` }"/>
						<span class="name">{{ CARDS[key].name }}</span>
						<span class="count" :style="{ background: rgba(colorOf(key), 0.18) }">×{{ specialCounts[key] }}</span>
						<span class="desc">{{ shortDesc(key) }}</span>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.overlay {
	position: fixed;
	inset: 0;
	z-index: 90;
	display: flex;
	align-items: safe center;
	justify-content: center;
	padding: 1rem;
	background: rgba(48, 34, 22, 0.45);
	backdrop-filter: blur(10px) saturate(120%);
}

.sheet {
	width: 100%;
	max-width: 32rem;
	max-height: 76vh;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
	border-radius: 1rem;
	background: linear-gradient(180deg, #fdf6e9, #f3e4cd);
	border: 1px solid rgba(255, 255, 255, 0.75);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9), 0 18px 48px rgba(40, 24, 10, 0.45);
}

.head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem 1.4rem 0.5rem;
	position: sticky;
	top: 0;
	z-index: 2;
	background: linear-gradient(180deg, #fdf6e9 82%, rgba(253, 246, 233, 0));
}

.head h2 {
	font-size: 1.5rem;
	color: #6b4a2b;
}

.close {
	width: 2.4rem;
	height: 2.4rem;
	border-radius: 50%;
	font-size: 1rem;
	color: #9a7a55;
	background: rgba(0, 0, 0, 0.05);
	flex-shrink: 0;
	transition: background 0.2s;
}

.close:hover {
	background: rgba(0, 0, 0, 0.1);
}

.body {
	padding: 0 1.4rem 1.4rem;
	overflow-y: auto;
	flex: 1;
}

section {
	margin-bottom: 0.55rem;
}

section h3 {
	font-size: 0.68rem;
	font-weight: 900;
	letter-spacing: 0.12em;
	text-transform: uppercase;
	color: #c4756a;
	margin-bottom: 0.1rem;
}

section p {
	font-size: 0.8rem;
	line-height: 1.45;
	font-weight: 600;
	color: #6b4a2b;
}

.deck-title {
	display: flex;
	align-items: baseline;
	justify-content: space-between;
	margin: 0.8rem 0 0.4rem;
	font-size: 0.7rem;
	font-weight: 900;
	letter-spacing: 0.12em;
	text-transform: uppercase;
	color: #c4756a;
}

.deck-sub {
	font-size: 0.66rem;
	font-weight: 700;
	color: #9a7a55;
}

.deck-list {
	border-radius: 0.8rem;
	overflow: hidden;
	background: rgba(255, 255, 255, 0.45);
	border: 1px solid rgba(196, 117, 106, 0.15);
}

.card-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.28rem 0.6rem;
	border-bottom: 1px solid rgba(154, 122, 85, 0.12);
}

.card-row:last-child {
	border-bottom: none;
}

.card-row img {
	width: 1.7rem;
	height: 1.7rem;
	object-fit: contain;
	flex-shrink: 0;
}

.name {
	width: 4rem;
	flex-shrink: 0;
	font-size: 0.76rem;
	font-weight: 900;
	color: #6b4a2b;
}

.count {
	flex-shrink: 0;
	border-radius: 1rem;
	padding: 0.02rem 0.45rem;
	font-size: 0.66rem;
	font-weight: 900;
	color: #6b4a2b;
}

.desc {
	min-width: 0;
	flex: 1;
	font-size: 0.7rem;
	font-weight: 600;
	line-height: 1.35;
	color: #9a7a55;
}
</style>
