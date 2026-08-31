<script setup lang="ts">
import {computed} from "vue"
import {game} from "../composables/useGame"
import {CARDS, cardSmallImage} from "../game/cards"

const props = defineProps<{
	onHelp: () => void
}>()

// 状态
const {state} = game

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
</script>

<template>
	<div class="hud">
		<div class="hud-inner" :class="{ mine: state.yourTurn }">
			<div class="hud-left">
				<button class="help-btn" @click="props.onHelp">
					<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2"
						 stroke-linecap="round">
						<circle cx="12" cy="12" r="10"/>
						<path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 2.2-3 4"/>
						<circle cx="12" cy="17.5" r="0.5" fill="currentColor"/>
					</svg>
					规则
				</button>
			</div>
			<div class="hud-center">
				<div class="status-pill" :class="{ active: state.yourTurn }">
					<span>{{ turnText }}</span>
					<Transition name="pop">
						<div v-if="currentClaim" :key="`${currentClaim.claim}-${currentClaim.cardCount}`"
							 class="claim-chip">
							<div class="divider"></div>
							<img :src="cardSmallImage(currentClaim.claim)" :alt="currentClaim.claim"/>
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
</style>
