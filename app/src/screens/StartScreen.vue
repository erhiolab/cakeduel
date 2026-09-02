<script setup lang="ts">
import {computed, ref, watch, onMounted, onUnmounted} from "vue"
import {game} from "../composables/useGame"
import {audioState, playSfx, toggleBgm} from "../game/audio"
import {appConfig} from "../config"
import HelpOverlay from "../components/HelpOverlay.vue"

// 房间状态
const {state, createRoom, joinRoom, leave, clearMessage} = game

// 玩家昵称(本地缓存, 避免每次重新输入)
const NAME_KEY = "cakeduel_name"
const name = ref(localStorage.getItem(NAME_KEY) ?? "")

// 昵称变更时写入本地缓存
watch(name, (v) => {
	if (v) localStorage.setItem(NAME_KEY, v)
})

// 视图状态
const view = ref<"menu" | "create" | "join">("menu")

// 房间码状态
const code = ref("")

// 显示帮助状态
const showHelp = ref(false)

// 匹配状态
const busy = ref(false)

// 匹配等待时间
const matchElapsed = ref(0)

// 匹配定时器
let matchTimer: number | undefined

// 服务器健康检查地址(与 WS 同源/同后端)
const PING_BASE = (() => {
	const CFG = appConfig.wsUrl
	if (CFG) {
		return CFG.replace(/^wss:\/\//, "https://").replace(/^ws:\/\//, "http://").replace(/\/+$/, "")
	}
	if (import.meta.env.DEV) return "http://127.0.0.1:8080"
	return window.location.origin
})()

// 服务器是否异常
const serverDown = ref(false)

// 健康检查定时器
let pingTimer: number | undefined

// 检查服务器状态(3 秒超时)
const checkServer = async () => {
	const CONTROLLER = new AbortController()
	const TIMER = window.setTimeout(() => CONTROLLER.abort(), 3000)
	try {
		const RES = await fetch(`${PING_BASE}/ping`, {cache: "no-store", signal: CONTROLLER.signal})
		serverDown.value = !RES.ok
	} catch {
		serverDown.value = true
	} finally {
		window.clearTimeout(TIMER)
	}
}

onMounted(() => {
	checkServer()
	pingTimer = window.setInterval(checkServer, 15000)
})

watch(() => state.matching, (v) => {
	if (matchTimer) window.clearInterval(matchTimer)
	if (v) {
		matchElapsed.value = 0
		matchTimer = window.setInterval(() => {
			matchElapsed.value++
		}, 1000)
	}
})

// 取消匹配
const cancelMatch = () => {
	leave()
	clearMessage()
}

// 消息状态
const message = computed(() => state.message)

// 切换视图
const go = (viewName: "create" | "join") => {
	playSfx("hoof")
	view.value = viewName
}

// 返回主菜单
const back = () => {
	view.value = "menu"
}

// 创建房间
const doCreate = (mode: "private" | "random") => {
	playSfx("hoof")
	clearMessage()
	busy.value = true
	createRoom(name.value.trim() || "神秘玩家", mode)
	window.setTimeout(() => (busy.value = false), 800)
}

// 加入房间
const doJoin = () => {
	playSfx("hoof")
	clearMessage()
	const C = code.value.trim().toUpperCase()
	if (!C) return
	busy.value = true
	joinRoom(C, name.value.trim() || "神秘玩家")
	window.setTimeout(() => (busy.value = false), 800)
}

onUnmounted(() => {
	if (matchTimer) window.clearInterval(matchTimer)
	if (pingTimer) window.clearInterval(pingTimer)
})
</script>

<template>
	<div class="start" data-cakeduel-screen="start">
		<img class="bg" src="/cakeduel/playmat.jpg" alt="" draggable="false"/>
		<div class="overlay"></div>
		<div class="top-btns">
			<button class="music-btn" :class="{ muted: !audioState.bgmOn }" title="背景音乐开关" @click="toggleBgm">
				<svg v-if="audioState.bgmOn" viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
					<path d="M9 18V5l12-2v13"/>
					<circle cx="6" cy="18" r="3"/>
					<circle cx="18" cy="16" r="3"/>
				</svg>
				<svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor"
					 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M9 18V5l12-2v13"/>
					<circle cx="6" cy="18" r="3"/>
					<path d="M16 8l5 5M21 8l-5 5"/>
				</svg>
			</button>
			<button class="help-btn" @click="showHelp = true">?</button>
		</div>
		<div class="content">
			<div class="title-block">
				<h1 class="title-font">蛋糕对决</h1>
				<p class="subtitle">绵羊将军，部署你的部队！</p>
			</div>
			<div class="panel glass">
				<div v-if="state.matching" class="menu matching">
					<div class="match-anim">
						<span></span><span></span><span></span>
					</div>
					<h2 class="panel-title">正在匹配对手…</h2>
					<p class="match-wait">已等待 {{ matchElapsed }} 秒</p>
					<button class="ghost-btn" @click="cancelMatch">取消匹配</button>
				</div>
				<template v-else>
					<div v-if="view === 'menu'" key="menu" class="menu">
						<label class="field">
							<span>你的昵称</span>
							<input v-model="name" maxlength="12" placeholder="神秘玩家"/>
						</label>
						<button class="main-btn" @click="go('create')">
							<span>创建房间</span>
							<small>私有房间或随机匹配</small>
						</button>
						<button class="main-btn secondary" @click="go('join')">
							<span>加入房间</span>
							<small>输入房间码加入好友</small>
						</button>
						<p v-if="message" class="message">{{ message }}</p>
					</div>
					<div v-else-if="view === 'create'" key="create" class="menu">
						<h2 class="panel-title">创建房间</h2>
						<button class="main-btn" :disabled="busy" @click="doCreate('private')">
							<span>创建私有房间</span>
							<small>生成房间码，好友输入即可加入</small>
						</button>
						<button class="main-btn accent" :disabled="busy" @click="doCreate('random')">
							<span>随机匹配</span>
							<small>匹配同样选择了随机的玩家</small>
						</button>
						<button class="ghost-btn" @click="back">返回</button>
					</div>
					<div v-else key="join" class="menu">
						<h2 class="panel-title">加入房间</h2>
						<label class="field">
							<span>房间码</span>
							<input
								v-model="code"
								class="code-input"
								maxlength="8"
								placeholder="例如 ABC123"
								@keyup.enter="doJoin"
							/>
						</label>
						<button class="main-btn" :disabled="busy || !code.trim()" @click="doJoin">
							<span>加入</span>
						</button>
						<button class="ghost-btn" @click="back">返回</button>
					</div>
				</template>
			</div>
		</div>
		<p v-if="serverDown" class="server-warning">服务器状态异常, 请联系管理员</p>
		<HelpOverlay :open="showHelp" @close="showHelp = false"/>
	</div>
</template>

<style scoped>
.start {
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
	background: radial-gradient(ellipse 70% 60% at 50% 45%, rgba(130, 160, 190, 0.25) 0%, transparent 70%),
	linear-gradient(180deg, rgba(70, 95, 120, 0.55) 0%, rgba(80, 105, 125, 0.4) 40%, rgba(140, 100, 70, 0.25) 100%);
	box-shadow: inset 0 0 6rem 1.5rem rgba(30, 45, 60, 0.6);
}

.top-btns {
	position: absolute;
	top: 1.2rem;
	right: 1.2rem;
	z-index: 20;
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.help-btn,
.music-btn {
	width: 2.2rem;
	height: 2.2rem;
	border-radius: 50%;
	background: rgba(0, 0, 0, 0.35);
	border: 1.5px solid rgba(240, 200, 120, 0.45);
	color: #f5dfae;
	font-weight: 800;
	font-size: 1rem;
	box-shadow: 0 2px 12px rgba(0, 0, 0, 0.4);
	transition: transform 0.15s, box-shadow 0.2s;
	display: flex;
	align-items: center;
	justify-content: center;
}

.help-btn:hover,
.music-btn:hover {
	transform: scale(1.08);
	box-shadow: 0 0 16px rgba(240, 200, 120, 0.5);
}

.music-btn.muted {
	color: rgba(245, 223, 174, 0.45);
	border-color: rgba(240, 200, 120, 0.2);
}

.content {
	position: relative;
	z-index: 10;
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 1rem;
	max-width: 22rem;
	width: 100%;
}

.title-block {
	text-align: center;
	margin-bottom: 1.4rem;
	animation: rise-in 0.5s ease both;
}

.title-block h1 {
	font-size: 2.4rem;
	font-weight: 900;
	letter-spacing: 0.02em;
	background: linear-gradient(180deg, #fff7e6 0%, #f5c54a 45%, #e8956a 75%, #c97050 100%);
	-webkit-background-clip: text;
	background-clip: text;
	-webkit-text-fill-color: transparent;
	filter: drop-shadow(0 2px 0 #c97050) drop-shadow(0 4px 10px rgba(60, 40, 20, 0.55));
}

.subtitle {
	margin-top: 0.4rem;
	font-size: 0.9rem;
	color: rgba(255, 240, 210, 0.85);
	font-weight: 600;
}

.panel {
	width: 100%;
	border-radius: 1.1rem;
	padding: 1.1rem;
	animation: rise-in 0.5s 0.15s ease both;
}

.menu {
	display: flex;
	flex-direction: column;
	gap: 0.7rem;
}

.panel-title {
	text-align: center;
	font-size: 1.05rem;
	font-weight: 800;
	color: #3a2c1f;
	margin-bottom: 0.2rem;
}

.field {
	display: flex;
	flex-direction: column;
	gap: 0.3rem;
	font-size: 0.72rem;
	font-weight: 800;
	color: #6b5438;
	letter-spacing: 0.04em;
}

.field input {
	border-radius: 0.7rem;
	border: 1.5px solid rgba(107, 84, 56, 0.25);
	background: rgba(255, 255, 255, 0.6);
	padding: 0.6rem 0.8rem;
	font-size: 0.95rem;
	font-weight: 600;
	color: #3a2c1f;
	outline: none;
	transition: border-color 0.2s, box-shadow 0.2s;
}

.field input:focus {
	border-color: rgba(245, 197, 24, 0.8);
	box-shadow: 0 0 0 3px rgba(245, 197, 24, 0.15);
}

.code-input {
	text-transform: uppercase;
	letter-spacing: 0.25em;
	text-align: center;
	font-weight: 900;
}

.main-btn {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.15rem;
	border-radius: 0.85rem;
	padding: 0.8rem 1rem;
	background: linear-gradient(135deg, #7ab55c, #5a9a40 50%, #4a5a40 100%);
	border: 2px solid #7ab55c;
	color: #fdf6e9;
	font-weight: 800;
	font-size: 1rem;
	box-shadow: 0 6px 24px rgba(90, 154, 64, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.25);
	transition: transform 0.15s, box-shadow 0.2s;
}

.main-btn small {
	font-size: 0.68rem;
	font-weight: 600;
	opacity: 0.8;
}

.main-btn:hover:not(:disabled) {
	transform: scale(1.02);
	box-shadow: 0 8px 32px rgba(90, 154, 64, 0.45), inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.main-btn:active:not(:disabled) {
	transform: scale(0.98);
}

.main-btn:disabled {
	opacity: 0.6;
}

.main-btn.secondary {
	background: rgba(255, 255, 255, 0.3);
	border-color: rgba(107, 84, 56, 0.35);
	color: #3a2c1f;
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5), 0 2px 8px rgba(0, 0, 0, 0.1);
}

.main-btn.accent {
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	border-color: #f5c54a;
	color: #3a2c1f;
	box-shadow: 0 6px 24px rgba(232, 162, 58, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.ghost-btn {
	padding: 0.45rem;
	border-radius: 0.7rem;
	font-size: 0.85rem;
	font-weight: 700;
	color: #6b5438;
	transition: background 0.2s;
}

.ghost-btn:hover {
	background: rgba(0, 0, 0, 0.05);
}

.message {
	text-align: center;
	font-size: 0.8rem;
	font-weight: 700;
	color: #b45309;
}

.server-warning {
	position: absolute;
	bottom: 0.8rem;
	left: 50%;
	transform: translateX(-50%);
	z-index: 30;
	padding: 0.5rem 1rem;
	border-radius: 0.6rem;
	background: rgba(127, 29, 29, 0.9);
	border: 1px solid rgba(248, 113, 113, 0.5);
	color: #fecaca;
	font-size: 0.85rem;
	font-weight: 800;
	text-align: center;
	animation: pulse-glow 1.6s ease-in-out infinite;
}

.matching {
	align-items: center;
	padding: 1rem 0 0.4rem;
}

.match-anim {
	display: flex;
	gap: 0.5rem;
	margin-bottom: 0.8rem;
}

.match-anim span {
	width: 0.7rem;
	height: 0.7rem;
	border-radius: 50%;
	background: #d97706;
	animation: match-bounce 1s ease-in-out infinite;
}

.match-anim span:nth-child(2) {
	animation-delay: 0.15s;
}

.match-anim span:nth-child(3) {
	animation-delay: 0.3s;
}

@keyframes match-bounce {
	0%,
	100% {
		transform: translateY(0);
		opacity: 0.5;
	}
	50% {
		transform: translateY(-0.5rem);
		opacity: 1;
	}
}

.match-wait {
	font-size: 0.85rem;
	color: #9a7a55;
	font-weight: 700;
	margin-bottom: 0.4rem;
}

@media (max-height: 560px) and (min-width: 640px) {
	.content {
		flex-direction: row;
		align-items: center;
		justify-content: center;
		gap: 2.5rem;
		max-width: 52rem;
		padding: 0.5rem 1.5rem;
	}

	.title-block {
		flex: 1;
		text-align: left;
		margin-bottom: 0;
		padding-left: 1rem;
	}

	.title-block h1 {
		font-size: clamp(2rem, 6vh, 3.2rem);
	}

	.subtitle {
		font-size: 0.8rem;
		margin-top: 0.25rem;
	}

	.panel {
		flex: 1;
		max-width: 21rem;
		padding: 0.9rem;
	}

	.main-btn {
		padding: 0.55rem 1rem;
		font-size: 0.95rem;
	}

	.main-btn small {
		font-size: 0.62rem;
	}

	.menu {
		gap: 0.55rem;
	}
}
</style>
