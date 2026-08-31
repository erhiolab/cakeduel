<script setup lang="ts">
import {computed} from "vue"
import type {CardEntity} from "../composables/useGame"
import PlayingCard from "./PlayingCard.vue"

// 牌堆容器相对卡牌的外扩尺寸(px)
const DECK_PADDING = 6

// 牌堆层叠偏移量(px)
const STACK_OFFSET = 1.5

const props = defineProps<{
	deckTop: readonly CardEntity[]
	deckCount: number
	cardWidth: number
	cardHeight: number
}>()

// 牌堆容器宽度
const width = computed(() => props.cardWidth + DECK_PADDING)

// 牌堆容器高度
const height = computed(() => props.cardHeight + DECK_PADDING)
</script>

<template>
	<div class="deck" :style="{ width: `${width}px`, height: `${height}px` }">
		<div
			v-for="(card, u) in deckTop"
			:key="card.entityId"
			class="deck-card"
			:style="{
				left: `${u * STACK_OFFSET}px`,
				top: `${u * STACK_OFFSET}px`,
				zIndex: deckTop.length - u,
			}"
		>
			<PlayingCard :width="cardWidth" :height="cardHeight" :disabled="true"/>
		</div>
		<div v-if="deckCount > 0" class="count">
			{{ deckCount }}
		</div>
		<div v-else class="zero">0</div>
	</div>
</template>

<style scoped>
.deck {
	position: relative;
}

.deck-card {
	position: absolute;
}

.count {
	position: absolute;
	bottom: -0.4rem;
	left: -0.9rem;
	z-index: 20;
	min-width: 1.8rem;
	height: 1.8rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.75rem;
	font-weight: bold;
	color: #44403c;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0.55), rgba(255, 255, 255, 0.4));
	border: 1px solid rgba(255, 255, 255, 0.6);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8), 0 4px 16px rgba(0, 0, 0, 0.15);
	backdrop-filter: blur(6px);
}

.zero {
	position: absolute;
	inset: 0;
	border: 1px dashed rgba(217, 119, 6, 0.2);
	border-radius: 0.75rem;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.6rem;
	color: rgba(255, 220, 160, 0.2);
}
</style>
