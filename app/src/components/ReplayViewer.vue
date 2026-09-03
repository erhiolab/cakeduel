<script setup lang="ts">
import {computed, onUnmounted, ref, watch} from "vue"
import type {CardEntity, PlayerInfo, ReplayData, Zones} from "../composables/useGame"
import SpectatorTable from "./SpectatorTable.vue"

// 自动播放步进间隔(ms)
const AUTO_STEP_MS = 1600

const props = defineProps<{
	replay: ReplayData
}>()

const emit = defineEmits<{
	close: []
}>()

// 当前帧索引
const current = ref(0)

// 是否自动播放
const playing = ref(false)

// 自动播放定时器
let playTimer: number | undefined

// 总帧数
const total = computed(() => props.replay.frames.length)

// 帧数兜底: 回放至少一帧
const safeTotal = computed(() => Math.max(1, total.value))

// 当前帧
const frame = computed(() => props.replay.frames[current.value] ?? props.replay.frames[0])

// 回放专用完整牌面(双方手牌/出牌堆都带真实牌名, 观战不可见)
const replayZones = computed<Zones | null>(() => {
	const F = frame.value
	if (!F || !F.zones) return null
	const NAMES = (list: string[] | undefined): CardEntity[] =>
		(list || []).map((name, i) => ({entityId: i, name}))
	return {
		...F.zones,
		playerHand: NAMES(F.playerHands?.[0]),
		opponentHand: NAMES(F.playerHands?.[1]),
		attackPile: NAMES(F.attackPile),
		blockPile: NAMES(F.blockPile),
		revealedPileCards: {...(F.zones.revealedPileCards ?? {})},
	}
})

// 是否显示聊天记录
const showChat = ref(false)

// 聊天记录开关
const toggleChat = () => {
	showChat.value = !showChat.value
}

// 玩家信息(供观战牌桌显示名字)
const playerInfos = computed<PlayerInfo[]>(() =>
	[0, 1].map((i) => ({
		index: i,
		name: props.replay.playerNames?.[i] || "",
		connected: true,
	})),
)

// 对局耗时文本
const durationText = computed(() => {
	const SECONDS = Math.max(1, Math.round((props.replay.durationMs ?? 0) / 1000))
	const M = Math.floor(SECONDS / 60)
	const S = SECONDS % 60
	return M > 0 ? `${M} 分 ${S} 秒` : `${S} 秒`
})

// 开始时间文本
const timeText = computed(() => {
	if (!props.replay.startedAt) return ""
	return new Date(props.replay.startedAt).toLocaleString("zh-CN", {
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
	})
})

// 步进到指定帧(越界自动夹紧)
const goTo = (index: number) => {
	stop()
	current.value = Math.max(0, Math.min(safeTotal.value - 1, index))
}

// 上一帧 / 下一帧
const prev = () => {
	goTo(current.value - 1)
}

const next = () => {
	if (current.value >= safeTotal.value - 1) return
	goTo(current.value + 1)
}

// 开始 / 暂停自动播放
const togglePlay = () => {
	if (total.value === 0) return
	if (playing.value) {
		stop()
		return
	}
	if (current.value >= safeTotal.value - 1) current.value = 0
	playing.value = true
	playTimer = window.setInterval(() => {
		if (current.value >= safeTotal.value - 1) {
			stop()
			return
		}
		current.value++
	}, AUTO_STEP_MS)
}

const stop = () => {
	playing.value = false
	if (playTimer) window.clearInterval(playTimer)
	playTimer = undefined
}

// 切换回放时从头播放
watch(() => props.replay,() => {
	stop()
	current.value = 0
})

// 视口尺寸
const viewport = ref({w: window.innerWidth, h: window.innerHeight})

const compact = computed(() => viewport.value.h < 700 || viewport.value.w < 960)

const onResize = () => {
	viewport.value = {w: window.innerWidth, h: window.innerHeight}
}

window.addEventListener("resize", onResize)

// 牌桌卡牌高度
const cardHeight = computed(() => {
	const AVAILABLE = viewport.value.h - (compact.value ? 168 : 236)
	const BASE = AVAILABLE * (compact.value ? 0.14 : 0.15)
	return Math.round(Math.max(compact.value ? 42 : 72, Math.min(compact.value ? 98 : 168, BASE)))
})

onUnmounted(() => {
	stop()
	window.removeEventListener("resize", onResize)
})
</script>

<template>
	<div class="replay-viewer" data-cakeduel-screen="replay">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="shade"></div>

		<header class="top-bar">
			<div class="bar-left">
				<button class="close-btn" title="关闭回放" @click="emit('close')">✕ 关闭</button>
				<span class="title">🎞 回放</span>
			</div>
			<div class="bar-center">
				<span class="vs">{{ replay.playerNames?.[0] || "玩家 A" }} vs {{ replay.playerNames?.[1] || "玩家 B" }}</span>
				<span class="meta">{{ timeText }} · 用时 {{ durationText }}</span>
			</div>
			<div class="bar-right">
				<span class="winner">🏆 {{ replay.playerNames?.[replay.winner] || `玩家 ${(replay.winner ?? 0) + 1}` }} 获胜</span>
			</div>
		</header>

		<main class="stage">
			<SpectatorTable
				:view="frame.view ?? null"
				:zones="replayZones"
				:players="playerInfos"
				:reveal="frame.reveal ?? null"
				:card-height="cardHeight"
				:compact="compact"
				:show-hands="true"
				:key="current"
			/>
			<div v-if="showChat" class="chat-panel glass">
				<div class="chat-head">
					<span>💬 对局聊天</span>
					<button class="chat-close" @click="showChat = false">✕</button>
				</div>
				<div class="chat-list">
					<div v-if="!replay.chats?.length" class="chat-empty">本局没有聊天记录</div>
					<div v-for="(m, i) in replay.chats ?? []" :key="`c${m.ts ?? i}`" class="chat-msg">
						<span class="chat-name">{{ m.name }}</span>
						<span class="chat-bubble">{{ m.text }}</span>
					</div>
				</div>
			</div>
		</main>

		<footer class="control-bar">
			<div class="controls">
				<button class="ctrl-btn" title="上一帧" @click="prev">⏮</button>
				<button class="ctrl-btn play" title="自动播放/暂停" @click="togglePlay">
					{{ playing ? "⏸" : "▶" }}
				</button>
				<button class="ctrl-btn" title="下一帧" @click="next">⏭</button>
				<button class="ctrl-btn" title="聊天记录" @click="toggleChat">💬</button>
			</div>
			<input
				class="timeline"
				type="range"
				min="0"
				:max="safeTotal - 1"
				:value="current"
				@input="goTo(Number(($event.target as HTMLInputElement).value))"
				@pointerdown="stop"
			/>
			<span class="frame-count">{{ current + 1 }} / {{ safeTotal }}</span>
		</footer>
	</div>
</template>

<style scoped>
.replay-viewer {
	position: fixed;
	inset: 0;
	z-index: 200;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.bg {
	position: absolute;
	inset: 0;
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.shade {
	position: absolute;
	inset: 0;
	background: radial-gradient(ellipse 70% 60% at 50% 45%, rgba(120, 140, 170, 0.12) 0%, transparent 70%);
	box-shadow: inset 0 0 6rem 1.5rem rgba(20, 28, 40, 0.5);
}

.top-bar {
	position: relative;
	z-index: 20;
	flex-shrink: 0;
	display: grid;
	grid-template-columns: 1fr auto 1fr;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	background: rgba(24, 20, 18, 0.88);
	backdrop-filter: blur(8px);
	border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.bar-left,
.bar-center,
.bar-right {
	display: flex;
	align-items: center;
	gap: 0.6rem;
	min-width: 0;
}

.bar-right {
	justify-content: flex-end;
}

.close-btn {
	display: flex;
	align-items: center;
	border-radius: 2rem;
	padding: 0.35rem 0.8rem;
	font-size: 0.75rem;
	font-weight: 700;
	color: rgba(255, 255, 255, 0.85);
	background: rgba(255, 255, 255, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.16);
	transition: background 0.2s;
}

.close-btn:hover {
	background: rgba(255, 255, 255, 0.22);
	color: #fff;
}

.title {
	font-size: 0.95rem;
	font-weight: 900;
	color: #ffd977;
	white-space: nowrap;
}

.vs {
	font-size: 0.85rem;
	font-weight: 800;
	color: #ffedd5;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	max-width: 14rem;
}

.meta,
.winner {
	font-size: 0.7rem;
	font-weight: 700;
	color: rgba(255, 235, 200, 0.7);
	white-space: nowrap;
}

.winner {
	color: #ffd977;
}

.stage {
	position: relative;
	z-index: 10;
	flex: 1;
	min-height: 0;
	padding: 0.5rem 1rem;
}

.chat-panel {
	position: absolute;
	right: 0.8rem;
	bottom: 0.6rem;
	z-index: 50;
	width: 19rem;
	max-width: calc(100vw - 1.5rem);
	max-height: 45vh;
	border-radius: 0.9rem;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.chat-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.5rem 0.75rem;
	font-size: 0.8rem;
	font-weight: 800;
	color: #3a2c1f;
	border-bottom: 1px solid rgba(107, 84, 56, 0.12);
}

.chat-close {
	width: 1.6rem;
	height: 1.6rem;
	border-radius: 50%;
	font-size: 0.8rem;
	color: #9a7a55;
	transition: background 0.2s;
}

.chat-close:hover {
	background: rgba(0, 0, 0, 0.08);
}

.chat-list {
	flex: 1;
	min-height: 5rem;
	max-height: calc(45vh - 2.5rem);
	overflow-y: auto;
	padding: 0.55rem 0.7rem;
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
}

.chat-empty {
	text-align: center;
	color: #a8947b;
	font-size: 0.75rem;
	padding: 1rem 0;
}

.chat-msg {
	display: flex;
	flex-direction: column;
	gap: 0.1rem;
	max-width: 95%;
}

.chat-name {
	font-size: 0.62rem;
	font-weight: 700;
	color: #a8947b;
}

.chat-bubble {
	font-size: 0.8rem;
	font-weight: 600;
	color: #3a2c1f;
	background: rgba(255, 255, 255, 0.65);
	border: 1px solid rgba(255, 255, 255, 0.7);
	border-radius: 0.6rem;
	padding: 0.35rem 0.6rem;
	word-break: break-word;
}

.control-bar {
	position: relative;
	z-index: 20;
	flex-shrink: 0;
	display: flex;
	align-items: center;
	gap: 0.8rem;
	padding: 0.55rem 1rem;
	background: rgba(24, 20, 18, 0.88);
	backdrop-filter: blur(8px);
	border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.controls {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-shrink: 0;
}

.ctrl-btn {
	width: 2.2rem;
	height: 2.2rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.95rem;
	color: #ffedd5;
	background: rgba(255, 255, 255, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.16);
	transition: background 0.2s, transform 0.15s;
}

.ctrl-btn:hover {
	background: rgba(255, 255, 255, 0.22);
	transform: scale(1.05);
}

.ctrl-btn.play {
	width: 2.6rem;
	height: 2.6rem;
	font-size: 1.1rem;
	background: linear-gradient(135deg, #d97706, #92400e);
	border-color: rgba(255, 200, 120, 0.5);
}

.timeline {
	flex: 1;
	min-width: 0;
	accent-color: #f5c54a;
	height: 0.35rem;
	cursor: pointer;
}

.frame-count {
	flex-shrink: 0;
	font-size: 0.75rem;
	font-weight: 800;
	color: rgba(255, 235, 200, 0.8);
	min-width: 3.4rem;
	text-align: right;
}

/* 小屏: 紧凑展示 */
@media (max-width: 700px) {
	.top-bar {
		grid-template-columns: 1fr auto;
		padding: 0.3rem 0.6rem;
		gap: 0.35rem;
	}

	.bar-center {
		grid-column: 1 / -1;
		grid-row: 2;
		justify-content: center;
	}

	.title {
		font-size: 0.8rem;
	}

	.close-btn {
		padding: 0.2rem 0.6rem;
		font-size: 0.68rem;
	}

	.vs {
		font-size: 0.75rem;
	}

	.meta,
	.winner {
		font-size: 0.62rem;
	}

	.stage {
		padding: 0.25rem 0.4rem;
	}

	.control-bar {
		padding: 0.35rem 0.5rem;
		gap: 0.5rem;
	}

	.ctrl-btn {
		width: 1.9rem;
		height: 1.9rem;
		font-size: 0.85rem;
	}

	.ctrl-btn.play {
		width: 2.3rem;
		height: 2.3rem;
	}

	.frame-count {
		font-size: 0.65rem;
		min-width: 3rem;
	}
}
</style>
