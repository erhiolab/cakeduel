<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue"
import {game} from "../composables/useGame"
import {selection, toggleCardSelection} from "../composables/useSelection"
import {cardImage} from "../game/cards"
import TopHud from "../components/TopHud.vue"
import CardFan from "../components/CardFan.vue"
import CakeTokens from "../components/CakeTokens.vue"
import DeckView from "../components/DeckView.vue"
import PilesView from "../components/PilesView.vue"
import ActionBar from "../components/ActionBar.vue"
import BannerLayer from "../components/BannerLayer.vue"
import WolfyTaunt from "../components/WolfyTaunt.vue"
import HelpOverlay from "../components/HelpOverlay.vue"
import ChatBox from "../components/ChatBox.vue"
import type {CardEntity} from "../composables/useGame"

// 游戏状态
const {state, clearHandReveal} = game

// 帮助弹窗
const showHelp = ref(false)

// 视口大小
const viewport = ref({w: window.innerWidth, h: window.innerHeight})

// 预览卡片
const previewCard = ref<CardEntity | null>(null)

// 预览卡片定时器
let previewTimer: number | undefined

// 视口大小变化时触发
const onResize = () => {
	viewport.value = {w: window.innerWidth, h: window.innerHeight}
}

// 是否紧凑布局
const compact = computed(() => viewport.value.h < 700 || viewport.value.w < 960)

// 卡片尺寸
const cardHeight = computed(() => {
	const H = viewport.value.h
	const BASE = compact.value ? H * 0.22 : H * 0.21
	return Math.round(Math.max(compact.value ? 84 : 96, Math.min(compact.value ? 150 : 185, BASE)))
})

// 卡片宽度
const cardWidth = computed(() => Math.round(cardHeight.value * 0.733))

// 卡片重叠
const overlap = computed(() => Math.round(cardWidth.value * 0.34))

// 堆叠重叠
const pileOverlap = computed(() => Math.round(cardWidth.value * 0.55))

// 蛋糕尺寸
const cakeSize = computed(() => Math.round(Math.max(24, Math.min(38, viewport.value.h * 0.036))))

// 对手卡片尺寸
const oppCardHeight = computed(() => Math.round(cardHeight.value * 0.72))

// 我的牌
const myHand = computed(() => state.zones?.playerHand ?? [])

// 对手牌
const oppHand = computed(() => state.zones?.opponentHand ?? [])

// 我的蛋糕数
const myCakes = computed(() => state.view?.me.cakes ?? 3)

// 对手蛋糕数
const oppCakes = computed(() => state.view?.opponent.cakes ?? 4)

// 我的出牌操作
const claimAction = computed(() => state.legal.find((l) => l.type === "claim"))

// 是否可以出牌
const canSelect = computed(() => state.yourTurn && !!claimAction.value && !state.reveal && !state.view?.gameEnded,)

// 点击牌时触发
const onCardSelect = (card: CardEntity) => {
	if (!canSelect.value) return
	toggleCardSelection(card.entityId)
}

// 鼠标悬停时触发
const onHover = (card: CardEntity | null) => {
	if (previewTimer) window.clearTimeout(previewTimer)
	if (!card || card.name == null) {
		previewCard.value = null
		return
	}
	previewTimer = window.setTimeout(() => {
		previewCard.value = card
	}, 180)
}

onMounted(() => {
	window.addEventListener("resize", onResize)
})

onUnmounted(() => {
	window.removeEventListener("resize", onResize)
	if (previewTimer) window.clearTimeout(previewTimer)
})
</script>

<template>
	<div class="game" data-cakeduel-screen="game" :class="{ compact }">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="shade"></div>
		<CakeTokens class="side left" :top="oppCakes" :bottom="myCakes" :size="cakeSize"/>
		<DeckView
			class="side right"
			:deck-top="state.zones?.deckTop ?? []"
			:deck-count="state.zones?.deckCount ?? 0"
			:card-width="cardWidth"
			:card-height="cardHeight"
		/>
		<div class="layout">
			<Transition name="pop">
				<div v-if="state.connectionLost" class="conn-banner">
					连接断开，正在重连…
				</div>
				<div v-else-if="state.paused" class="conn-banner">
					对局已暂停，等待对方重连…
				</div>
			</Transition>
			<TopHud :on-help="() => (showHelp = true)"/>
			<div class="table">
				<Transition name="fade">
					<div v-if="state.reveal" class="table-dim"></div>
				</Transition>
				<div class="center">
					<div class="opp-area">
						<CardFan
							:cards="oppHand"
							:face-up="false"
							:inverted="true"
							:disabled="true"
							:card-width="cardWidth"
							:card-height="oppCardHeight"
							:overlap="Math.round(cardWidth * 0.42)"
						/>
					</div>
					<div class="piles-area">
						<PilesView
							:card-width="cardWidth"
							:card-height="cardHeight"
							:overlap="pileOverlap"
							:compact="compact"
						/>
					</div>
				</div>
			</div>
			<ActionBar/>
			<div class="hand-area">
				<div class="hand-label">
					<span>{{ state.view?.me.handLimit ?? 4 }} 张手牌</span>
					<span class="hint-text">点击手牌出牌</span>
				</div>
				<CardFan
					:cards="myHand"
					:face-up="true"
					:disabled="!canSelect"
					:selected-ids="selection.ids"
					:card-width="cardWidth"
					:card-height="cardHeight"
					:overlap="overlap"
					:hover-preview="onHover"
					@select="onCardSelect"
				/>
			</div>
		</div>
		<BannerLayer/>
		<WolfyTaunt/>
		<Transition name="pop">
			<div v-if="state.toast" class="toast glass">{{ state.toast }}</div>
		</Transition>
		<Transition name="pop">
			<div v-if="state.error" class="error">{{ state.error }}</div>
		</Transition>
		<Transition name="fade">
			<div v-if="previewCard" class="preview">
				<img :src="cardImage(previewCard.name!)" :alt="previewCard.name"/>
			</div>
		</Transition>
		<Transition name="fade">
			<div v-if="state.handReveal" class="hand-reveal" @click="clearHandReveal()">
				<div class="hand-reveal-box glass">
					<h3>{{ state.handReveal.player === state.playerIndex ? "你的手牌被查看" : "对手的手牌被查看" }}</h3>
					<div class="hand-reveal-cards">
						<img v-for="(name, i) in state.handReveal.cards" :key="i" :src="cardImage(name)" :alt="name"/>
					</div>
					<p>点击任意处关闭</p>
				</div>
			</div>
		</Transition>
		<HelpOverlay :open="showHelp" @close="showHelp = false"/>
		<ChatBox/>
	</div>
</template>

<style scoped>
.game {
	position: relative;
	width: 100%;
	height: 100%;
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

.layout {
	position: relative;
	z-index: 10;
	height: 100%;
	display: flex;
	flex-direction: column;
}

.table {
	position: relative;
	flex: 1;
	min-height: 0;
}

.center {
	position: absolute;
	inset: 0;
	z-index: 30;
	min-width: 0;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.6rem;
}

.side.left {
  position: absolute;
  left: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  z-index: 15;
  height: 68vh;
  max-height: 44rem;
  min-height: 20rem;
}

.side.right {
	position: absolute;
	right: 1rem;
	top: 50%;
	transform: translateY(-50%);
	z-index: 15;
}

.table-dim {
	position: absolute;
	inset: 0;
	z-index: 20;
	background: rgba(15, 20, 30, 0.55);
	border-radius: 0.8rem;
	pointer-events: none;
}

.opp-area {
	flex-shrink: 0;
}

.piles-area {
	flex-shrink: 0;
}

.hand-area {
	flex-shrink: 0;
	padding: 0.3rem 0.8rem 0.6rem;
	position: relative;
	z-index: 30;
}

.hand-label {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.8rem;
	font-size: 0.7rem;
	font-weight: 700;
	color: rgba(255, 230, 190, 0.65);
	margin-bottom: 0.2rem;
}

.hint-text {
	font-weight: 500;
	opacity: 0.6;
}

.conn-banner {
	position: absolute;
	top: 0.6rem;
	left: 50%;
	transform: translateX(-50%);
	z-index: 88;
	border-radius: 2rem;
	padding: 0.45rem 1.2rem;
	font-size: 0.85rem;
	font-weight: 800;
	color: #fde68a;
	background: rgba(120, 53, 15, 0.92);
	border: 1px solid rgba(251, 191, 36, 0.45);
	box-shadow: 0 4px 18px rgba(0, 0, 0, 0.35);
	backdrop-filter: blur(6px);
	animation: pulse-glow 1.6s ease-in-out infinite;
	white-space: nowrap;
}

.toast {
	position: absolute;
	left: 50%;
	bottom: 1rem;
	transform: translateX(-50%);
	z-index: 85;
	border-radius: 0.8rem;
	padding: 0.55rem 1.1rem;
	font-size: 0.8rem;
	font-weight: 700;
	color: #44403c;
}

.error {
	position: absolute;
	left: 50%;
	bottom: 1rem;
	transform: translateX(-50%);
	z-index: 86;
	border-radius: 0.8rem;
	padding: 0.55rem 1.1rem;
	font-size: 0.8rem;
	font-weight: 700;
	color: #fecaca;
	background: rgba(127, 29, 29, 0.85);
	border: 1px solid rgba(248, 113, 113, 0.4);
	backdrop-filter: blur(6px);
}

.preview {
	position: absolute;
	right: 18%;
	top: 50%;
	transform: translateY(-50%);
	z-index: 100;
	pointer-events: none;
}

.preview img {
	height: 55vh;
	max-height: 30rem;
	width: auto;
	border-radius: 0.8rem;
	box-shadow: 0 0 80px 16px rgba(255, 180, 60, 0.07), 0 0 40px rgba(255, 200, 100, 0.1), 0 20px 60px rgba(0, 0, 0, 0.55);
}

.hand-reveal {
	position: absolute;
	inset: 0;
	z-index: 95;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(20, 25, 35, 0.55);
	backdrop-filter: blur(4px);
	cursor: pointer;
}

.hand-reveal-box {
	border-radius: 1rem;
	padding: 1.2rem 1.5rem;
	text-align: center;
	max-width: 30rem;
	width: 90%;
}

.hand-reveal-box h3 {
	font-size: 1rem;
	color: #3a2c1f;
	margin-bottom: 0.8rem;
}

.hand-reveal-cards {
	display: flex;
	justify-content: center;
	gap: 0.5rem;
	flex-wrap: wrap;
}

.hand-reveal-cards img {
	height: 8rem;
	width: auto;
	border-radius: 0.4rem;
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.hand-reveal-box p {
	margin-top: 0.7rem;
	font-size: 0.7rem;
	color: #9a7a55;
}
</style>
