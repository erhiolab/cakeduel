<script setup lang="ts">
import {computed, ref} from "vue"
import {game} from "../composables/useGame"
import {playSfx} from "../game/audio"

// 房间状态
const {state, startGame, leave} = game

// 玩家状态
const players = computed(() => state.players)

// 是否可以开始游戏
const canStart = computed(() => state.players.filter((p) => p.connected).length === 2)

// 是否复制房间码
const copied = ref(false)

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
</script>

<template>
	<div class="lobby" data-cakeduel-screen="lobby">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="overlay"></div>
		<div class="content">
			<div class="card glass">
				<h1 class="title-font">房间</h1>
				<div class="code-box" @click="copyCode">
					<span class="code">{{ state.roomCode }}</span>
					<span class="copy-hint">{{ copied ? "已复制！" : "点击复制" }}</span>
				</div>
				<div class="players">
					<div v-for="i in [0, 1]" :key="i" class="player-row" :class="{ empty: !players[i] }">
						<div class="seat">{{ i === 0 ? "攻" : "守" }}</div>
						<span>{{ players[i]?.name || "等待玩家…" }}</span>
						<span v-if="i === state.playerIndex" class="me">我</span>
						<span v-else-if="players[i] && !players[i].connected" class="me offline">已掉线</span>
						<span v-else-if="!players[i]" class="loading">…</span>
					</div>
				</div>
				<button class="start-btn" :disabled="!canStart" @click="startGame">
					{{ canStart ? "开始决斗" : "等待对手加入…" }}
				</button>
				<button class="leave-btn" @click="leave">离开房间</button>
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
</style>
