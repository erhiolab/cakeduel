<script setup lang="ts">
import {computed} from "vue"
import type {CardEntity, PlayerInfo, RevealMsg, SpectatorViewMsg, Zones} from "../composables/useGame"
import {CARD_BACK, CARDS, cardImage} from "../game/cards"

// 翻牌逐张错开时间(ms)
const FLIP_STAGGER_MS = 160

const props = withDefaults(
	defineProps<{
		view: SpectatorViewMsg | null
		zones: Zones | null
		players?: readonly PlayerInfo[] | null
		reveal?: RevealMsg | null
		cardHeight: number
		compact?: boolean
		showHands?: boolean
	}>(),
	{
		players: null,
		reveal: null,
		compact: false,
		showHands: false,
	},
)

// 卡牌宽度(按 3:4 比例)
const cardWidth = computed(() => Math.round(props.cardHeight * 0.733))

// 翻牌错开延迟
const flipDelay = (index: number): string => `${index * FLIP_STAGGER_MS}ms`

// 玩家名
const nameOf = (index: number): string => {
	return props.players?.find((p) => p.index === index)?.name || ""
}

// 蛋糕数(观战视角)
const cakesOf = (index: number): number => {
	return props.view?.players.find((p) => p.index === index)?.cakes ?? 0
}

// 手牌数(只显示数量)
const handOf = (index: number): number => {
	return props.view?.players.find((p) => p.index === index)?.handCount ?? 0
}

// 各玩家本局胜场
const winsOf = (index: number): number => {
	return props.view?.boutWinners.filter((w) => w === index).length ?? 0
}

// 阶段文案
const phaseText = computed(() => {
	const V = props.view
	if (!V) return ""
	if (V.gameEnded) return "对局结束"
	switch (V.phase) {
		case "attack":
			return "进攻回合"
		case "block":
			return "防守回合"
		case "review":
			return "质疑 / 接受"
		case "pick":
			return "特殊选牌"
		default:
			return V.phase
	}
})

// 当前行动方(防守阶段为守方, 其余为进攻方)
const activeIndex = computed(() => {
	if (!props.view) return -1
	return props.view.phase === "block" ? 1 - props.view.attackerIndex : props.view.attackerIndex
})

// 声明文案
const claimText = (claim: { claim: string; cardCount: number } | null | undefined): string => {
	if (!claim) return ""
	return `${CARDS[claim.claim]?.name ?? claim.claim} ×${claim.cardCount}`
}

// 翻开过的牌(带名称), 用于牌堆显示
interface PileCard extends CardEntity {
	name?: string
}

const withNames = (pile: readonly CardEntity[]): PileCard[] => {
	const REVEALED = props.zones?.revealedPileCards ?? {}
	return pile.map((c) => ({...c, name: REVEALED[c.entityId] ?? c.name}))
}

// 进攻牌堆
const attackCards = computed<PileCard[]>(() => withNames(props.zones?.attackPile ?? []))

// 防守牌堆
const blockCards = computed<PileCard[]>(() => withNames(props.zones?.blockPile ?? []))

// 回放/观战中翻开的牌
const revealedNames = computed(() => {
	if (!props.reveal) return []
	return props.reveal.cards.filter((c) => c.name)
})

// 是否暂停
const paused = computed(() => !!props.view?.paused)
</script>

<template>
	<div class="spectator-table" :class="{ compact }">
		<div class="board">
			<!-- 上方玩家(玩家 0) -->
			<div class="player-side">
				<div class="player-row" :class="{ active: activeIndex === 0, ended: view?.gameEnded }">
					<div class="player-chip">
						<span class="avatar">{{ nameOf(0)?.slice(0, 1) || "A" }}</span>
						<div class="player-meta">
							<span class="p-name">{{ nameOf(0) || "玩家 1" }}</span>
							<span class="p-sub">
								<img src="/cakeduel/token-cake.png" alt="" draggable="false"/>
								{{ cakesOf(0) }}
								<img class="back" src="/cakeduel/card-back-hd.jpg" alt="" draggable="false"/>
								{{ handOf(0) }} 张
							</span>
						</div>
						<span v-if="activeIndex === 0 && !view?.gameEnded" class="turn-badge">{{ phaseText }}</span>
					</div>
					<div class="wins">
						<span v-for="n in winsOf(0)" :key="`a${n}`">🏆</span>
					</div>
				</div>
				<div v-if="showHands && (zones?.playerHand ?? []).length" class="hand-row">
					<span class="hand-label">手牌</span>
					<img
						v-for="(card, i) in zones?.playerHand ?? []"
						:key="`ph${card.entityId ?? i}`"
						:src="card.name ? cardImage(card.name) : CARD_BACK"
						:alt="card.name ?? 'back'"
						draggable="false"
					/>
				</div>
			</div>

			<div class="center-area">
				<div class="info-bar">
					<span class="round">第 {{ view?.roundNumber ?? 1 }} 回合</span>
					<span class="phase">{{ phaseText }}</span>
					<span class="score">{{ winsOf(0) }} : {{ winsOf(1) }}</span>
					<span class="deck-count">牌库 {{ zones?.deckCount ?? 0 }}</span>
					<span class="discard-count">弃牌 {{ zones?.discardCount ?? 0 }}</span>
				</div>

				<div class="piles-area">
					<div class="claim-tag block" v-if="view?.blockingClaim">
						<span>守 · {{ claimText(view?.blockingClaim) }}</span>
					</div>
					<div class="pile-row">
						<div v-for="(card, i) in blockCards" :key="card.entityId" class="pile-card">
							<img
								:src="card.name ? cardImage(card.name) : CARD_BACK"
								:alt="card.name ?? 'back'"
								:style="{
									height: `${cardHeight}px`,
									width: `${cardWidth}px`,
									transitionDelay: flipDelay(i),
									marginRight: i < blockCards.length - 1 ? `${-Math.round(cardWidth * 0.45)}px` : '0px',
								}"
								:class="[card.name ? 'front' : 'back', { revealed: !!card.name, 'flip-in': !!card.name }]"
								draggable="false"
							/>
						</div>
						<div v-if="blockCards.length > 1" class="count-pill block">×{{ blockCards.length }}</div>
						<div v-if="!blockCards.length" class="pile-empty block-empty">
							<img src="/cakeduel/card-back-hd.jpg" alt="" draggable="false"/>
							<span>防守区</span>
						</div>
					</div>
					<div class="claim-tag attack" v-if="view?.attackingClaim">
						<span>攻 · {{ claimText(view?.attackingClaim) }}</span>
					</div>
					<div class="pile-row">
						<div v-for="(card, i) in attackCards" :key="card.entityId" class="pile-card">
							<img
								:src="card.name ? cardImage(card.name) : CARD_BACK"
								:alt="card.name ?? 'back'"
								:style="{
									height: `${cardHeight}px`,
									width: `${cardWidth}px`,
									transitionDelay: flipDelay(i),
									marginRight: i < attackCards.length - 1 ? `${-Math.round(cardWidth * 0.45)}px` : '0px',
								}"
								:class="[card.name ? 'front' : 'back', { revealed: !!card.name, 'flip-in': !!card.name }]"
								draggable="false"
							/>
						</div>
						<div v-if="attackCards.length > 1" class="count-pill attack">×{{ attackCards.length }}</div>
						<div v-if="!attackCards.length" class="pile-empty attack-empty">
							<img src="/cakeduel/card-back-hd.jpg" alt="" draggable="false"/>
							<span>进攻区</span>
						</div>
					</div>
				</div>

				<div class="table-hint" v-if="!view?.gameEnded">
					{{ showHands ? "回放视角 · 展示双方手牌与真实牌面" : "观战视角 · 手牌背面朝下，质疑开牌时自动揭示" }}
				</div>
				<div v-else class="table-hint ended">
					🎉 对局结束 · 获胜方：{{ nameOf(view?.gameEnded?.winner ?? 0) || `玩家 ${(view?.gameEnded?.winner ?? 0) + 1}` }}
				</div>
			</div>

			<!-- 下方玩家(玩家 1) -->
			<div class="player-side">
				<div class="player-row" :class="{ active: activeIndex === 1, ended: view?.gameEnded }">
					<div class="player-chip">
						<span class="avatar">{{ nameOf(1)?.slice(0, 1) || "B" }}</span>
						<div class="player-meta">
							<span class="p-name">{{ nameOf(1) || `玩家 ${1 + 1}` }}</span>
							<span class="p-sub">
								<img src="/cakeduel/token-cake.png" alt="" draggable="false"/>
								{{ cakesOf(1) }}
								<img class="back" src="/cakeduel/card-back-hd.jpg" alt="" draggable="false"/>
								{{ handOf(1) }} 张
							</span>
						</div>
						<span v-if="activeIndex === 1 && !view?.gameEnded" class="turn-badge">{{ phaseText }}</span>
					</div>
					<div class="wins">
						<span v-for="n in winsOf(1)" :key="`b${n}`">🏆</span>
					</div>
				</div>
				<div v-if="showHands && (zones?.opponentHand ?? []).length" class="hand-row">
					<span class="hand-label">手牌</span>
					<img
						v-for="(card, i) in zones?.opponentHand ?? []"
						:key="`oh${card.entityId ?? i}`"
						:src="card.name ? cardImage(card.name) : CARD_BACK"
						:alt="card.name ?? 'back'"
						draggable="false"
					/>
				</div>
			</div>
		</div>

		<Transition name="pop">
			<div v-if="revealedNames.length" class="reveal-overlay">
				<div class="reveal-label">质疑开牌</div>
				<div class="reveal-row">
					<div
						v-for="(card, i) in revealedNames"
						:key="card.entityId"
						class="reveal-item"
						:style="{ animationDelay: flipDelay(i) }"
					>
						<img :src="cardImage(card.name!)" :alt="card.name" draggable="false"/>
						<span>{{ CARDS[card.name!]?.name ?? card.name }}</span>
					</div>
				</div>
			</div>
		</Transition>

		<Transition name="fade">
			<div v-if="paused" class="pause-cover">
				<div class="pause-box glass">⏸ 有玩家掉线，对局已暂停</div>
			</div>
		</Transition>
	</div>
</template>

<style scoped>
.spectator-table {
	position: relative;
	width: 100%;
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	user-select: none;
	-webkit-user-select: none;
	-webkit-touch-callout: none;
}

.board {
	width: 100%;
	max-width: 54rem;
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
}

.player-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.6rem;
}

.player-side {
	width: 100%;
	display: flex;
	flex-direction: column;
	gap: 0.18rem;
}

.hand-row {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.22rem;
	flex-wrap: wrap;
	padding: 0.14rem 0.4rem;
	border-radius: 0.55rem;
	background: rgba(20, 28, 38, 0.4);
	border: 1px solid rgba(255, 255, 255, 0.08);
}

.hand-row .hand-label {
	font-size: 0.6rem;
	font-weight: 800;
	color: rgba(255, 230, 190, 0.55);
	margin-right: 0.2rem;
}

.hand-row img {
	height: 2.6rem;
	width: auto;
	border-radius: 0.22rem;
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
}

.player-chip {
	display: flex;
	align-items: center;
	gap: 0.55rem;
	min-width: 0;
	padding: 0.4rem 0.75rem;
	border-radius: 0.85rem;
	background: rgba(255, 255, 255, 0.12);
	border: 1px solid rgba(255, 255, 255, 0.16);
	backdrop-filter: blur(6px);
}

.player-row.active .player-chip {
	background: rgba(255, 205, 90, 0.18);
	border-color: rgba(255, 205, 90, 0.5);
	box-shadow: 0 0 1.2rem rgba(255, 190, 60, 0.25);
}

.player-row.ended .player-chip {
	opacity: 0.8;
}

.avatar {
	flex-shrink: 0;
	width: 2rem;
	height: 2rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.95rem;
	font-weight: 900;
	color: #fff;
	background: linear-gradient(135deg, #d97706, #92400e);
}

.player-meta {
	display: flex;
	flex-direction: column;
	gap: 0.15rem;
	min-width: 0;
}

.p-name {
	font-size: 0.85rem;
	font-weight: 800;
	color: rgba(255, 245, 225, 0.95);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	max-width: 8rem;
}

.p-sub {
	display: flex;
	align-items: center;
	gap: 0.25rem;
	font-size: 0.7rem;
	font-weight: 700;
	color: rgba(255, 230, 190, 0.7);
	white-space: nowrap;
}

.p-sub img {
	width: 0.95rem;
	height: 0.95rem;
	object-fit: cover;
	border-radius: 0.15rem;
	filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.4));
}

.p-sub img.back {
	width: 0.85rem;
	height: 1.05rem;
	margin-left: 0.25rem;
}

.turn-badge {
	flex-shrink: 0;
	margin-left: 0.5rem;
	padding: 0.25rem 0.6rem;
	border-radius: 2rem;
	font-size: 0.68rem;
	font-weight: 800;
	color: #3a2c1f;
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	animation: pulse-glow 1.6s ease-in-out infinite;
}

.wins {
	display: flex;
	gap: 0.15rem;
	font-size: 0.85rem;
	letter-spacing: 0.05rem;
}

.center-area {
	flex: 1;
	min-height: 0;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.35rem;
	padding: 0.35rem;
	border-radius: 1rem;
	background: rgba(20, 28, 38, 0.38);
	border: 1px solid rgba(255, 255, 255, 0.1);
	backdrop-filter: blur(4px);
}

.info-bar {
	display: flex;
	align-items: center;
	gap: 0.55rem;
	flex-wrap: wrap;
	justify-content: center;
	font-size: 0.7rem;
	font-weight: 800;
	color: rgba(255, 235, 200, 0.85);
}

.info-bar .round,
.info-bar .phase,
.info-bar .score {
	padding: 0.2rem 0.65rem;
	border-radius: 2rem;
	background: rgba(0, 0, 0, 0.28);
	border: 1px solid rgba(255, 255, 255, 0.12);
	white-space: nowrap;
}

.info-bar .phase {
	color: #ffd977;
}

.info-bar .deck-count,
.info-bar .discard-count {
	opacity: 0.75;
	font-weight: 600;
}

.piles-area {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.3rem;
}

.pile-row {
	display: flex;
	align-items: center;
	position: relative;
}

.pile-card {
	animation: rise 0.25s cubic-bezier(0.23, 1, 0.32, 1);
}

.pile-card img {
	border-radius: 0.35rem;
	object-fit: cover;
	box-shadow: 0 4px 14px rgba(0, 0, 0, 0.45), inset 0 0 0 1px rgba(255, 255, 255, 0.12);
	transform: perspective(40rem) rotateY(0deg);
}

.count-pill {
	position: absolute;
	right: -0.6rem;
	padding: 0.15rem 0.4rem;
	border-radius: 0.6rem;
	font-size: 0.62rem;
	font-weight: 800;
	color: #3a2c1f;
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
}

.count-pill.block {
	top: -0.2rem;
}

.count-pill.attack {
	bottom: -0.2rem;
}

.pile-card img.front.revealed {
	animation: flip-in 0.5s cubic-bezier(0.23, 1, 0.32, 1) backwards;
}

.pile-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.25rem;
	padding: 0.4rem 1rem;
	border: 1px dashed rgba(255, 255, 255, 0.18);
	border-radius: 0.7rem;
	color: rgba(255, 230, 190, 0.5);
	font-size: 0.68rem;
	font-weight: 700;
}

.pile-empty img {
	width: 2.4rem;
	height: 3.3rem;
	border-radius: 0.3rem;
	opacity: 0.75;
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
}

.claim-tag {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 0.18rem 0.75rem;
	border-radius: 2rem;
	font-size: 0.72rem;
	font-weight: 800;
	color: #fff8e6;
	background: rgba(90, 154, 64, 0.55);
	border: 1px solid rgba(160, 220, 120, 0.45);
	backdrop-filter: blur(4px);
}

.claim-tag.block {
	background: rgba(90, 130, 190, 0.55);
	border-color: rgba(150, 190, 235, 0.45);
}

.table-hint {
	font-size: 0.68rem;
	color: rgba(255, 235, 200, 0.55);
	font-weight: 600;
	text-align: center;
}

.table-hint.ended {
	color: #ffd977;
	font-weight: 800;
}

.reveal-overlay {
	position: absolute;
	inset: 0;
	z-index: 30;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	background: rgba(10, 15, 24, 0.62);
	backdrop-filter: blur(3px);
	border-radius: 1rem;
}

.reveal-label {
	padding: 0.25rem 0.9rem;
	border-radius: 2rem;
	font-size: 0.75rem;
	font-weight: 800;
	color: #fff;
	background: rgba(127, 29, 29, 0.85);
	border: 1px solid rgba(248, 113, 113, 0.4);
}

.reveal-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.reveal-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.25rem;
	animation: flip-in 0.5s cubic-bezier(0.23, 1, 0.32, 1) backwards;
}

.reveal-item img {
	height: 8rem;
	width: auto;
	border-radius: 0.45rem;
	box-shadow: 0 8px 30px rgba(0, 0, 0, 0.55);
}

.reveal-item span {
	font-size: 0.72rem;
	font-weight: 800;
	color: rgba(255, 245, 225, 0.95);
}

.pause-cover {
	position: absolute;
	inset: 0;
	z-index: 25;
	display: flex;
	align-items: flex-start;
	justify-content: center;
	padding-top: 0.8rem;
	border-radius: 1rem;
	background: rgba(10, 14, 22, 0.35);
}

.pause-box {
	padding: 0.5rem 1.2rem;
	border-radius: 2rem;
	font-size: 0.85rem;
	font-weight: 800;
	color: #fde68a;
}

/* 紧凑(小屏/横屏矮窗口) */
.spectator-table.compact .player-chip {
	padding: 0.2rem 0.5rem;
	gap: 0.35rem;
	border-radius: 0.6rem;
}

.spectator-table.compact .avatar {
	width: 1.5rem;
	height: 1.5rem;
	font-size: 0.75rem;
}

.spectator-table.compact .p-name {
	font-size: 0.72rem;
	max-width: 6rem;
}

.spectator-table.compact .p-sub {
	font-size: 0.6rem;
}

.spectator-table.compact .turn-badge {
	font-size: 0.6rem;
	padding: 0.15rem 0.45rem;
}

.spectator-table.compact .center-area {
	gap: 0.15rem;
	padding: 0.2rem;
}

.spectator-table.compact .info-bar {
	gap: 0.3rem;
	font-size: 0.6rem;
}

.spectator-table.compact .info-bar .round,
.spectator-table.compact .info-bar .phase,
.spectator-table.compact .info-bar .score {
	padding: 0.1rem 0.45rem;
}

.spectator-table.compact .claim-tag {
	padding: 0.1rem 0.5rem;
	font-size: 0.62rem;
}

.spectator-table.compact .table-hint {
	font-size: 0.58rem;
}

.spectator-table.compact .reveal-item img {
	height: 5rem;
}

.spectator-table.compact .pile-empty {
	padding: 0.2rem 0.6rem;
	font-size: 0.55rem;
}

.spectator-table.compact .pile-empty img {
	width: 1.6rem;
	height: 2.2rem;
}

.spectator-table.compact .hand-row {
	padding: 0.1rem 0.25rem;
	gap: 0.12rem;
}

.spectator-table.compact .hand-row img {
	height: 1.8rem;
}

.spectator-table.compact .hand-row .hand-label {
	font-size: 0.52rem;
}

@keyframes flip-in {
	from {
		opacity: 0;
		transform: rotateY(90deg) scale(0.85);
	}
	to {
		opacity: 1;
		transform: rotateY(0deg) scale(1);
	}
}

@keyframes rise {
	from {
		opacity: 0;
		transform: scale(0.85);
	}
	to {
		opacity: 1;
		transform: scale(1);
	}
}

@media (max-width: 700px) {
	.board {
		gap: 0.2rem;
	}

	.player-chip {
		padding: 0.2rem 0.5rem;
	}

	.wins {
		font-size: 0.7rem;
	}
}
</style>
