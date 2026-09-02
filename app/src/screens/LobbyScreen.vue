<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue"
import {game} from "../composables/useGame"
import {playSfx} from "../game/audio"
import {CARDS, SPECIAL_CARD_NAMES} from "../game/cards"

// 房间状态
const {state, startGame, leave} = game

// 玩家状态
const players = computed(() => state.players)

// 是否可以开始游戏
const canStart = computed(() => state.players.filter((p) => p.connected).length === 2)

// 是否复制房间码
const copied = ref(false)

// 视口尺寸(小屏切换为横向布局)
const viewport = ref({w: window.innerWidth, h: window.innerHeight})

// 视口变化时更新
const onResize = () => {
	viewport.value = {w: window.innerWidth, h: window.innerHeight}
}

// 高度不够时使用横向布局: 左侧房间号/右侧玩家与开始按钮
const compact = computed(() => viewport.value.h < 620)

// 标题: 随机匹配成功 vs 普通房间
const title = computed(() => (state.mode === "random" ? "匹配成功" : "房间"))

// 复制房间码定时器
let copyTimer: number | undefined

// 复制房间码
const copyCode = () => {
	playSfx("hoof")
	navigator.clipboard?.writeText(state.roomCode).catch(() => {})
	copied.value = true
	if (copyTimer) window.clearTimeout(copyTimer)
	copyTimer = window.setTimeout(() => (copied.value = false), 1600)
}

// 卡组配置摘要(默认经典 / 自定义列表 / 无特殊卡)
const deckText = computed(() => {
	const CFG = state.deckConfig
	if (!CFG) return "卡组：经典（基础 29 + 特殊 11）"
	const ITEMS = SPECIAL_CARD_NAMES.filter((n) => (CFG[n] ?? 0) > 0)
	if (ITEMS.length === 0) return "卡组：仅基础卡（不使用特殊卡）"
	return `卡组：${ITEMS.map((n) => `${CARDS[n]?.name ?? n}×${CFG[n]}`).join(" ")}`
})

onMounted(() => {
	window.addEventListener("resize", onResize)
})

onUnmounted(() => {
	window.removeEventListener("resize", onResize)
})
</script>

<template>
	<div class="lobby" data-cakeduel-screen="lobby" :class="{ compact }">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="overlay"></div>
		<div class="content">
			<div class="card glass" :class="{ compact }">
				<div class="card-head">
					<h1 class="title-font">{{ title }}</h1>
					<p v-if="state.mode === 'random'" class="match-line">双方已匹配，准备开战</p>
				</div>
				<div class="split">
					<section class="room-col">
						<div class="code-box" @click="copyCode">
							<span class="code">{{ state.roomCode }}</span>
							<span class="copy-hint">{{ copied ? "已复制！" : "点击复制" }}</span>
						</div>
						<p class="deck-info">{{ deckText }}</p>
					</section>
					<section class="battle-col">
						<div class="players">
							<div v-for="i in [0, 1]" :key="i" class="player-row" :class="{ empty: !players[i] }">
								<div class="seat">{{ i === 0 ? "攻" : "守" }}</div>
								<span class="pname">{{ players[i]?.name || "等待玩家…" }}</span>
								<span v-if="i === state.playerIndex" class="me">我</span>
								<span v-else-if="players[i] && !players[i].connected" class="me offline">已掉线</span>
								<span v-else-if="!players[i]" class="loading">…</span>
							</div>
						</div>
						<button class="start-btn" :disabled="!canStart" @click="startGame">
							{{ canStart ? "开始决斗" : "等待对手加入…" }}
						</button>
						<button class="leave-btn" @click="leave">离开房间</button>
					</section>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.lobby {
	position: relative;
	width: 100%;
	height: 100%;
	overflow: hidden;
	display: flex;
	align-items: center;
	justify-content: center;
}

.bg {
	position: absolute;
	inset: 0;
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.overlay {
	position: absolute;
	inset: 0;
	background: rgba(30, 45, 60, 0.45);
	box-shadow: inset 0 0 6rem 1.5rem rgba(30, 45, 60, 0.6);
}

.content {
	position: relative;
	z-index: 10;
	width: 100%;
	max-width: 24rem;
	padding: 1rem;
}

.card {
	border-radius: 1.1rem;
	padding: 1.6rem 1.4rem;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 1rem;
	animation: rise-in 0.4s ease both;
}

.card h1 {
	font-size: 1.5rem;
	color: #3a2c1f;
}

.card-head {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.15rem;
}

.match-line {
	font-size: 0.78rem;
	font-weight: 700;
	color: #b45309;
	animation: pulse-glow 1.6s ease-in-out infinite;
}

.split {
	width: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 1rem;
}

.room-col,
.battle-col {
	width: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.6rem;
}

.code-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.2rem;
	cursor: pointer;
	padding: 0.6rem 1.6rem;
	border-radius: 0.9rem;
	background: rgba(255, 255, 255, 0.5);
	border: 2px dashed rgba(107, 84, 56, 0.35);
	transition: background 0.2s, border-color 0.2s;
}

.code-box:hover {
	background: rgba(255, 255, 255, 0.7);
	border-color: rgba(245, 197, 24, 0.8);
}

.code {
	font-size: 1.8rem;
	font-weight: 900;
	letter-spacing: 0.25em;
	color: #3a2c1f;
}

.copy-hint {
	font-size: 0.7rem;
	color: #9a7a55;
}

.deck-info {
	width: 100%;
	text-align: center;
	font-size: 0.78rem;
	font-weight: 700;
	color: #6b4a2b;
	line-height: 1.4;
}

.players {
	width: 100%;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.player-row {
	display: flex;
	align-items: center;
	gap: 0.7rem;
	padding: 0.65rem 0.8rem;
	border-radius: 0.8rem;
	background: rgba(255, 255, 255, 0.45);
	border: 1px solid rgba(255, 255, 255, 0.6);
	font-weight: 700;
	font-size: 0.95rem;
	color: #3a2c1f;
}

.player-row.empty {
	color: #9a7a55;
	font-weight: 600;
}

.pname {
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.seat {
	width: 1.8rem;
	height: 1.8rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.75rem;
	font-weight: 900;
	color: #fff;
	background: linear-gradient(135deg, #c97050, #a85a3a);
}

.player-row:first-child .seat {
	background: linear-gradient(135deg, #d97706, #92400e);
}

.me {
	margin-left: auto;
	font-size: 0.7rem;
	color: #b45309;
	font-weight: 800;
}

.me.offline {
	color: #dc2626;
}

.loading {
	margin-left: auto;
	animation: blink 1.2s ease-in-out infinite;
}

@keyframes blink {
	0%,
	100% {
		opacity: 0.3;
	}
	50% {
		opacity: 1;
	}
}

.start-btn {
	width: 100%;
	border-radius: 0.85rem;
	padding: 0.9rem;
	background: linear-gradient(135deg, #7ab55c, #5a9a40 50%, #4a5a40);
	color: #fdf6e9;
	font-weight: 800;
	font-size: 1.05rem;
	box-shadow: 0 6px 24px rgba(90, 154, 64, 0.35);
	transition: transform 0.15s, opacity 0.2s;
}

.start-btn:hover:not(:disabled) {
	transform: scale(1.02);
}

.start-btn:disabled {
	opacity: 0.55;
}

.leave-btn {
	width: 100%;
	padding: 0.55rem;
	border-radius: 0.7rem;
	font-size: 0.85rem;
	font-weight: 700;
	color: #6b5438;
	transition: background 0.2s;
}

.leave-btn:hover {
	background: rgba(0, 0, 0, 0.05);
}

/* 小屏/矮窗口: 横向布局, 左侧房间号, 右侧玩家与开始决斗 */
.lobby.compact .content {
	max-width: 46rem;
	padding: 0.45rem 0.75rem;
}

.card.compact {
	padding: 0.7rem 1.1rem;
	gap: 0.5rem;
}

.card.compact .card-head {
	gap: 0.05rem;
}

.card.compact h1 {
	font-size: 1.25rem;
}

.card.compact .match-line {
	font-size: 0.68rem;
}

.card.compact .split {
	flex-direction: row;
	align-items: stretch;
	gap: 0.9rem;
}

.card.compact .room-col {
	width: 40%;
	min-width: 0;
	justify-content: center;
	gap: 0.35rem;
	padding-right: 0.9rem;
	border-right: 1px dashed rgba(107, 84, 56, 0.3);
}

.card.compact .battle-col {
	flex: 1;
	min-width: 0;
	justify-content: center;
	gap: 0.4rem;
}

.card.compact .code-box {
	padding: 0.35rem 1rem;
	border-radius: 0.7rem;
	gap: 0.08rem;
}

.card.compact .code {
	font-size: 1.45rem;
}

.card.compact .copy-hint {
	font-size: 0.62rem;
}

.card.compact .deck-info {
	font-size: 0.62rem;
	line-height: 1.3;
}

.card.compact .players {
	gap: 0.35rem;
}

.card.compact .player-row {
	padding: 0.35rem 0.55rem;
	gap: 0.5rem;
	border-radius: 0.55rem;
	font-size: 0.8rem;
}

.card.compact .seat {
	width: 1.5rem;
	height: 1.5rem;
	font-size: 0.65rem;
}

.card.compact .me {
	font-size: 0.62rem;
}

.card.compact .start-btn {
	padding: 0.55rem;
	border-radius: 0.6rem;
	font-size: 0.9rem;
}

.card.compact .leave-btn {
	padding: 0.3rem;
	border-radius: 0.55rem;
	font-size: 0.75rem;
}

/* 极矮窗口(如 568×320)进一步压缩 */
@media (max-height: 380px) {
	.card.compact {
		padding: 0.45rem 0.8rem;
		gap: 0.3rem;
	}

	.card.compact h1 {
		font-size: 1.05rem;
	}

	.card.compact .match-line {
		font-size: 0.6rem;
	}

	.card.compact .split {
		gap: 0.6rem;
	}

	.card.compact .room-col {
		padding-right: 0.6rem;
		gap: 0.2rem;
	}

	.card.compact .code {
		font-size: 1.2rem;
	}

	.card.compact .deck-info {
		font-size: 0.55rem;
	}

	.card.compact .player-row {
		padding: 0.2rem 0.4rem;
		font-size: 0.72rem;
	}

	.card.compact .seat {
		width: 1.25rem;
		height: 1.25rem;
		font-size: 0.58rem;
	}

	.card.compact .start-btn {
		padding: 0.4rem;
		font-size: 0.82rem;
	}

	.card.compact .leave-btn {
		padding: 0.15rem;
		font-size: 0.68rem;
	}
}
</style>
