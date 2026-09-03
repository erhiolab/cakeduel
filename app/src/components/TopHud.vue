<script setup lang="ts">
import {computed, ref} from "vue"
import {game} from "../composables/useGame"
import {CARDS, cardSmallImage} from "../game/cards"
import {audioState, playSfx, toggleBgm} from "../game/audio"

// 认输二次确认的自动取消时间(ms)
const CONCEDE_CONFIRM_TIMEOUT_MS = 3000

// 复制反馈显示时间(ms)
const COPY_FEEDBACK_MS = 1400

const props = defineProps<{
	onHelp: () => void
	onCardPreview?: (name: string | null) => void
}>()

// 状态
const {state, act} = game

// 是否已确认认输
const concedeArmed = ref(false)

// 是否已复制房间号
const copiedRoom = ref(false)

// 认输确认自动取消定时器
let concedeTimer: number | undefined

// 复制反馈定时器
let copyTimer: number | undefined

// 复制房间号(方便分享给好友观战)
const copyRoom = () => {
	const CODE = state.roomCode
	if (!CODE) return
	playSfx("hoof")
	navigator.clipboard?.writeText(CODE).catch(() => {})
	copiedRoom.value = true
	if (copyTimer) window.clearTimeout(copyTimer)
	copyTimer = window.setTimeout(() => {
		copiedRoom.value = false
	}, COPY_FEEDBACK_MS)
}

// 对局进行中且未暂停时可认输
const canConcede = computed(() => !!state.view && !state.view.gameEnded && !state.paused)

// 认输(需二次确认, 防止误触)
const doConcede = () => {
	if (!concedeArmed.value) {
		concedeArmed.value = true
		if (concedeTimer) window.clearTimeout(concedeTimer)
		concedeTimer = window.setTimeout(() => {
			concedeArmed.value = false
		}, CONCEDE_CONFIRM_TIMEOUT_MS)
		return
	}
	if (concedeTimer) window.clearTimeout(concedeTimer)
	concedeArmed.value = false
	act({type: "concede"})
}

// 玩家名称
const myName = computed(() => state.players.find((p) => p.index === state.playerIndex)?.name || "你")

// 对手名称
const oppName = computed(() => state.players.find((p) => p.index === 1 - state.playerIndex)?.name || "对手")

// 回合文本
const turnText = computed(() => {
	if (!state.view) return ""
	const MY = state.yourTurn
	const PHASE = state.view.phase
	switch (PHASE) {
		case "attack":
			return MY ? "你的进攻回合" : `${oppName.value} 正在进攻`
		case "block":
			return MY ? "你的防守回合" : `${oppName.value} 正在防守`
		case "review":
			return MY ? "要质疑吗？" : `${oppName.value} 正在考虑`
		case "pick":
			return MY ? "选一张牌" : `${oppName.value} 正在选牌`
	}
	return ""
})

// 当前声明
const currentClaim = computed(() => state.view?.blockingClaim ?? state.view?.attackingClaim ?? null)

// 当前声明者
const claimOwner = computed(() => {
	if (!state.view || !currentClaim.value) return ""
	if (state.view.blockingClaim) {
		return state.view.blockingClaim === currentClaim.value ? oppName.value : myName.value
	}
	return state.view.attackerIndex === state.playerIndex ? myName.value : oppName.value
})

// 我的胜利次数
const myWins = computed(() => state.view?.boutWinners.filter((w) => w === state.playerIndex).length ?? 0)

// 对手胜利次数
const oppWins = computed(() => state.view?.boutWinners.filter((w) => w !== state.playerIndex).length ?? 0)

// 长按预览定时器(移动端)
let pressTimer: number | undefined

// 桌面: 悬浮立即预览
const hoverPreview = () => {
	if (pressTimer) window.clearTimeout(pressTimer)
	props.onCardPreview?.(currentClaim.value?.claim ?? null)
}

// 离开清除预览
const leavePreview = () => {
	if (pressTimer) window.clearTimeout(pressTimer)
	props.onCardPreview?.(null)
}

// 移动端: 长按 500ms 后显示大图
const pressStart = () => {
	if (pressTimer) window.clearTimeout(pressTimer)
	pressTimer = window.setTimeout(() => {
		props.onCardPreview?.(currentClaim.value?.claim ?? null)
	}, 500)
}

// 松开/移出取消长按
const pressEnd = () => {
	if (pressTimer) window.clearTimeout(pressTimer)
	pressTimer = undefined
	props.onCardPreview?.(null)
}
</script>

<template>
	<div class="hud">
		<div class="hud-inner" :class="{ mine: state.yourTurn }">
			<div class="hud-left">
				<button
					class="room-btn"
					:class="{ copied: copiedRoom }"
					:title="copiedRoom ? '房间号已复制' : '点击复制房间号，好友输入可观战'"
					@click="copyRoom"
				>
					<span class="room-prefix">🏠</span>
					<b class="room-code">{{ state.roomCode || "------" }}</b>
					<svg v-if="!copiedRoom" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor"
						 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<rect x="9" y="9" width="13" height="13" rx="2"/>
						<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
					</svg>
					<svg v-else viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor"
						 stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
						<path d="M20 6L9 17l-5-5"/>
					</svg>
				</button>
				<button class="music-btn" :class="{ muted: !audioState.bgmOn }" title="背景音乐开关" @click="toggleBgm">
					<svg v-if="audioState.bgmOn" viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
						<path d="M9 18V5l12-2v13"/>
						<circle cx="6" cy="18" r="3"/>
						<circle cx="18" cy="16" r="3"/>
					</svg>
					<svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor"
						 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d="M9 18V5l12-2v13"/>
						<circle cx="6" cy="18" r="3"/>
						<path d="M16 8l5 5M21 8l-5 5"/>
					</svg>
				</button>
				<button class="help-btn" @click="props.onHelp">
					<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2"
						 stroke-linecap="round">
						<circle cx="12" cy="12" r="10"/>
						<path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 2.2-3 4"/>
						<circle cx="12" cy="17.5" r="0.5" fill="currentColor"/>
					</svg>
					<span class="help-text">规则</span>
				</button>
				<button v-if="canConcede" class="concede-btn" :class="{ armed: concedeArmed }" @click="doConcede">
					{{ concedeArmed ? "确认认输？" : "认输" }}
				</button>
			</div>
			<div class="hud-center">
				<div class="status-pill" :class="{ active: state.yourTurn }">
					<span>{{ turnText }}</span>
					<Transition name="pop">
						<div v-if="currentClaim" :key="`${currentClaim.claim}-${currentClaim.cardCount}`"
							 class="claim-chip">
							<div class="divider"></div>
							<img
								:src="cardSmallImage(currentClaim.claim)"
								:alt="currentClaim.claim"
								@pointerenter="hoverPreview"
								@pointerleave="leavePreview"
								@pointerdown="pressStart"
								@pointerup="pressEnd"
							/>
							<div class="claim-text">
								<span class="claim-who">{{ claimOwner }}声明了</span>
								<span class="claim-name">
									{{ CARDS[currentClaim.claim]?.name ?? currentClaim.claim }}
									×
									{{ currentClaim.cardCount }}
								</span>
							</div>
						</div>
					</Transition>
				</div>
			</div>
			<div class="hud-right">
				<div v-if="state.spectatorCount > 0" class="viewer-pill" title="当前观战人数">
					👁 {{ state.spectatorCount }}
				</div>
				<div
					v-if="state.yourTurn && state.turnRemaining > 0"
					class="countdown"
					:class="{ urgent: state.turnRemaining <= 10 }"
				>
					⏱ {{ state.turnRemaining }}s
				</div>
				<div class="score-pill">
					<span>{{ myName }}</span>
					<b :class="{ lead: myWins >= oppWins && myWins > 0 }">{{ myWins }}</b>
					<img src="/cakeduel/trophy.png" alt="trophy"/>
					<b :class="{ lead: oppWins > myWins }">{{ oppWins }}</b>
					<span>{{ oppName }}</span>
				</div>
			</div>
		</div>
		<div class="hud-line" :class="{ active: state.yourTurn }"></div>
	</div>
</template>

<style scoped>
.hud {
	flex-shrink: 0;
	position: relative;
	z-index: 20;
}

.hud-inner {
	display: grid;
	grid-template-columns: 1fr auto 1fr;
	align-items: center;
	padding: 0 1.1rem;
	height: 3.2rem;
	background: rgba(28, 25, 23, 0.88);
	backdrop-filter: blur(8px);
	transition: background 0.3s;
}

.hud-inner.mine {
	background: rgba(30, 27, 24, 0.92);
}

.help-btn {
	display: flex;
	align-items: center;
	gap: 0.4rem;
	border-radius: 2rem;
	background: rgba(255, 255, 255, 0.08);
	padding: 0.35rem 0.8rem;
	color: rgba(255, 255, 255, 0.75);
	font-size: 0.75rem;
	font-weight: 700;
	transition: background 0.2s, color 0.2s;
}

.help-btn:hover {
	background: rgba(255, 255, 255, 0.16);
	color: #fff;
}

.hud-left {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.music-btn {
	width: 1.8rem;
	height: 1.8rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	color: rgba(255, 255, 255, 0.7);
	background: rgba(255, 255, 255, 0.08);
	border: 1px solid rgba(255, 255, 255, 0.15);
	transition: background 0.2s, color 0.2s;
}

.music-btn:hover {
	background: rgba(255, 255, 255, 0.16);
	color: #fff;
}

.music-btn.muted {
	color: rgba(255, 255, 255, 0.35);
}

.room-btn {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	height: 1.8rem;
	padding: 0 0.65rem;
	border-radius: 2rem;
	color: rgba(255, 235, 200, 0.9);
	background: rgba(255, 255, 255, 0.08);
	border: 1px solid rgba(255, 255, 255, 0.16);
	transition: background 0.2s, color 0.2s, transform 0.15s;
}

.room-btn:hover {
	background: rgba(255, 255, 255, 0.18);
	color: #fff;
	transform: scale(1.02);
}

.room-btn .room-prefix {
	font-size: 0.78rem;
	line-height: 1;
}

.room-btn .room-code {
	font-size: 0.8rem;
	font-weight: 800;
	letter-spacing: 0.12rem;
	color: #ffd977;
	font-variant-numeric: tabular-nums;
	white-space: nowrap;
}

.room-btn.copied {
	background: rgba(52, 211, 153, 0.18);
	border-color: rgba(52, 211, 153, 0.45);
}

.room-btn.copied .room-code {
	color: #6ee7b7;
}

.concede-btn {
	display: flex;
	align-items: center;
	border-radius: 2rem;
	background: rgba(220, 38, 38, 0.16);
	border: 1px solid rgba(248, 113, 113, 0.35);
	padding: 0.35rem 0.8rem;
	color: #fca5a5;
	font-size: 0.75rem;
	font-weight: 700;
	transition: background 0.2s, color 0.2s;
}

.concede-btn:hover {
	background: rgba(220, 38, 38, 0.3);
	color: #fff;
}

.concede-btn.armed {
	background: rgba(220, 38, 38, 0.85);
	color: #fff;
	animation: pulse-glow 1s ease-in-out infinite;
}

.hud-center {
	display: flex;
	justify-content: center;
	padding: 0 1rem;
}

.status-pill {
	display: flex;
	align-items: center;
	white-space: nowrap;
	padding: 0.35rem 1rem;
	clip-path: polygon(0.7rem 0%, calc(100% - 0.7rem) 0%, 100% 50%, calc(100% - 0.7rem) 100%, 0.7rem 100%, 0% 50%);
	background: rgba(255, 255, 255, 0.08);
	font-size: 0.85rem;
	font-weight: 800;
	color: rgba(255, 255, 255, 0.55);
	letter-spacing: 0.04rem;
	transition: background 0.3s, color 0.3s;
}

.status-pill.active {
	background: linear-gradient(135deg, #92400e, #d97706, #92400e);
	color: #fffbeb;
	animation: pulse-glow 2.8s ease-in-out infinite;
}

.claim-chip {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-left: 0.75rem;
}

.divider {
	width: 1px;
	height: 1.25rem;
	background: rgba(255, 255, 255, 0.2);
}

.claim-chip img {
	height: 1.8rem;
	width: auto;
	border-radius: 2px;
	box-shadow: 0 2px 6px rgba(0, 0, 0, 0.35);
}

.claim-text {
	display: flex;
	flex-direction: column;
	line-height: 1.05;
	gap: 0.1rem;
}

.claim-who {
	font-size: 0.6rem;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.5);
}

.claim-name {
	font-size: 0.78rem;
	font-weight: 800;
	color: rgba(255, 255, 255, 0.9);
}

.hud-right {
	display: flex;
	justify-content: flex-end;
}

.score-pill {
	display: flex;
	align-items: center;
	gap: 0.4rem;
	border-radius: 2rem;
	background: rgba(255, 255, 255, 0.08);
	padding: 0.3rem 0.9rem;
}

.countdown {
	margin-right: 0.6rem;
	border-radius: 2rem;
	padding: 0.3rem 0.8rem;
	font-size: 0.8rem;
	font-weight: 800;
	color: #fbbf24;
	background: rgba(245, 158, 11, 0.12);
	border: 1px solid rgba(245, 158, 11, 0.3);
	white-space: nowrap;
}

.viewer-pill {
	margin-right: 0.6rem;
	border-radius: 2rem;
	padding: 0.3rem 0.7rem;
	font-size: 0.75rem;
	font-weight: 800;
	color: #a7f3d0;
	background: rgba(16, 90, 70, 0.35);
	border: 1px solid rgba(110, 231, 183, 0.3);
	white-space: nowrap;
}

.countdown.urgent {
	color: #f87171;
	background: rgba(220, 38, 38, 0.15);
	border-color: rgba(248, 113, 113, 0.45);
	animation: pulse-glow 1s ease-in-out infinite;
}

.score-pill span {
	font-size: 0.65rem;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.55);
	text-transform: uppercase;
	letter-spacing: 0.04rem;
}

.score-pill b {
	font-size: 0.9rem;
	color: rgba(255, 255, 255, 0.9);
}

.score-pill b.lead {
	color: #fbbf24;
}

.score-pill img {
	width: 1.25rem;
	height: 1.25rem;
}

.hud-line {
	height: 1px;
	background: rgba(255, 255, 255, 0.08);
}

.hud-line.active {
	background: linear-gradient(90deg, transparent, rgba(245, 158, 11, 0.4), transparent);
}

/* 窄屏(手机横屏等): 收窄房间号与帮助按钮, 给中间状态留空间 */
@media (max-width: 860px) {
	.hud-inner {
		padding: 0 0.55rem;
	}

	.help-text {
		display: none;
	}

	.room-btn {
		gap: 0.2rem;
		padding: 0 0.5rem;
	}

	.room-btn .room-prefix {
		display: none;
	}

	.room-btn .room-code {
		font-size: 0.72rem;
		letter-spacing: 0.08rem;
	}
}
</style>
