<script setup lang="ts">
import {computed} from "vue"
import {CARD_BACK, cardImage} from "../game/cards"

const props = withDefaults(
	defineProps<{
		name?: string
		faceUp?: boolean
		width?: number
		height?: number
		radius?: number
		inverted?: boolean
		selected?: boolean
		highlighted?: boolean
		dimmed?: boolean
		glow?: string
		flipDelayMs?: number
		onClick?: () => void
	}>(),
	{
		faceUp: false,
		width: 110,
		height: 150,
		radius: 12,
		inverted: false,
		selected: false,
		highlighted: false,
		dimmed: false,
		glow: "rgba(255,255,255,0.85)",
		flipDelayMs: 0,
	},
)

// 卡牌样式
const style = computed(() => ({
	width: `${props.width}px`,
	height: `${props.height}px`,
	borderRadius: `${props.radius}px`,
	perspective: "800px",
}))

// 卡牌内部样式
const innerStyle = computed(() => ({
	width: `${props.width}px`,
	height: `${props.height}px`,
	borderRadius: `${props.radius}px`,
	transitionDelay: props.flipDelayMs ? `${props.flipDelayMs}ms` : undefined,
}))

// 卡牌正面图片
const faceSrc = computed(() => (props.name ? cardImage(props.name) : CARD_BACK))
</script>

<template>
	<div class="card-outer" :style="style" :class="{ clickable: onClick && !dimmed }" @click="onClick">
		<div
			class="card-inner"
			:style="innerStyle"
			:class="{ flipped: !faceUp }"
		>
			<div class="card-face" :style="{ borderRadius: `${radius}px` }">
				<img :src="faceSrc" :alt="name || 'back'" :class="{ inverted }" draggable="false"/>
				<div class="ring"></div>
			</div>
			<div class="card-back" :style="{ borderRadius: `${radius}px` }">
				<img :src="CARD_BACK" alt="back" :class="{ inverted }" draggable="false"/>
				<div class="ring"></div>
			</div>
			<div v-if="highlighted && !selected" class="highlight-glow"></div>
			<div
				v-if="selected"
				class="select-glow"
				:style="{ boxShadow: `inset 0 0 0 2px ${glow}, 0 0 16px ${glow}` }">
			</div>
		</div>
	</div>
</template>

<style scoped>
.card-outer {
	position: relative;
	flex-shrink: 0;
}

.card-inner {
	position: relative;
	transform-style: preserve-3d;
	transition: transform 0.45s cubic-bezier(0.23, 1, 0.32, 1), opacity 0.3s ease;
	opacity: v-bind("dimmed ? 0.45 : 1");
	filter: v-bind("dimmed ? 'saturate(0.35) brightness(0.85)' : 'none'");
}

.card-inner.flipped {
	transform: rotateY(180deg);
}

.card-face,
.card-back {
	position: absolute;
	inset: 0;
	overflow: hidden;
	backface-visibility: hidden;
	-webkit-backface-visibility: hidden;
}

.card-back {
	transform: rotateY(180deg);
}

.card-face img,
.card-back img {
	width: 100%;
	height: 100%;
	object-fit: cover;
}

img.inverted {
	transform: rotate(180deg);
}

.ring {
	position: absolute;
	inset: 0;
	border-radius: inherit;
	box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
	pointer-events: none;
}

.highlight-glow {
	position: absolute;
	inset: 0;
	border-radius: inherit;
	pointer-events: none;
	animation: hl 1.2s ease-in-out infinite;
}

@keyframes hl {
	0%,
	100% {
		box-shadow: inset 0 0 0 2.5px rgba(232, 199, 58, 0.9), 0 0 14px rgba(232, 199, 58, 0.45);
	}
	50% {
		box-shadow: inset 0 0 0 2.5px rgba(255, 226, 110, 1), 0 0 30px rgba(232, 199, 58, 0.8);
	}
}

.select-glow {
	position: absolute;
	inset: 0;
	border-radius: inherit;
	pointer-events: none;
	animation: sel 1.6s ease-in-out infinite;
}

@keyframes sel {
	0%,
	100% {
		box-shadow: inset 0 0 0 2px rgba(255, 255, 255, 0.85), 0 0 16px rgba(255, 255, 255, 0.4);
	}
	50% {
		box-shadow: inset 0 0 0 2px rgba(255, 255, 255, 0.85), 0 0 28px rgba(255, 255, 255, 0.65);
	}
}

.clickable {
	cursor: pointer;
}
</style>
