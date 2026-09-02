<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue"
import {game} from "../composables/useGame"

// 游戏状态
const {state, rematch, leave} = game

// 游戏结果(整场结束优先采用服务端权威 youWon, 防止本地身份时序错误)
const won = computed(() => (state.view?.gameEnded ? !!state.youWon : state.view?.gameEnded?.winner === state.playerIndex))

// 我的胜利次数
const myWins = computed(() => state.view?.boutWinners.filter((w) => w === state.playerIndex).length ?? 0)

// 对手胜利次数
const oppWins = computed(() => state.view?.boutWinners.filter((w) => w !== state.playerIndex).length ?? 0)

// 我的名称
const myName = computed(() => state.players.find((p) => p.index === state.playerIndex)?.name || "你")

// 对手名称
const oppName = computed(() => state.players.find((p) => p.index === 1 - state.playerIndex)?.name || "对手")

// 我是否已同意再来一局
const mineVoted = computed(() => state.rematchVotes[state.playerIndex])

// 对方是否已同意再来一局
const oppVoted = computed(() => state.rematchVotes[1 - state.playerIndex])

// 再来一局按钮文案
const againText = computed(() => {
	if (mineVoted) return "等待对方同意…"
	return "再来一场"
})

// 再来一局提示文案
const againHint = computed(() => {
	if (mineVoted) return "已同意，等待对方确认后开始新一局"
	if (oppVoted) return "对方已同意再来一局，点击「再来一场」开始"
	return "需要双方都同意后才会开始新一局"
})

// 视口尺寸(小屏切换为横向布局)
const viewport = ref({w: window.innerWidth, h: window.innerHeight})

// 视口变化时更新
const onResize = () => {
	viewport.value = {w: window.innerWidth, h: window.innerHeight}
}

// 高度不够时使用横向布局: 左侧胜负与比分/右侧操作按钮
const compact = computed(() => viewport.value.h < 620)

onMounted(() => {
	window.addEventListener("resize", onResize)
})

onUnmounted(() => {
	window.removeEventListener("resize", onResize)
})
</script>

<template>
	<div class="results" data-cakeduel-screen="results" :class="{ compact }">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="overlay"></div>
		<div class="content">
			<div class="card" :class="{ win: won, lose: !won, compact }">
				<div class="result-left">
					<img class="trophy" :src="won ? '/cakeduel/trophy.png' : '/cakeduel/token-cake.png'" alt=""/>
					<h1 class="title-font">{{ won ? "你赢了！" : "你输了" }}</h1>
					<p class="sub">{{ won ? `击败了 ${oppName}` : `输给了 ${oppName}` }}</p>
					<div class="score">
						<div class="score-item">
							<span>{{ myName }}</span>
							<b>{{ myWins }}</b>
						</div>
						<span class="vs">:</span>
						<div class="score-item">
							<span>{{ oppName }}</span>
							<b>{{ oppWins }}</b>
						</div>
					</div>
				</div>
				<div class="result-right">
					<button class="again" :disabled="mineVoted" @click="rematch">{{ againText }}</button>
					<p class="again-hint">{{ againHint }}</p>
					<p class="replay-note">🎞 回放已保存，主菜单可查看</p>
					<button class="menu" @click="leave">返回主菜单</button>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.results {
	position: relative;
	width: 100%;
	height: 100%;
	overflow: hidden;
	display: flex;
	align-items: center;
	justify-content: center;
}

.bg {
	position: absolute;
	inset: 0;
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.overlay {
	position: absolute;
	inset: 0;
	background: rgba(20, 30, 40, 0.5);
	box-shadow: inset 0 0 6rem 1.5rem rgba(30, 45, 60, 0.6);
}

.content {
	position: relative;
	z-index: 10;
	width: 100%;
	max-width: 24rem;
	padding: 1rem;
}

.result-left,
.result-right {
	width: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.55rem;
}

.card {
	border-radius: 1.1rem;
	padding: 1.8rem 1.5rem;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.7rem;
	text-align: center;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(255, 255, 255, 0.7));
	border: 1px solid rgba(255, 255, 255, 0.8);
	box-shadow: 0 18px 48px rgba(0, 0, 0, 0.35);
	backdrop-filter: blur(8px);
	animation: rise-in 0.4s ease both;
}

.card.win {
	border-top: 0.35rem solid #34d399;
}

.card.lose {
	border-top: 0.35rem solid #f87171;
}

.trophy {
	width: 4rem;
	height: 4rem;
	object-fit: contain;
	filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
}

.card h1 {
	font-size: 1.8rem;
	color: #3a2c1f;
}

.sub {
	font-size: 0.9rem;
	color: #7a6a55;
	font-weight: 600;
}

.score {
	display: flex;
	align-items: center;
	gap: 1rem;
	margin: 0.5rem 0;
}

.score-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.2rem;
}

.score-item span {
	font-size: 0.75rem;
	font-weight: 700;
	color: #7a6a55;
}

.score-item b {
	font-size: 2.2rem;
	color: #3a2c1f;
}

.vs {
	font-size: 1.4rem;
	font-weight: 900;
	color: #b9a58c;
}

.again {
	width: 100%;
	border-radius: 0.85rem;
	padding: 0.9rem;
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	color: #3a2c1f;
	font-weight: 900;
	font-size: 1rem;
	box-shadow: 0 6px 20px rgba(232, 162, 58, 0.35);
	transition: transform 0.15s;
}

.again:hover {
	transform: scale(1.02);
}

.again:disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.again-hint {
	font-size: 0.7rem;
	color: #9a7a55;
}

.replay-note {
	font-size: 0.68rem;
	font-weight: 600;
	color: #b45309;
}

.menu {
	width: 100%;
	padding: 0.55rem;
	border-radius: 0.7rem;
	font-size: 0.85rem;
	font-weight: 700;
	color: #6b5438;
	transition: background 0.2s;
}

.menu:hover {
	background: rgba(0, 0, 0, 0.05);
}

/* 小屏/矮窗口: 横向布局, 左侧胜负比分, 右侧操作按钮 */
.results.compact .content {
	max-width: 46rem;
	padding: 0.5rem 0.75rem;
}

.card.compact {
	flex-direction: row;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: 1rem 1.25rem;
	text-align: left;
	border-radius: 0.9rem;
}

.card.compact .result-left {
	width: 40%;
	flex-shrink: 0;
	justify-content: center;
	gap: 0.2rem;
	padding-right: 1rem;
	border-right: 1px dashed rgba(107, 84, 56, 0.3);
}

.card.compact .result-right {
	flex: 1;
	min-width: 0;
	justify-content: center;
	gap: 0.4rem;
}

.card.compact .trophy {
	width: 3rem;
	height: 3rem;
}

.card.compact h1 {
	font-size: 1.5rem;
}

.card.compact .sub {
	font-size: 0.78rem;
}

.card.compact .score {
	gap: 0.7rem;
	margin: 0.2rem 0 0;
}

.card.compact .score-item span {
	font-size: 0.68rem;
}

.card.compact .score-item b {
	font-size: 1.9rem;
}

.card.compact .vs {
	font-size: 1.1rem;
}

.card.compact .again {
	padding: 0.6rem;
	border-radius: 0.65rem;
	font-size: 0.95rem;
}

.card.compact .again-hint {
	font-size: 0.62rem;
}

.card.compact .replay-note {
	font-size: 0.6rem;
}

.card.compact .menu {
	padding: 0.4rem;
	border-radius: 0.6rem;
	font-size: 0.8rem;
}

/* 极矮窗口(如 568×320)进一步压缩 */
@media (max-height: 380px) {
	.card.compact {
		padding: 0.55rem 0.9rem;
		gap: 0.7rem;
	}

	.card.compact .result-left {
		padding-right: 0.7rem;
		gap: 0.1rem;
	}

	.card.compact .trophy {
		width: 2.2rem;
		height: 2.2rem;
	}

	.card.compact h1 {
		font-size: 1.2rem;
	}

	.card.compact .sub {
		font-size: 0.68rem;
	}

	.card.compact .score-item b {
		font-size: 1.5rem;
	}

	.card.compact .again {
		padding: 0.45rem;
		font-size: 0.85rem;
	}

	.card.compact .menu {
		padding: 0.25rem;
		font-size: 0.72rem;
	}

	.card.compact .again-hint,
	.card.compact .replay-note {
		font-size: 0.56rem;
	}
}

@media (max-width: 600px) {
	.content {
		padding: 0.5rem;
	}

	.card {
		padding: 1.1rem 0.9rem;
		gap: 0.5rem;
		border-radius: 0.9rem;
	}

	.trophy {
		width: 3rem;
		height: 3rem;
	}

	.card h1 {
		font-size: 1.4rem;
	}

	.sub {
		font-size: 0.8rem;
	}

	.score-item b {
		font-size: 1.7rem;
	}

	.again {
		padding: 0.7rem;
		font-size: 0.9rem;
	}
}
</style>
