<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue"
import {game} from "../composables/useGame"
import SpectatorTable from "../components/SpectatorTable.vue"
import {audioState, toggleBgm} from "../game/audio"

// 游戏状态
const {state, leave} = game

// 视口尺寸
const viewport = ref({w: window.innerWidth, h: window.innerHeight})

// 视口变化时更新
const onResize = () => {
	viewport.value = {w: window.innerWidth, h: window.innerHeight}
}

// 是否紧凑布局
const compact = computed(() => viewport.value.h < 700 || viewport.value.w < 960)

// 牌桌卡牌高度(留出顶部栏/内边距后自适应, 不顶爆屏幕)
const cardHeight = computed(() => {
	const AVAILABLE = viewport.value.h - (compact.value ? 76 : 112)
	const BASE = AVAILABLE * (compact.value ? 0.15 : 0.16)
	return Math.round(Math.max(compact.value ? 46 : 76, Math.min(compact.value ? 104 : 175, BASE)))
})

onMounted(() => {
	window.addEventListener("resize", onResize)
})

onUnmounted(() => {
	window.removeEventListener("resize", onResize)
})
</script>

<template>
	<div class="spectate" data-cakeduel-screen="spectate" :class="{ compact }">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="shade"></div>

		<header class="top-bar">
			<div class="bar-left">
				<button class="music-btn" :class="{ muted: !audioState.bgmOn }" title="背景音乐开关" @click="toggleBgm">
					<svg v-if="audioState.bgmOn" viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
						<path d="M9 18V5l12-2v13"/>
						<circle cx="6" cy="18" r="3"/>
						<circle cx="18" cy="16" r="3"/>
					</svg>
					<svg v-else viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor"
						 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d="M9 18V5l12-2v13"/>
						<circle cx="6" cy="18" r="3"/>
						<path d="M16 8l5 5M21 8l-5 5"/>
					</svg>
				</button>
				<span class="title">👁 观战中</span>
			</div>
			<div class="bar-center">
				<span v-if="state.roomCode" class="room">房间 {{ state.roomCode }}</span>
				<span v-if="state.spectatorView?.gameEnded" class="ended-tag">对局结束</span>
			</div>
			<div class="bar-right">
				<button class="leave-btn" @click="leave">离开观战</button>
			</div>
		</header>

		<main class="stage">
			<SpectatorTable
				:view="state.spectatorView"
				:zones="state.zones"
				:players="state.players"
				:reveal="state.reveal"
				:card-height="cardHeight"
				:compact="compact"
			/>
			<Transition name="pop">
				<div v-if="state.toast" class="toast glass">{{ state.toast }}</div>
			</Transition>
		</main>
	</div>
</template>

<style scoped>
.spectate {
	position: relative;
	width: 100%;
	height: 100%;
	overflow: hidden;
	display: flex;
	flex-direction: column;
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
	padding: 0.45rem 1rem;
	background: rgba(24, 20, 18, 0.82);
	backdrop-filter: blur(8px);
	border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.bar-left,
.bar-right {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	min-width: 0;
}

.bar-right {
	justify-content: flex-end;
}

.music-btn,
.leave-btn {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	border-radius: 2rem;
	padding: 0.35rem 0.8rem;
	font-size: 0.75rem;
	font-weight: 700;
	color: rgba(255, 255, 255, 0.85);
	background: rgba(255, 255, 255, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.16);
	transition: background 0.2s, color 0.2s;
}

.music-btn {
	padding: 0.35rem;
	width: 2rem;
	justify-content: center;
}

.music-btn.muted {
	color: rgba(255, 255, 255, 0.35);
}

.music-btn:hover,
.leave-btn:hover {
	background: rgba(255, 255, 255, 0.2);
	color: #fff;
}

.title {
	font-size: 0.95rem;
	font-weight: 900;
	color: #ffd977;
	letter-spacing: 0.05rem;
	white-space: nowrap;
}

.bar-center {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.room {
	padding: 0.3rem 0.9rem;
	border-radius: 2rem;
	font-size: 0.8rem;
	font-weight: 800;
	letter-spacing: 0.12rem;
	color: #ffedd5;
	background: rgba(0, 0, 0, 0.3);
	border: 1px solid rgba(255, 255, 255, 0.15);
}

.ended-tag {
	padding: 0.25rem 0.7rem;
	border-radius: 2rem;
	font-size: 0.72rem;
	font-weight: 800;
	color: #3a2c1f;
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	animation: pulse-glow 1.6s ease-in-out infinite;
}

.stage {
	position: relative;
	z-index: 10;
	flex: 1;
	min-height: 0;
	padding: 0.4rem 0.8rem 0.6rem;
}

.toast {
	position: absolute;
	left: 50%;
	bottom: 0.8rem;
	transform: translateX(-50%);
	z-index: 40;
	border-radius: 0.8rem;
	padding: 0.5rem 1rem;
	font-size: 0.8rem;
	font-weight: 700;
	color: #44403c;
}

/* 小屏: 顶部栏更紧凑, 为牌桌留空间 */
.spectate.compact .top-bar {
	padding: 0.25rem 0.6rem;
}

.spectate.compact .title {
	font-size: 0.8rem;
}

.spectate.compact .music-btn {
	width: 1.7rem;
	height: 1.7rem;
	padding: 0;
}

.spectate.compact .leave-btn {
	padding: 0.25rem 0.6rem;
	font-size: 0.68rem;
}

.spectate.compact .room {
	font-size: 0.68rem;
	padding: 0.15rem 0.6rem;
}

.spectate.compact .stage {
	padding: 0.2rem 0.35rem 0.3rem;
}
</style>
