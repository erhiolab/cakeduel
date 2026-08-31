<script setup lang="ts">
import {computed} from "vue"
import type {CardEntity} from "../composables/useGame"
import PlayingCard from "./PlayingCard.vue"

// 默认卡片宽度(px)
const DEFAULT_CARD_WIDTH = 110

// 默认卡片高度(px)
const DEFAULT_CARD_HEIGHT = 150

// 默认卡片叠(px)
const DEFAULT_OVERLAP = 38

// 默认扇形弧度(deg)
const ARC_MAX_ROTATION = 5

// 默认卡片下沉(px)
const ARC_DROOP_PX = 14

const props = withDefaults(
	defineProps<{
		cards: readonly CardEntity[]
		faceUp?: boolean
		inverted?: boolean
		disabled?: boolean
		selectedIds?: Set<number>
		highlightedIds?: Set<number>
		dimmedIds?: Set<number>
		cardWidth?: number
		cardHeight?: number
		overlap?: number
		hoverPreview?: (card: CardEntity | null) => void
	}>(),
	{
		faceUp: true,
		inverted: false,
		disabled: false,
		selectedIds: () => new Set(),
		highlightedIds: () => new Set(),
		dimmedIds: () => new Set(),
		cardWidth: DEFAULT_CARD_WIDTH,
		cardHeight: DEFAULT_CARD_HEIGHT,
		overlap: DEFAULT_OVERLAP,
		hoverPreview: undefined,
	},
)

const emit = defineEmits<{
	select: [card: CardEntity]
}>()

// 根据位置计算扇形弧度与下沉
const arcTransform = (index: number, total: number) => {
	if (total <= 1) return {rotation: 0, yOffset: 0}
	const A = (index - (total - 1) / 2) / ((total - 1) / 2)
	return {
		rotation: A * ARC_MAX_ROTATION * (props.inverted ? -1 : 1),
		yOffset: A * A * ARC_DROOP_PX * (props.inverted ? -1 : 1),
	}
}

// 手牌总数
const total = computed(() => props.cards.length)
</script>

<template>
	<div v-if="cards.length > 0" class="fan">
		<div
			v-for="(card, idx) in cards"
			:key="card.entityId"
			class="fan-item"
			:style="{
				zIndex: idx + 1,
				marginRight: idx < total - 1 ? `${-overlap}px` : '0',
				transform: `rotate(${arcTransform(idx, total).rotation}deg) translateY(${arcTransform(idx, total).yOffset}px)`,
			}"
			@pointerenter="hoverPreview?.(card)"
			@pointerleave="hoverPreview?.(null)"
		>
			<PlayingCard
				:name="card.name"
				:face-up="faceUp"
				:width="cardWidth"
				:height="cardHeight"
				:inverted="inverted"
				:disabled="disabled"
				:selected="selectedIds.has(card.entityId)"
				:highlighted="highlightedIds.has(card.entityId)"
				:dimmed="dimmedIds.has(card.entityId)"
				:on-click="disabled ? undefined : () => emit('select', card)"
			/>
		</div>
	</div>
	<div v-else class="empty">
		<span>无手牌</span>
	</div>
</template>

<style scoped>
.fan {
	display: flex;
	align-items: flex-end;
	justify-content: center;
}

.fan-item {
	position: relative;
}

.empty {
	display: flex;
	justify-content: center;
	padding: 0.6rem 0;
	color: rgba(255, 220, 160, 0.35);
	font-size: 0.8rem;
	font-style: italic;
}
</style>
