<script setup lang="ts">
import {computed} from "vue"
import {game} from "../composables/useGame"
import {CARDS, cardImage} from "../game/cards"

// 声明横幅最多展示的卡牌数
const MAX_CLAIM_CARDS = 5

// 默认卡牌颜色
const DEFAULT_CARD_COLOR = "#B7C0C3"

// 游戏状态
const {state} = game

// 当前横幅
const banner = computed(() => state.banner)

// 取卡牌主题色
function colorOf(claim?: string) {
	return CARDS[claim ?? ""]?.color ?? DEFAULT_CARD_COLOR
}

// 十六进制颜色转 rgba
function rgba(hex: string, a: number) {
	const r = parseInt(hex.slice(1, 3), 16)
	const g = parseInt(hex.slice(3, 5), 16)
	const b = parseInt(hex.slice(5, 7), 16)
	return `rgba(${r},${g},${b},${a})`
}

// 横幅背景样式(按横幅类型区分)
const bannerStyle = computed(() => {
	if (!banner.value) return {}
	if (banner.value.kind === "claim") {
		const c = colorOf(banner.value.claim)
		return {
			background: `linear-gradient(to right, ${rgba(c, 0.15)}, ${rgba(c, 0.85)}, ${rgba(c, 0.85)}, ${rgba(c, 0.15)})`,
			borderTop: "1px solid rgba(255,255,255,0.1)",
			borderBottom: "1px solid rgba(255,255,255,0.1)",
		}
	}
	if (banner.value.kind === "bout_end") {
		// 胜利绿色 / 失败红色
		return banner.value.victory
			? {
				background: "linear-gradient(to right, rgba(6,78,59,0.85), rgba(4,120,87,0.95), rgba(4,120,87,0.95), rgba(6,78,59,0.85))",
				borderTop: "1px solid rgba(110,231,183,0.2)",
				borderBottom: "1px solid rgba(110,231,183,0.2)",
			}
			: {
				background: "linear-gradient(to right, rgba(127,29,29,0.85), rgba(153,27,27,0.95), rgba(153,27,27,0.95), rgba(127,29,29,0.85))",
				borderTop: "1px solid rgba(252,129,129,0.2)",
				borderBottom: "1px solid rgba(252,129,129,0.2)",
			}
	}
	// 默认琥珀色(开赛/接受等)
	return {
		background: "linear-gradient(to right, rgba(69,26,3,0.9), rgba(120,53,15,0.95), rgba(120,53,15,0.95), rgba(69,26,3,0.9))",
		borderTop: "1px solid rgba(245,158,11,0.3)",
		borderBottom: "1px solid rgba(245,158,11,0.3)",
	}
})
</script>

<template>
	<Transition name="fade">
		<div v-if="banner" :key="banner.id" class="banner-layer">
			<div class="banner-bg" :style="bannerStyle"></div>
			<div class="banner-content">
				<template v-if="banner.kind === 'claim'">
					<div class="claim-cards">
						<img v-for="i in Math.min(banner.cardCount ?? 1, MAX_CLAIM_CARDS)" :key="i" :src="cardImage(banner.claim!)" :alt="banner.claim"/>
					</div>
					<span class="claim-label">
						{{ banner.cardCount }} × {{ CARDS[banner.claim!]?.name ?? banner.claim }}
					</span>
				</template>
				<template v-else-if="banner.kind === 'accepted'">
					<span class="big-text amber">接受</span>
				</template>
				<template v-else-if="banner.kind === 'challenge'">
					<span class="big-text red">质疑！</span>
				</template>
				<template v-else-if="banner.kind === 'bout_start'">
					<span class="big-text amber">第 {{ banner.boutNumber }} 局</span>
				</template>
				<template v-else-if="banner.kind === 'bout_end'">
					<div class="end-box">
						<span class="big-text" :class="banner.victory ? 'green' : 'red'">
							{{ banner.victory ? "胜利" : "失败" }}
						</span>
						<span class="end-reason">{{ banner.reason }}</span>
					</div>
				</template>
			</div>
		</div>
	</Transition>
</template>

<style scoped>
.banner-layer {
	position: absolute;
	inset: 0;
	z-index: 70;
	display: flex;
	align-items: center;
	justify-content: center;
	pointer-events: none;
}

.banner-bg {
	position: absolute;
	inset: 0;
	background: rgba(0, 0, 0, 0.5);
}

.banner-content {
	position: relative;
	animation: banner-in 0.35s cubic-bezier(0.22, 1, 0.36, 1);
	display: flex;
	align-items: center;
	justify-content: center;
}

.claim-cards {
	display: flex;
	align-items: flex-end;
	height: 7rem;
	margin-right: 1rem;
}

.claim-cards img {
	height: 100%;
	width: auto;
	object-fit: contain;
	margin-left: -1.6rem;
	position: relative;
	filter: drop-shadow(0 0.3rem 0.5rem rgba(0, 0, 0, 0.4));
}

.claim-cards img:first-child {
	margin-left: 0;
}

.claim-label {
	font-size: 1.6rem;
	font-weight: 900;
	color: #fff;
	text-shadow: 0 2px 6px rgba(0, 0, 0, 0.6);
}

.big-text {
	font-size: 2rem;
	font-weight: 900;
	letter-spacing: 0.1em;
	text-transform: uppercase;
	text-shadow: 0 0 12px rgba(0, 0, 0, 0.5);
}

.amber {
	color: #fcd34d;
}

.red {
	color: #f87171;
	text-shadow: 0 0 12px rgba(248, 113, 113, 0.5);
}

.green {
	color: #a7f3d0;
	text-shadow: 0 0 14px rgba(110, 231, 183, 0.6);
}

.end-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.4rem;
}

.end-reason {
	font-size: 0.9rem;
	font-weight: 600;
	color: rgba(255, 255, 255, 0.8);
}
</style>
