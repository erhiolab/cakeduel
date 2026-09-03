<script setup lang="ts">
import {computed, ref, watch, onMounted, onUnmounted} from "vue"
import {game} from "../composables/useGame"
import {audioState, playSfx, toggleBgm} from "../game/audio"
import {appConfig} from "../config"
import {CARDS, DEFAULT_DECK, SPECIAL_CARD_NAMES} from "../game/cards"
import {deleteSavedReplay, loadSavedReplays, type PlayerInfo, type ReplayData} from "../composables/useGame"
import HelpOverlay from "../components/HelpOverlay.vue"
import AboutOverlay from "../components/AboutOverlay.vue"
import ReplayViewer from "../components/ReplayViewer.vue"

// 房间状态
const {state, createRoom, joinRoom, cancelSpectate, leave, clearMessage, clearError} = game

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

// 显示关于信息
const showAbout = ref(false)

// 回放列表弹层
const showReplays = ref(false)

// 回放列表数据
const replays = ref<ReplayData[]>([])

// 正在播放的回放
const activeReplay = ref<ReplayData | null>(null)

// 分享生成中的回放 startedAt
const sharingId = ref<number | null>(null)

// 最近一次分享信息
const shareInfo = ref<{id: string; url: string} | null>(null)

// 公开房间列表类型
interface PublicRoom {
	code: string
	mode: string
	status: string
	phase?: string
	paused: boolean
	gameOver: boolean
	players: PlayerInfo[]
	spectatorCount: number
}

// 房间列表弹层
const showRooms = ref(false)

// 房间列表数据
const rooms = ref<PublicRoom[]>([])

// 房间列表自动刷新定时器
let roomsTimer: number | undefined

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
	// 随机匹配使用经典卡组; 私有房间使用用户自定义(默认经典)
	const DECK = mode === "private" && deckCustom.value ? {...deckCounts.value} : undefined
	createRoom(name.value.trim() || "神秘玩家", mode, DECK)
	window.setTimeout(() => (busy.value = false), 800)
}

// 是否展开卡组配置
const deckCustom = ref(false)

// 当前编辑的特殊卡数量
const deckCounts = ref<Record<string, number>>({...DEFAULT_DECK})

// 是否全部为 0
const deckZero = computed(() => SPECIAL_CARD_NAMES.every((n) => (deckCounts.value[n] ?? 0) === 0))

// 配置摘要(用于按钮副标题)
const deckSummary = computed(() => {
	if (deckZero.value) return "不使用特殊卡"
	return SPECIAL_CARD_NAMES.filter((n) => (deckCounts.value[n] ?? 0) > 0)
		.map((n) => `${CARDS[n]?.name ?? n}×${deckCounts.value[n]}`)
		.join(" ")
})

// 展开/收起卡组配置
const toggleDeck = () => {
	deckCustom.value = !deckCustom.value
}

// 调整某张特殊卡数量(0-3)
const changeDeck = (name: string, delta: number) => {
	const NEXT = {...deckCounts.value}
	const V = Math.max(0, Math.min(3, (NEXT[name] ?? 0) + delta))
	NEXT[name] = V
	deckCounts.value = NEXT
}

// 恢复经典配置
const resetDeck = () => {
	deckCounts.value = {...DEFAULT_DECK}
}

// 全部置 0(不使用特殊卡)
const zeroDeck = () => {
	const NEXT: Record<string, number> = {}
	for (const n of SPECIAL_CARD_NAMES) NEXT[n] = 0
	deckCounts.value = NEXT
}

// 加入房间
const doJoin = (as: "player" | "spectator" = "player") => {
	playSfx("hoof")
	clearMessage()
	const C = code.value.trim().toUpperCase()
	if (!C) return
	busy.value = true
	joinRoom(C, name.value.trim() || "神秘玩家", as)
	window.setTimeout(() => (busy.value = false), 800)
}

// 加入失败自动返回主菜单的定时器
let joinErrorTimer: number | undefined

// 房间不存在/已满/已开局: 提示后返回主菜单
watch(
	() => state.error,
	(v) => {
		if (!v || view.value !== "join") return
		if (!/不存在|已满|只能观战/.test(v)) return
		if (joinErrorTimer) window.clearTimeout(joinErrorTimer)
		joinErrorTimer = window.setTimeout(() => {
			clearError()
			view.value = "menu"
		}, 1600)
	},
)

onUnmounted(() => {
	if (matchTimer) window.clearInterval(matchTimer)
	if (pingTimer) window.clearInterval(pingTimer)
	if (roomsTimer) window.clearInterval(roomsTimer)
})

// 打开房间列表并开始自动刷新
const openRooms = async () => {
	playSfx("hoof")
	showRooms.value = true
	await refreshRooms()
	if (roomsTimer) window.clearInterval(roomsTimer)
	roomsTimer = window.setInterval(refreshRooms, 5000)
}

// 关闭房间列表
const closeRooms = () => {
	showRooms.value = false
	if (roomsTimer) window.clearInterval(roomsTimer)
	roomsTimer = undefined
}

// 刷新公开房间列表
const refreshRooms = async () => {
	try {
		const RES = await fetch(`${PING_BASE}/api/rooms`, {cache: "no-store"})
		if (!RES.ok) return
		const DATA = await RES.json()
		const ORDER = {waiting: 0, playing: 1, finished: 2} as Record<string, number>
		rooms.value = (DATA.body?.rooms || []).sort((a: PublicRoom, b: PublicRoom) => {
			return (ORDER[a.status] ?? 9) - (ORDER[b.status] ?? 9) || a.code.localeCompare(b.code)
		})
	} catch {
		// 后端暂时不可达时保留旧列表
	}
}

// 一键进入观战
const enterSpectate = (room: PublicRoom) => {
	playSfx("hoof")
	joinRoom(room.code, name.value.trim() || "神秘玩家", "spectator")
	closeRooms()
}

// 房间状态文案
const roomStatusText = (room: PublicRoom): string => {
	if (room.gameOver) return "已结束"
	if (room.status === "playing") return room.paused ? "暂停中" : "对局中"
	return "等待开局"
}

// 玩家名摘要
const roomPlayersText = (room: PublicRoom): string => {
	const NAMES = room.players?.filter((p) => p.name).map((p) => p.name) || []
	if (NAMES.length === 0) return "等待玩家加入…"
	return NAMES.join(" vs ")
}

// 打开回放列表
const openReplays = () => {
	playSfx("hoof")
	replays.value = loadSavedReplays()
	showReplays.value = true
}

// 关闭回放列表
const closeReplays = () => {
	showReplays.value = false
	activeReplay.value = null
}

// 删除一份回放
const removeReplay = (replay: ReplayData) => {
	deleteSavedReplay(replay.startedAt)
	replays.value = loadSavedReplays()
}

// 播放回放
const watchReplay = (replay: ReplayData) => {
	playSfx("hoof")
	activeReplay.value = replay
}

// 分享回放(后端存 Redis, 24 小时有效)
const shareReplay = async (replay: ReplayData) => {
	if (sharingId.value != null) return
	sharingId.value = replay.startedAt
	try {
		const RES = await fetch(`${PING_BASE}/api/replay/share`, {
			method: "POST",
			headers: {"Content-Type": "application/json"},
			body: JSON.stringify(replay),
		})
		const DATA = await RES.json()
		if (!RES.ok || DATA.error) throw new Error(DATA.message || "分享失败")
		const URL = DATA.body.url.startsWith("http") ? DATA.body.url : `${PING_BASE}${DATA.body.url}`
		shareInfo.value = {id: DATA.body.id, url: URL}
		playSfx("hoof")
	} catch {
		shareInfo.value = null
	} finally {
		sharingId.value = null
	}
}

// 复制分享链接
const copyShareLink = () => {
	if (!shareInfo.value) return
	playSfx("hoof")
	navigator.clipboard?.writeText(shareInfo.value.url).catch(() => {})
}

// 回放耗时文本
const replayDuration = (replay: ReplayData): string => {
	const SECONDS = Math.max(1, Math.round((replay.durationMs ?? 0) / 1000))
	return SECONDS >= 60 ? `${Math.floor(SECONDS / 60)}分${SECONDS % 60}秒` : `${SECONDS}秒`
}

// 回放时间文本
const replayTime = (replay: ReplayData): string => {
	if (!replay.startedAt) return ""
	return new Date(replay.startedAt).toLocaleString("zh-CN", {
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
	})
}
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
			<button class="about-btn" title="关于 / 仓库 / 版权" @click="showAbout = true">i</button>
			<button class="help-btn" @click="showHelp = true">?</button>
		</div>
		<div class="content">
			<div class="title-block">
				<h1 class="title-font">蛋糕对决</h1>
				<p class="subtitle">绵羊将军，部署你的部队！</p>
			</div>
	<div class="panel glass">
		<div v-if="state.matching || state.spectateWait" class="menu matching">
			<div class="match-anim">
				<span></span><span></span><span></span>
			</div>
			<h2 v-if="state.matching" class="panel-title">正在匹配对手…</h2>
			<h2 v-else class="panel-title">观战等待中</h2>
			<p v-if="state.matching" class="match-wait">已等待 {{ matchElapsed }} 秒</p>
			<p v-else class="match-wait">房间 {{ state.roomCode }} · {{ state.message }}</p>
			<button class="ghost-btn" @click="state.matching ? cancelMatch() : cancelSpectate()">
				{{ state.matching ? "取消匹配" : "取消观战" }}
			</button>
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
						<button class="replay-link" @click="openRooms">
							<span>🏠 房间列表</span>
							<small>查看当前房间，一键进入观战</small>
						</button>
						<button class="replay-link" @click="openReplays">
							<span>🎞 回放录像</span>
							<small>查看本地保存的最近对局</small>
						</button>
						<p v-if="message" class="message">{{ message }}</p>
						<p v-if="state.error" class="message error-text">{{ state.error }}</p>
					</div>
					<div v-else-if="view === 'create'" key="create" class="menu">
						<h2 class="panel-title">创建房间</h2>
						<p v-if="state.error" class="message error-text">{{ state.error }}</p>
						<button class="main-btn" :disabled="busy" @click="doCreate('private')">
							<span>创建私有房间</span>
							<small>{{ deckCustom ? `自定义卡组: ${deckSummary}` : "生成房间码，好友输入即可加入" }}</small>
						</button>
						<button class="main-btn accent" :disabled="busy" @click="doCreate('random')">
							<span>随机匹配</span>
							<small>匹配同样选择了随机的玩家（经典卡组）</small>
						</button>

						<button class="ghost-btn deck-toggle" @click="toggleDeck">
							{{ deckCustom ? "收起卡组配置" : "自定义卡组（特殊卡数量）" }}
						</button>
						<div v-if="deckCustom" class="deck-editor">
							<div v-for="name in SPECIAL_CARD_NAMES" :key="name" class="deck-row">
								<span class="deck-name">{{ CARDS[name]?.name ?? name }}</span>
								<div class="deck-stepper">
									<button class="step-btn" @click="changeDeck(name, -1)">−</button>
									<span class="step-val">{{ deckCounts[name] ?? 0 }}</span>
									<button class="step-btn" @click="changeDeck(name, 1)">＋</button>
								</div>
							</div>
							<p v-if="deckZero" class="deck-zero">全部为 0：本局不使用特殊卡</p>
							<div class="deck-presets">
								<button class="preset-btn" @click="resetDeck">经典</button>
								<button class="preset-btn" @click="zeroDeck">无特殊卡</button>
							</div>
						</div>

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
								@keyup.enter="doJoin('player')"
							/>
						</label>
						<button class="main-btn" :disabled="busy || !code.trim()" @click="doJoin('player')">
							<span>加入房间</span>
							<small>成为玩家，满员/已开局会提示</small>
						</button>
						<button class="main-btn accent" :disabled="busy || !code.trim()" @click="doJoin('spectator')">
							<span>观战</span>
							<small>已开局直接观战，未开局自动等待</small>
						</button>
						<p v-if="state.error" class="join-error">{{ state.error }}</p>
						<button class="ghost-btn" @click="back">返回</button>
					</div>
				</template>
			</div>
		</div>
		<p v-if="serverDown" class="server-warning">服务器状态异常, 请联系管理员</p>
		<HelpOverlay :open="showHelp" @close="showHelp = false"/>
		<AboutOverlay :open="showAbout" @close="showAbout = false"/>

		<Transition name="fade">
			<div v-if="showRooms" class="rooms-overlay" data-cakeduel-screen="rooms">
				<div class="rooms-panel glass">
					<div class="rooms-head">
						<h2>🏠 房间列表</h2>
						<div class="rooms-head-actions">
							<button class="refresh-btn" @click="refreshRooms">刷新</button>
							<button class="close-btn" @click="closeRooms">✕</button>
						</div>
					</div>
					<p class="rooms-tip">点击「观战」直接进入；房间未开局会自动等待，开局后进入观战</p>
					<div v-if="rooms.length === 0" class="rooms-empty">
						<p>暂无房间</p>
						<small>房间会随对局结束自动移除，稍后刷新看看</small>
					</div>
					<div v-else class="rooms-list">
						<div v-for="room in rooms" :key="room.code" class="rooms-item">
							<div class="rooms-info">
								<span class="rooms-code">{{ room.code }}</span>
								<span class="rooms-sub">
									<span class="pill" :class="room.status">{{ roomStatusText(room) }}</span>
									<span class="players-text">{{ roomPlayersText(room) }}</span>
									<span v-if="room.spectatorCount > 0" class="viewers">👁 {{ room.spectatorCount }}</span>
								</span>
							</div>
							<button class="spectate-btn" @click="enterSpectate(room)">进入观战</button>
						</div>
					</div>
				</div>
			</div>
		</Transition>

		<Transition name="fade">
			<div v-if="showReplays && !activeReplay" class="replay-overlay" data-cakeduel-screen="replays">
				<div class="replay-panel glass">
					<div class="replay-head">
						<h2>🎞 回放录像</h2>
						<button class="close-btn" @click="closeReplays">✕</button>
					</div>
					<p class="replay-tip">对局结束后回放会自动保存到本地（最多 10 场）</p>
					<div v-if="shareInfo" class="share-box">
						<p class="share-label">🔗 分享链接已生成（24 小时内有效）</p>
						<div class="share-link">{{ shareInfo.url }}</div>
						<div class="share-actions">
							<button class="copy-share" @click="copyShareLink">复制链接</button>
							<button class="close-share" @click="shareInfo = null">收起</button>
						</div>
					</div>
					<div v-if="replays.length === 0" class="empty">
						<p>暂无回放</p>
						<small>打完一局后会自动保存，可随时回来复盘</small>
					</div>
					<div v-else class="replay-list">
						<div v-for="replay in replays" :key="replay.startedAt" class="replay-item">
							<div class="replay-info">
								<span class="names">{{ replay.playerNames?.[0] || "玩家A" }} vs {{ replay.playerNames?.[1] || "玩家B" }}</span>
								<span class="sub">
									{{ replayTime(replay) }} · 用时 {{ replayDuration(replay) }} ·
									🏆 {{ replay.playerNames?.[replay.winner] || `玩家 ${(replay.winner ?? 0) + 1}` }} 获胜
								</span>
							</div>
							<div class="replay-actions">
								<button class="watch-btn" @click="watchReplay(replay)">观看</button>
								<button class="share-btn" :disabled="sharingId != null" @click="shareReplay(replay)">
									{{ sharingId === replay.startedAt ? "分享中…" : "分享" }}
								</button>
								<button class="del-btn" title="删除" @click="removeReplay(replay)">🗑</button>
							</div>
						</div>
					</div>
				</div>
			</div>
		</Transition>

		<ReplayViewer
			v-if="activeReplay"
			:replay="activeReplay"
			@close="closeReplays"
		/>
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
.music-btn,
.about-btn {
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
.music-btn:hover,
.about-btn:hover {
	transform: scale(1.08);
	box-shadow: 0 0 16px rgba(240, 200, 120, 0.5);
}

.about-btn {
	font-family: Georgia, "Times New Roman", serif;
	font-style: italic;
	font-weight: 800;
	font-size: 1.05rem;
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

.replay-link {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 0.1rem;
	padding: 0.55rem;
	border-radius: 0.7rem;
	font-size: 0.88rem;
	font-weight: 800;
	color: #6b4a2b;
	background: rgba(255, 255, 255, 0.42);
	border: 1px solid rgba(107, 84, 56, 0.3);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5);
	transition: background 0.2s, transform 0.15s;
}

.replay-link:hover {
	background: rgba(255, 255, 255, 0.62);
	transform: scale(1.01);
}

.replay-link small {
	font-size: 0.62rem;
	font-weight: 600;
	color: #9a7a55;
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

.join-error {
	text-align: center;
	font-size: 0.78rem;
	font-weight: 700;
	color: #dc2626;
}

.error-text {
	color: #dc2626;
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

.replay-overlay {
	position: fixed;
	inset: 0;
	z-index: 120;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1rem;
	background: rgba(12, 18, 26, 0.72);
	backdrop-filter: blur(5px);
}

.replay-panel {
	width: 100%;
	max-width: 30rem;
	max-height: min(84vh, 40rem);
	display: flex;
	flex-direction: column;
	border-radius: 1.1rem;
	padding: 1.1rem 1.2rem;
	overflow: hidden;
	animation: rise-in 0.3s ease both;
}

.replay-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
}

.replay-head h2 {
	font-size: 1.15rem;
	font-weight: 900;
	color: #3a2c1f;
}

.close-btn {
	width: 2rem;
	height: 2rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.9rem;
	font-weight: 800;
	color: #6b5438;
	background: rgba(0, 0, 0, 0.06);
	transition: background 0.2s;
}

.close-btn:hover {
	background: rgba(0, 0, 0, 0.12);
}

.replay-tip {
	margin-top: 0.35rem;
	font-size: 0.68rem;
	color: #9a7a55;
	font-weight: 600;
}

.share-box {
	margin-top: 0.6rem;
	border: 1px dashed rgba(107, 84, 56, 0.35);
	background: rgba(245, 197, 36, 0.12);
	border-radius: 0.7rem;
	padding: 0.55rem 0.7rem;
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
}

.share-label {
	font-size: 0.72rem;
	font-weight: 800;
	color: #92400e;
}

.share-link {
	font-size: 0.72rem;
	font-weight: 600;
	color: #3a2c1f;
	background: rgba(255, 255, 255, 0.65);
	border-radius: 0.5rem;
	padding: 0.45rem 0.6rem;
	word-break: break-all;
	user-select: all;
}

.share-actions {
	display: flex;
	gap: 0.5rem;
}

.copy-share {
	flex: 1;
	padding: 0.4rem;
	border-radius: 2rem;
	font-size: 0.75rem;
	font-weight: 800;
	color: #fff;
	background: linear-gradient(135deg, #b45309, #92400e);
}

.close-share {
	padding: 0.4rem 0.8rem;
	border-radius: 2rem;
	font-size: 0.72rem;
	font-weight: 700;
	color: #6b5438;
	background: rgba(0, 0, 0, 0.06);
}

.empty {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.35rem;
	padding: 2rem 0;
	text-align: center;
}

.empty p {
	font-size: 0.95rem;
	font-weight: 800;
	color: #6b5438;
}

.empty small {
	font-size: 0.72rem;
	color: #9a7a55;
}

.replay-list {
	margin-top: 0.7rem;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	flex: 1;
	min-height: 0;
	padding-right: 0.15rem;
}

.replay-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.6rem;
	padding: 0.65rem 0.75rem;
	border-radius: 0.8rem;
	background: rgba(255, 255, 255, 0.55);
	border: 1px solid rgba(255, 255, 255, 0.7);
}

.replay-info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
	min-width: 0;
}

.names {
	font-size: 0.88rem;
	font-weight: 800;
	color: #3a2c1f;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.sub {
	font-size: 0.66rem;
	font-weight: 600;
	color: #9a7a55;
}

.replay-actions {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	flex-shrink: 0;
}

.watch-btn {
	padding: 0.4rem 0.9rem;
	border-radius: 2rem;
	font-size: 0.78rem;
	font-weight: 800;
	color: #fdf6e9;
	background: linear-gradient(135deg, #d97706, #b45309);
	transition: transform 0.15s;
}

.watch-btn:hover {
	transform: scale(1.04);
}

.del-btn {
	width: 1.9rem;
	height: 1.9rem;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 0.85rem;
	background: rgba(220, 38, 38, 0.1);
	border: 1px solid rgba(220, 38, 38, 0.2);
	transition: background 0.2s;
}

.del-btn:hover {
	background: rgba(220, 38, 38, 0.25);
}

.share-btn {
	padding: 0.4rem 0.8rem;
	border-radius: 2rem;
	font-size: 0.74rem;
	font-weight: 800;
	color: #14532d;
	background: rgba(110, 231, 183, 0.3);
	border: 1px solid rgba(110, 231, 183, 0.45);
	transition: background 0.2s;
}

.share-btn:hover {
	background: rgba(110, 231, 183, 0.5);
}

.share-btn:disabled {
	opacity: 0.55;
}

.rooms-overlay {
	position: fixed;
	inset: 0;
	z-index: 120;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1rem;
	background: rgba(12, 18, 26, 0.72);
	backdrop-filter: blur(5px);
}

.rooms-panel {
	width: 100%;
	max-width: 32rem;
	max-height: min(84vh, 40rem);
	display: flex;
	flex-direction: column;
	border-radius: 1.1rem;
	padding: 1.1rem 1.2rem;
	overflow: hidden;
	animation: rise-in 0.3s ease both;
}

.rooms-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
}

.rooms-head h2 {
	font-size: 1.15rem;
	font-weight: 900;
	color: #3a2c1f;
}

.rooms-head-actions {
	display: flex;
	align-items: center;
	gap: 0.4rem;
}

.refresh-btn {
	padding: 0.35rem 0.8rem;
	border-radius: 2rem;
	font-size: 0.72rem;
	font-weight: 800;
	color: #6b4a2b;
	background: rgba(245, 197, 36, 0.35);
	transition: background 0.2s;
}

.refresh-btn:hover {
	background: rgba(245, 197, 36, 0.6);
}

.rooms-tip {
	margin-top: 0.35rem;
	font-size: 0.68rem;
	color: #9a7a55;
	font-weight: 600;
}

.rooms-empty {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.35rem;
	padding: 2rem 0;
	text-align: center;
}

.rooms-empty p {
	font-size: 0.95rem;
	font-weight: 800;
	color: #6b5438;
}

.rooms-empty small {
	font-size: 0.72rem;
	color: #9a7a55;
}

.rooms-list {
	margin-top: 0.7rem;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	flex: 1;
	min-height: 0;
	padding-right: 0.15rem;
}

.rooms-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.6rem;
	padding: 0.65rem 0.75rem;
	border-radius: 0.8rem;
	background: rgba(255, 255, 255, 0.55);
	border: 1px solid rgba(255, 255, 255, 0.7);
}

.rooms-info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
	min-width: 0;
}

.rooms-code {
	font-size: 1rem;
	font-weight: 900;
	letter-spacing: 0.18rem;
	color: #3a2c1f;
}

.rooms-sub {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	font-size: 0.68rem;
	font-weight: 600;
	color: #9a7a55;
	flex-wrap: wrap;
}

.rooms-sub .pill {
	padding: 0.12rem 0.55rem;
	border-radius: 2rem;
	font-size: 0.62rem;
	font-weight: 800;
}

.rooms-sub .pill.waiting {
	color: #92400e;
	background: rgba(245, 197, 36, 0.35);
}

.rooms-sub .pill.playing {
	color: #14532d;
	background: rgba(110, 231, 183, 0.35);
}

.rooms-sub .pill.finished {
	color: #6b5438;
	background: rgba(154, 122, 85, 0.2);
}

.players-text {
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	max-width: 12rem;
}

.viewers {
	font-weight: 800;
	color: #0f766e;
}

.spectate-btn {
	flex-shrink: 0;
	padding: 0.5rem 1rem;
	border-radius: 2rem;
	font-size: 0.78rem;
	font-weight: 800;
	color: #fff;
	background: linear-gradient(135deg, #0f766e, #115e59);
	box-shadow: 0 4px 12px rgba(15, 118, 110, 0.3);
	transition: transform 0.15s;
}

.spectate-btn:hover {
	transform: scale(1.04);
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

.deck-toggle {
	align-self: center;
	color: #b45309;
	font-size: 0.8rem;
}

.deck-editor {
	border-top: 1px dashed rgba(107, 84, 56, 0.25);
	padding-top: 0.55rem;
	max-height: 14rem;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
}

.deck-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
}

.deck-name {
	font-size: 0.85rem;
	font-weight: 800;
	color: #3a2c1f;
}

.deck-stepper {
	display: flex;
	align-items: center;
	gap: 0.65rem;
}

.step-btn {
	width: 1.7rem;
	height: 1.7rem;
	border-radius: 0.5rem;
	background: rgba(255, 255, 255, 0.6);
	border: 1px solid rgba(107, 84, 56, 0.3);
	color: #6b4a2b;
	font-size: 1rem;
	font-weight: 900;
	line-height: 1;
	display: flex;
	align-items: center;
	justify-content: center;
}

.step-btn:active {
	transform: scale(0.92);
}

.step-val {
	min-width: 1.2rem;
	text-align: center;
	font-size: 0.95rem;
	font-weight: 900;
	color: #3a2c1f;
}

.deck-zero {
	font-size: 0.75rem;
	color: #dc2626;
	font-weight: 700;
	text-align: center;
}

.deck-presets {
	display: flex;
	gap: 0.5rem;
	justify-content: center;
}

.preset-btn {
	flex: 1;
	padding: 0.3rem;
	border-radius: 0.6rem;
	background: rgba(255, 255, 255, 0.5);
	border: 1px solid rgba(107, 84, 56, 0.25);
	color: #6b4a2b;
	font-size: 0.75rem;
	font-weight: 800;
}

.preset-btn:active {
	transform: scale(0.97);
}

/* 小屏(横屏/矮窗口): 配置区更紧凑可滚动, 不挤爆面板 */
@media (max-height: 560px) and (min-width: 640px) {
	.deck-editor {
		max-height: 9rem;
		gap: 0.2rem;
	}

	.deck-name {
		font-size: 0.72rem;
	}

	.step-btn {
		width: 1.4rem;
		height: 1.4rem;
		font-size: 0.85rem;
	}

	.panel {
		max-height: calc(100vh - 1rem);
		overflow-y: auto;
	}
}
</style>
