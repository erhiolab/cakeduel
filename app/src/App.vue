<script setup lang="ts">
import {onMounted, ref} from "vue"
import {game} from "./composables/useGame"
import {enableAudio} from "./game/audio"
import {preloadAssets, shouldPreloadAssets} from "./game/assets"
import ReplayViewer from "./components/ReplayViewer.vue"
import type {ReplayData} from "./composables/useGame"
import StartScreen from "./screens/StartScreen.vue"
import LobbyScreen from "./screens/LobbyScreen.vue"
import GameScreen from "./screens/GameScreen.vue"
import ResultsScreen from "./screens/ResultsScreen.vue"
import SpectateScreen from "./screens/SpectateScreen.vue"
import LandscapePrompt from "./components/LandscapePrompt.vue"

const {state} = game

// 首启预缓存状态
const caching = ref(false)

// 缓存进度
const cacheDone = ref(0)

// 缓存总数
const cacheTotal = ref(0)

// 分享回放数据(/replay/:id)
const sharedReplay = ref<ReplayData | null>(null)

// 分享回放加载错误
const sharedError = ref("")

const handleFirstClick = () => {
	enableAudio()
}

// 关闭分享回放, 回到主菜单
const closeShared = () => {
	sharedReplay.value = null
	sharedError.value = ""
	window.history.replaceState(null, "", "/")
}

onMounted(() => {
	window.addEventListener("pointerdown", handleFirstClick, {once: true})
	// 注册资源缓存 Service Worker(生产/后端托管环境)
	if (!import.meta.env.DEV && "serviceWorker" in navigator) {
		void navigator.serviceWorker.register("/sw.js").catch(() => {})
	}
	// 首次启动(或资源版本升级)时预缓存卡片/音频/背景
	if (shouldPreloadAssets()) {
		caching.value = true
		void preloadAssets((done, total) => {
			cacheDone.value = done
			cacheTotal.value = total
		}).finally(() => {
			caching.value = false
		})
	}
	// 仅当此前处于房间/对局中(刷新恢复)时才自动连接, 避免主界面挂空闲连接
	if (sessionStorage.getItem("cakeduel_resume")) {
		game.connect()
	}
	// 分享回放链接: /replay/{id}
	const SHARE_MATCH = window.location.pathname.match(/^\/replay\/([A-Za-z0-9]{6,32})\/?$/)
	if (SHARE_MATCH) {
		const API_BASE = import.meta.env.DEV ? "http://127.0.0.1:8080" : window.location.origin
		fetch(`${API_BASE}/api/replay/${SHARE_MATCH[1]}`)
			.then(async (res) => {
				if (!res.ok) throw new Error("回放不存在或已过期")
				const DATA = await res.json()
				if (!DATA || !Array.isArray(DATA.frames)) throw new Error("回放数据无效")
				sharedReplay.value = DATA as ReplayData
			})
			.catch((e) => {
				sharedError.value = e.message || "加载失败"
			})
	}
})
</script>

<template>
	<div class="app-root" @pointerdown="handleFirstClick" @contextmenu.prevent>
		<StartScreen v-if="state.screen === 'start'" key="start"/>
		<LobbyScreen v-else-if="state.screen === 'lobby'" key="lobby"/>
		<GameScreen v-else-if="state.screen === 'game'" key="game"/>
		<ResultsScreen v-else-if="state.screen === 'results'" key="results"/>
		<SpectateScreen v-else-if="state.screen === 'spectate'" key="spectate"/>
		<LandscapePrompt/>
		<ReplayViewer
			v-if="sharedReplay"
			:replay="sharedReplay"
			@close="closeShared"
		/>
		<Transition name="fade">
			<div v-if="sharedError" class="shared-error">
				<div class="shared-error-box glass">
					<h2>😢 {{ sharedError }}</h2>
					<p>分享链接 24 小时内有效，过期或已删除则无法查看</p>
					<button class="shared-back" @click="closeShared">返回主菜单</button>
				</div>
			</div>
		</Transition>
		<Transition name="fade">
			<div v-if="caching" class="cache-overlay">
				<div class="cache-box glass">
					<div class="cache-icon">
						<span class="cache-slice"></span>
					</div>
					<h2 class="title-font">正在缓存资源</h2>
					<p>卡片、音效与背景首次加载需要一点时间，请稍候</p>
					<div class="cache-bar">
						<div class="cache-fill" :style="{ width: `${cacheTotal > 0 ? Math.round((cacheDone / cacheTotal) * 100) : 0}%` }"></div>
					</div>
					<span class="cache-count">{{ cacheDone }} / {{ cacheTotal }}</span>
				</div>
			</div>
		</Transition>
	</div>
</template>

<style scoped>
.app-root {
	width: 100%;
	height: 100%;
	position: relative;
	overflow: hidden;
}

.cache-overlay {
	position: fixed;
	inset: 0;
	z-index: 10010;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1.5rem;
	background: radial-gradient(ellipse at center, rgba(34, 45, 58, 0.92), rgba(14, 19, 27, 0.96));
	backdrop-filter: blur(6px);
}

.cache-box {
	max-width: 24rem;
	width: 100%;
	border-radius: 1.2rem;
	padding: 2rem 1.6rem;
	text-align: center;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.7rem;
}

.cache-icon {
	position: relative;
	width: 3.4rem;
	height: 3.4rem;
	border-radius: 50%;
	background: linear-gradient(135deg, #f5c54a, #e8956a);
	display: flex;
	align-items: center;
	justify-content: center;
	animation: cache-bounce 1.1s ease-in-out infinite;
}

.cache-slice {
	width: 1.5rem;
	height: 1.5rem;
	border-radius: 50% 50% 50% 0;
	background: #fff7e6;
	transform: rotate(-45deg);
}

.cache-box h2 {
	font-size: 1.45rem;
	color: #6b4a2b;
}

.cache-box p {
	font-size: 0.82rem;
	color: #9a7a55;
	font-weight: 600;
}

.cache-bar {
	width: 100%;
	height: 0.55rem;
	border-radius: 2rem;
	background: rgba(107, 84, 56, 0.18);
	overflow: hidden;
	margin-top: 0.3rem;
}

.cache-fill {
	height: 100%;
	border-radius: 2rem;
	background: linear-gradient(90deg, #7ab55c, #e8a23a);
	transition: width 0.25s ease;
}

.cache-count {
	font-size: 0.75rem;
	font-weight: 800;
	color: #b45309;
}

.shared-error {
	position: fixed;
	inset: 0;
	z-index: 10020;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1.5rem;
	background: rgba(20, 25, 35, 0.6);
	backdrop-filter: blur(5px);
}

.shared-error-box {
	max-width: 24rem;
	width: 100%;
	border-radius: 1rem;
	padding: 1.5rem;
	text-align: center;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.7rem;
}

.shared-error-box h2 {
	font-size: 1.2rem;
	color: #3a2c1f;
}

.shared-error-box p {
	font-size: 0.78rem;
	color: #9a7a55;
	font-weight: 600;
}

.shared-back {
	width: 100%;
	padding: 0.6rem;
	border-radius: 0.7rem;
	font-size: 0.9rem;
	font-weight: 800;
	color: #fdf6e9;
	background: linear-gradient(135deg, #d97706, #b45309);
}

@keyframes cache-bounce {
	0%,
	100% {
		transform: translateY(0) scale(1);
	}
	50% {
		transform: translateY(-0.4rem) scale(1.05);
	}
}

</style>
