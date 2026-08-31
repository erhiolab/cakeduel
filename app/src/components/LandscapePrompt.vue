<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue"

// 视为移动端的最大宽度(px)
const MOBILE_BREAKPOINT = 900

// 方向检测兜底轮询间隔(ms)
const POLL_INTERVAL_MS = 500

// 是否为移动设备
const isMobile = ref(false)

// 是否为竖屏
const isPortrait = ref(false)

// 轮询定时器
let timer: number | undefined

// 检查当前设备与屏幕方向
const check = () => {
	isMobile.value = window.matchMedia("(pointer: coarse)").matches || window.innerWidth < MOBILE_BREAKPOINT
	isPortrait.value = window.innerHeight > window.innerWidth
}

// 移动端竖屏时显示提示
const show = computed(() => isMobile.value && isPortrait.value)

onMounted(() => {
	check()
	window.addEventListener("resize", check)
	window.addEventListener("orientationchange", check)
	// 兜底轮询(应对模拟器/部分浏览器不派发 resize 的情况)
	timer = window.setInterval(check, POLL_INTERVAL_MS)
})

onUnmounted(() => {
	window.removeEventListener("resize", check)
	window.removeEventListener("orientationchange", check)
	if (timer) window.clearInterval(timer)
})
</script>

<template>
	<Transition name="fade">
		<div v-if="show" class="prompt">
			<div class="prompt-box glass">
				<div class="phone-icon">
					<span class="phone-screen"></span>
				</div>
				<h2 class="title-font">请横屏游玩</h2>
				<p>旋转设备至横向以获得最佳体验</p>
			</div>
		</div>
	</Transition>
</template>

<style scoped>
.prompt {
	position: fixed;
	inset: 0;
	z-index: 9999;
	background: rgba(16, 27, 18, 0.94);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1.5rem;
}

.prompt-box {
	max-width: 20rem;
	width: 100%;
	border-radius: 1.2rem;
	padding: 2rem 1.5rem;
	text-align: center;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(255, 255, 255, 0.75));
}

.prompt-box h2 {
	font-size: 1.6rem;
	margin: 1rem 0 0.4rem;
	color: #3a2c1f;
}

.prompt-box p {
	font-size: 0.9rem;
	color: #7a6a55;
}

.phone-icon {
	width: 5.2rem;
	height: 3.2rem;
	border: 0.25rem solid #4a5a40;
	border-radius: 0.6rem;
	margin: 0 auto;
	position: relative;
	transform: rotate(90deg);
	animation: tilt 2.2s ease-in-out infinite;
}

.phone-screen {
	position: absolute;
	inset: 0.35rem;
	border-radius: 0.2rem;
	background: linear-gradient(135deg, #7ab55c, #4a5a40);
}

@keyframes tilt {
	0%,
	100% {
		transform: rotate(90deg);
	}
	50% {
		transform: rotate(0deg);
	}
}
</style>
