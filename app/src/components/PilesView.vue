<script setup lang="ts">
import {computed} from "vue"
import {game} from "../composables/useGame"
import type {CardEntity} from "../composables/useGame"
import PlayingCard from "./PlayingCard.vue"

// 翻牌逐张错开的时间(ms)
const FLIP_STAGGER_MS = 160

const props = defineProps<{
	cardWidth: number
	cardHeight: number
	overlap: number
	compact: boolean
}>()

// 游戏状态
const {state} = game

// 牌堆中的牌(可带翻开状态)
interface PileCard extends CardEntity {
	revealed: boolean
	name?: string
}

// 攻击牌堆: 命中翻开记录时显示牌面
const attackCards = computed<PileCard[]>(() => {
	const PILE = state.zones?.attackPile ?? []
	const REVEALED = state.revealedPileNames
	return PILE.map((c) => ({...c, revealed: REVEALED[c.entityId] != null, name: REVEALED[c.entityId] ?? c.name}))
})

// 防守牌堆: 命中翻开记录时显示牌面
const blockCards = computed<PileCard[]>(() => {
	const PILE = state.zones?.blockPile ?? []
	const REVEALED = state.revealedPileNames
	return PILE.map((c) => ({...c, revealed: REVEALED[c.entityId] != null, name: REVEALED[c.entityId] ?? c.name}))
})

// 质疑翻开动画数据(牌堆已清空时单独渲染)
const revealCards = computed<PileCard[]>(() => {
	if (!state.reveal) return []
	return state.reveal.cards.map((c) => ({...c, revealed: true}))
})

// 牌堆内除最后一张外向右重叠
const pileStyle = (cards: PileCard[]) => {
	return {
		marginRight: cards.length > 1 ? `${-props.overlap}px` : "0",
	}
}
</script>

<template>
	<div class="piles" :class="{ compact }">
		<Transition name="fade">
			<div v-if="revealCards.length > 0" class="reveal-strip" :class="{ inverted: state.reveal?.pile === 'block_pile' }">
				<div
					v-for="(card, i) in revealCards"
					:key="card.entityId"
					class="reveal-item"
					:style="{ transitionDelay: `${i * FLIP_STAGGER_MS}ms` }"
				>
					<div class="flip-in" :style="{ animationDelay: `${i * FLIP_STAGGER_MS}ms` }">
						<PlayingCard
							:name="card.name"
							:face-up="true"
							:width="cardWidth"
							:height="cardHeight"
							:inverted="state.reveal?.pile === 'block_pile'"
							:disabled="true"
						/>
					</div>
				</div>
			</div>
		</Transition>
		<div class="pile-row" :class="{ inverted: true }">
			<div v-for="(card, i) in blockCards" :key="card.entityId" class="pile-item" :style="pileStyle(blockCards)">
				<div class="flip-wrap" :class="{ revealed: card.revealed }" :style="{ transitionDelay: `${i * FLIP_STAGGER_MS}ms` }">
					<PlayingCard
						:name="card.name"
						:face-up="card.revealed"
						:width="cardWidth"
						:height="cardHeight"
						:inverted="true"
						:disabled="true"
						:flip-delay-ms="i * FLIP_STAGGER_MS"
					/>
				</div>
			</div>
		</div>
		<div class="pile-row">
			<div v-for="(card, i) in attackCards" :key="card.entityId" class="pile-item" :style="pileStyle(attackCards)">
				<div class="flip-wrap" :class="{ revealed: card.revealed }" :style="{ transitionDelay: `${i * FLIP_STAGGER_MS}ms` }">
					<PlayingCard
						:name="card.name"
						:face-up="card.revealed"
						:width="cardWidth"
						:height="cardHeight"
						:disabled="true"
						:flip-delay-ms="i * FLIP_STAGGER_MS"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.piles {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.4rem;
	pointer-events: none;
}

.piles.compact {
	flex-direction: row;
	align-items: flex-start;
}

.pile-row {
	display: flex;
	align-items: flex-end;
}

.pile-row.inverted {
	transform: scale(-1, 1);
}

.pile-item {
	position: relative;
	animation: rise 0.25s cubic-bezier(0.23, 1, 0.32, 1);
}

.flip-wrap {
	transition: transform 0.2s ease;
}

.reveal-strip {
	position: absolute;
	inset: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.3rem;
	z-index: 40;
	pointer-events: none;
}

.reveal-strip.inverted {
	transform: scale(-1, 1);
}

.reveal-item {
	animation: flip-in 0.5s cubic-bezier(0.23, 1, 0.32, 1) backwards;
}

.flip-in {
	animation: flip-in 0.5s cubic-bezier(0.23, 1, 0.32, 1) backwards;
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
</style>
