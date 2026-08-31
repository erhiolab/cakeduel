<script setup lang="ts">
import {computed} from "vue"

// 默认蛋糕图标尺寸(px)
const DEFAULT_SIZE = 48

// 每侧最多显示的蛋糕数量
const MAX_TOKENS = 5

// 列宽相对蛋糕尺寸的额外宽度(px)
const COLUMN_PADDING = 16

const props = defineProps<{
	top: number
	bottom: number
	size?: number
}>()

// 蛋糕图标尺寸
const size = computed(() => props.size ?? DEFAULT_SIZE)

// 上方(对手)蛋糕
const topTokens = computed(() => Array.from({length: Math.min(props.top, MAX_TOKENS)}, (_, i) => i))

// 下方(自己)蛋糕
const bottomTokens = computed(() => Array.from({length: Math.min(props.bottom, MAX_TOKENS)}, (_, i) => i))
</script>

<template>
	<div class="cake-col" :style="{ width: `${size + COLUMN_PADDING}px` }">
		<div class="cake-panel"></div>
		<div class="spine"></div>
		<div class="spine-cap cap-top"></div>
		<div class="spine-cap cap-bottom"></div>
		<div class="tokens">
			<div class="tokens-top">
				<img
					v-for="(t, idx) in topTokens"
					:key="`t${t}`"
					src="/cakeduel/token-cake.png"
					alt="cake"
					:style="{
						width: `${size}px`,
						height: `${size}px`,
						marginBottom: idx < topTokens.length - 1 ? `-${size / 2}px` : '0px',
					}"
				/>
			</div>
			<div class="tokens-bottom">
				<img
					v-for="(t, idx) in bottomTokens"
					:key="`b${t}`"
					src="/cakeduel/token-cake.png"
					alt="cake"
					:style="{
						width: `${size}px`,
						height: `${size}px`,
						marginBottom: idx < bottomTokens.length - 1 ? `-${size / 2}px` : '0px',
					}"
				/>
			</div>
		</div>
	</div>
</template>

<style scoped>
.cake-col {
	position: relative;
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0.4rem 0.3rem;
	flex-shrink: 0;
}

.cake-panel {
	position: absolute;
	inset: 0.4rem 0.1rem;
	border-radius: 0.8rem;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0.55), rgba(255, 255, 255, 0.4) 50%, rgba(255, 255, 255, 0.55));
	border: 1px solid rgba(255, 255, 255, 0.6);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8), 0 4px 16px rgba(0, 0, 0, 0.15);
	backdrop-filter: blur(6px);
}

.spine {
	position: absolute;
	left: 50%;
	transform: translateX(-50%);
	top: 1.4rem;
	bottom: 1.4rem;
	width: 0.4rem;
	border-radius: 0.2rem;
	background: #d4944a;
	border: 1px solid #b87736;
	z-index: 1;
}

.spine-cap {
	position: absolute;
	left: 50%;
	transform: translateX(-50%);
	width: 0.8rem;
	height: 0.8rem;
	border-radius: 50%;
	background: #e8a85c;
	border: 1px solid #b87736;
	z-index: 2;
}

.cap-top {
	top: 0.9rem;
}

.cap-bottom {
	bottom: 0.9rem;
}

.tokens {
	position: relative;
	z-index: 3;
	height: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: space-between;
	padding: 0.5rem 0.35rem;
}

.tokens-top,
.tokens-bottom {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.tokens img {
	filter: drop-shadow(0 0.2rem 0.2rem rgba(0, 0, 0, 0.35));
}
</style>
