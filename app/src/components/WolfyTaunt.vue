<script setup lang="ts">
import {ref, watch} from "vue"
import {game} from "../composables/useGame"
import {WAGGLE_FRAMES} from "../game/cards"

// 狼爵士动画
const {state} = game

// 动画帧
const frame = ref(0)

// 动画定时器
let timer: number | undefined

watch(() => state.wolfyTaunt, (v) => {
	if (timer) window.clearInterval(timer)
	if (v > 0) {
		frame.value = 0
		timer = window.setInterval(() => {
			frame.value = (frame.value + 1) % WAGGLE_FRAMES.length
		}, 110)
	}
})
</script>

<template>
	<Transition name="fade">
		<div v-if="state.wolfyTaunt > 0" class="taunt">
			<img :src="WAGGLE_FRAMES[frame]" alt="狼爵士"/>
		</div>
	</Transition>
</template>

<style scoped>
.taunt {
	position: absolute;
	left: 50%;
	top: 50%;
	transform: translate(-50%, -50%);
	z-index: 80;
	pointer-events: none;
	filter: drop-shadow(0 0.6rem 1.4rem rgba(0, 0, 0, 0.5));
}

.taunt img {
	height: 9rem;
	width: auto;
}
</style>
