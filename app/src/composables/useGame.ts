import {reactive, readonly} from "vue"
import {defineStore} from "pinia"
import {playSfx} from "../game/audio"
import {appConfig} from "../config"
import {pinia} from "../stores/pinia"

/**
 * 游戏界面类型
 */
export type Screen = "start" | "lobby" | "game" | "results" | "spectate"

/**
 * 玩家信息
 */
export interface PlayerInfo {
	index: number
	name: string
	connected: boolean
}

/**
 * 卡牌实体
 */
export interface CardEntity {
	entityId: number
	name?: string
}

/**
 * 出牌消息
 */
export interface ClaimMsg {
	claim: string
	cardCount: number
}

/**
 * 游戏事件
 */
export interface GameEvent {
	id: number
	type: string
	player?: number
	phase?: string
	pile?: string
	claim?: string
	cardNames?: string[]
	revealedCards?: { cardName: string; transformedTo?: string }[]
	challenger?: number
	claimedCard?: string
	success?: boolean
	from?: number
	to?: number
	amount?: number
	cakesAfter?: [number, number]
	winner?: number
	attackerIndex?: number
	boutNumber?: number
	pickType?: string
	picks?: string[]
	zone?: string
}

/**
 * 合法操作
 */
export interface LegalAction {
	type: string
	claimFrom?: readonly string[]
	availableHandIndices?: readonly number[]
	pickType?: string
	pickFrom?: readonly string[]
	minPicks?: number
	maxPicks?: number
}

/**
 * 游戏视图
 */
export interface GameView {
	frame: number
	me: { index: number; cakes: number; hand: readonly string[]; handLimit: number }
	opponent: { index: number; cakes: number; handCount: number }
	deckCount: number
	discardCount: number
	attackingClaim: ClaimMsg | null
	blockingClaim: ClaimMsg | null
	phase: string
	attackerIndex: number
	boutWinners: readonly number[]
	gameEnded: { winner: number } | null
	config: { roundsToWin: number; specialCardsToAdd: number; startingHandLimit: number; turnTimeoutSeconds: number }
	lastAttackPassed: boolean
	roundNumber: number
}

/**
 * 游戏区域
 */
export interface Zones {
	playerHand: readonly CardEntity[]
	opponentHand: readonly CardEntity[]
	attackPile: readonly CardEntity[]
	blockPile: readonly CardEntity[]
	deckTop: readonly CardEntity[]
	revealedPileCards: Record<number, string>
	deckCount: number
	discardCount: number
}

/**
 * 观战视角(不含任何一方手牌)
 */
export interface SpectatorViewMsg {
	frame: number
	phase: string
	attackerIndex: number
	roundNumber: number
	boutWinners: readonly number[]
	gameEnded: { winner: number } | null
	config: { roundsToWin: number; specialCardsToAdd: number; startingHandLimit: number; turnTimeoutSeconds: number }
	lastAttackPassed: boolean
	paused: boolean
	players: readonly { index: number; cakes: number; handCount: number }[]
	attackingClaim: ClaimMsg | null
	blockingClaim: ClaimMsg | null
}

/**
 * 揭示消息
 */
export interface RevealMsg {
	pile: string
	cards: readonly CardEntity[]
}

/**
 * 回放单帧
 */
export interface ReplayFrame {
	view: SpectatorViewMsg
	zones: Zones
	events: GameEvent[]
	reveal?: RevealMsg
}

/**
 * 对局回放数据
 */
export interface ReplayData {
	roomCode: string
	mode: string
	playerNames: [string, string]
	startedAt: number
	durationMs: number
	winner: number
	deckConfig: Record<string, number>
	roundsToWin: number
	frames: ReplayFrame[]
}

/**
 * 聊天消息
 */
export interface ChatMsg {
	from: number
	name: string
	text: string
	ts: number
}

/**
 * 通知消息
 */
export interface BannerMsg {
	id: number
	kind: "claim" | "accepted" | "challenge" | "bout_start" | "bout_end" | "info"
	claim?: string
	cardCount?: number
	actualCards?: { entityId: number; name: string }[]
	isMine?: boolean
	victory?: boolean
	reason?: string
	text?: string
	player?: number
	boutNumber?: number
}

// 游戏状态
interface State {
	screen: Screen
	socket: WebSocket | null
	connected: boolean
	roomCode: string
	mode: string
	playerIndex: number
	players: PlayerInfo[]
	matching: boolean
	message: string
	error: string
	view: GameView | null
	zones: Zones | null
	legal: LegalAction[]
	yourTurn: boolean
	youWon: boolean
	reveal: RevealMsg | null
	banner: BannerMsg | null
	wolfyTaunt: number
	toast: string
	pendingReveal: RevealMsg | null
	revealedPileNames: Record<number, string>
	handReveal: { player: number; cards: string[] } | null
	chatMessages: ChatMsg[]
	turnRemaining: number
	paused: boolean
	connectionLost: boolean
	opponentStatus: "online" | "offline"
	rematchVotes: [boolean, boolean]
	deckConfig: Record<string, number> | null
	spectator: boolean
	spectatorView: SpectatorViewMsg | null
}

// 游戏状态
const state = reactive<State>({
	screen: "start",
	socket: null,
	connected: false,
	roomCode: "",
	mode: "",
	playerIndex: 0,
	players: [],
	matching: false,
	message: "",
	error: "",
	view: null,
	zones: null,
	legal: [],
	yourTurn: false,
	youWon: false,
	reveal: null,
	banner: null,
	wolfyTaunt: 0,
	toast: "",
	pendingReveal: null,
	revealedPileNames: {},
	handReveal: null,
	chatMessages: [],
	turnRemaining: 0,
	paused: false,
	connectionLost: false,
	opponentStatus: "online",
	rematchVotes: [false, false],
	deckConfig: null,
	spectator: false,
	spectatorView: null,
})

// 通知定时器
let bannerTimer: number | undefined

// 狼爵士定时器
let tauntTimer: number | undefined

// 提示定时器
let toastTimer: number | undefined

// 结果定时器
let resultsTimer: number | undefined

// 手牌定时器
let handRevealTimer: number | undefined

// 倒计时定时器
let countdownTimer: number | undefined

// 重连定时器
let reconnectTimer: number | undefined

// 回放存储键名
const REPLAYS_KEY = "cakeduel_replays"

// 本地最多保存的回放数量
const MAX_SAVED_REPLAYS = 10

// 待处理通知 ID 集合
const pendingBanners = new Set<number>()

// 显示通知
const showBanner = (banner: BannerMsg) => {
	state.banner = banner
	if (bannerTimer) window.clearTimeout(bannerTimer)
	const DURATION = banner.kind === "bout_end" ? 2600 : banner.kind === "challenge" ? 2600 : 1800
	bannerTimer = window.setTimeout(() => {
		state.banner = null
	}, DURATION)
}

// 显示提示
const showToast = (text: string) => {
	state.toast = text
	if (toastTimer) window.clearTimeout(toastTimer)
	toastTimer = window.setTimeout(() => {
		state.toast = ""
	}, 2600)
}

/**
 * 读取本地保存的回放列表(新的在前)
 */
const readReplays = (): ReplayData[] => {
	try {
		const RAW = localStorage.getItem(REPLAYS_KEY)
		if (!RAW) return []
		const LIST = JSON.parse(RAW)
		return Array.isArray(LIST) ? (LIST as ReplayData[]) : []
	} catch {
		return []
	}
}

/**
 * 加载本地保存的回放列表
 */
export const loadSavedReplays = (): ReplayData[] => readReplays()

/**
 * 删除本地保存的一份回放
 * @param startedAt 回放开始时间戳(用作唯一标识)
 */
export const deleteSavedReplay = (startedAt: number) => {
	const NEXT = readReplays().filter((r) => r.startedAt !== startedAt)
	localStorage.setItem(REPLAYS_KEY, JSON.stringify(NEXT))
}

/**
 * 保存一份回放到本地(超出上限时丢弃最旧)
 * @param data 回放数据
 */
const saveReplay = (data: ReplayData) => {
	try {
		const LIST = readReplays().filter((r) => r.startedAt !== data.startedAt)
		LIST.unshift(data)
		localStorage.setItem(REPLAYS_KEY, JSON.stringify(LIST.slice(0, MAX_SAVED_REPLAYS)))
	} catch {
		// 存储不可用(隐私模式/超限)时静默丢弃, 不影响对局
	}
}

/**
 * 开始质疑翻牌动画(250ms 后翻出, 2.4s 后收起)
 * @param revealMsg 翻开数据
 */
const startReveal = (revealMsg: RevealMsg) => {
	state.pendingReveal = revealMsg
	playSfx("challenge")
	window.setTimeout(() => {
		state.reveal = state.pendingReveal
		state.revealedPileNames = {}
		for (const CARD of state.pendingReveal?.cards || []) {
			if (CARD.name) state.revealedPileNames[CARD.entityId] = CARD.name
		}
		window.setTimeout(() => {
			state.reveal = null
			state.pendingReveal = null
		}, 2400)
	}, 250)
}

/**
 * 获取会话令牌
 */
export const useGame = defineStore("game", () => {
	// 是否在对局/房间中, 用于刷新后自动恢复连接(放在 sessionStorage, 标签页隔离)
	const RESUME_KEY = "cakeduel_resume"

	const markResume = () => {
		sessionStorage.setItem(RESUME_KEY, "1")
	}

	const clearResume = () => {
		sessionStorage.removeItem(RESUME_KEY)
	}

	/**
	 * 获取会话令牌
	 */
	const getSessionToken = (): string => {
		const KEY = "cakeduel_session"
		// 用 sessionStorage: 刷新页面保留(token 不变可重连),
		// 但不同标签页相互隔离, 避免同浏览器双开共用 token 导致席位/状态串扰
		let TOKEN = sessionStorage.getItem(KEY)
		if (!TOKEN) {
			TOKEN = `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`
			sessionStorage.setItem(KEY, TOKEN)
		}
		return TOKEN
	}

	/**
	 * 重连
	 */
	const scheduleReconnect = () => {
		if (reconnectTimer) return
		reconnectTimer = window.setTimeout(() => {
			reconnectTimer = undefined
			openSocket()
		}, 2000)
	}

	/**
	 * 打开 WebSocket 连接
	 */
	const openSocket = (onOpen?: () => void) => {
		const TOKEN = getSessionToken()
		const CONFIGURED = appConfig.wsUrl ?? ""
		const BASE =
			CONFIGURED ||
			(import.meta.env.DEV
				? "ws://127.0.0.1:8080"
				: `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}`)
		const URL = `${BASE}/ws?token=${encodeURIComponent(TOKEN)}`
		const WS = new WebSocket(URL)
		state.socket = WS
		WS.onopen = () => {
			state.connected = true
			// 主界面且不在匹配中时不保留空闲连接标记
			if (state.screen === "start" && !state.matching) clearResume()
			onOpen?.()
		}
		WS.onclose = () => {
			state.connected = false
			// 观战没有可恢复的席位: 断开后直接回主界面, 不反复重连
			if (state.screen === "spectate") {
				clearResume()
				resetRoom()
				state.screen = "start"
				state.message = "观战连接已断开"
				return
			}
			// 在房间/对局/匹配中: 自动重连, 服务器按 token 恢复席位
			if (state.screen !== "start" || state.matching) {
				state.connectionLost = true
				scheduleReconnect()
			} else {
				state.connectionLost = false
			}
		}
		WS.onerror = () => {
			state.connected = false
		}
		WS.onmessage = (ev) => handleMessage(ev.data)
	}

	/**
	 * 连接 WebSocket 服务器
	 * @param onOpen 连接成功回调函数
	 */
	const connect = (onOpen?: () => void) => {
		if (state.socket && state.socket.readyState === WebSocket.OPEN) {
			onOpen?.()
			return
		}
		openSocket(onOpen)
	}

	/**
	 * 发送消息到 WebSocket 服务器
	 * @param obj 消息对象
	 */
	const send = (obj: Record<string, unknown>) => {
		if (state.socket && state.socket.readyState === WebSocket.OPEN) {
			state.socket.send(JSON.stringify(obj))
		}
	}

	/**
	 * 处理 WebSocket 消息
	 * @param raw 原始消息
	 */
	const handleMessage = (raw: string) => {
		let msg: any
		try {
			msg = JSON.parse(raw)
		} catch {
			return
		}
		switch (msg.type) {
			case "room_joined":
				state.roomCode = msg.roomCode
				state.mode = msg.mode || ""
				state.players = msg.players || []
				state.deckConfig = msg.deckConfig || null
				state.matching = false
				markResume()
				state.connectionLost = false
				state.paused = false
				if (msg.spectator) {
					// 房间已满且对局中: 以观战身份进入
					state.spectator = true
					state.playerIndex = -1
					state.screen = "spectate"
					break
				}
				state.spectator = false
				state.playerIndex = msg.playerIndex
				state.opponentStatus = msg.players?.find((p: any) => p.index !== state.playerIndex)?.connected ? "online" : "offline"
				state.screen = "lobby"
				break
			case "player_joined":
				state.players = msg.players || []
				state.opponentStatus = msg.players?.find((p: any) => p.index !== state.playerIndex)?.connected ? "online" : "offline"
				break
			case "player_disconnected":
				state.paused = true
				state.opponentStatus = "offline"
				showToast(msg.reason || "对方已掉线, 对局已暂停")
				break
			case "player_reconnected":
				state.paused = false
				state.opponentStatus = "online"
				state.players = msg.players || []
				showToast("对方已重连, 对局继续")
				break
			case "rematch_status":
				state.rematchVotes = msg.rematchVotes || [false, false]
				break
			case "matching":
				state.matching = true
				state.message = msg.message || "正在匹配对手^"
				break
			case "match_timeout":
				state.matching = false
				clearResume()
				state.screen = "start"
				state.message = msg.message || "匹配超时, 请重试"
				break
			case "room_closed":
				clearResume()
				resetRoom()
				state.screen = "start"
				state.message = msg.reason || "房间已关闭"
				break
			case "opponent_left":
				clearResume()
				resetRoom()
				state.screen = "start"
				state.message = msg.reason || "对手已离开"
				break
			case "game_state":
				applyGameState(msg)
				break
			case "spectator_state":
				applySpectatorState(msg)
				break
			case "replay_data":
				if (msg.replay) {
					saveReplay(msg.replay as ReplayData)
					showToast("本局回放已保存, 可在主菜单查看")
				}
				break
			case "chat":
				state.chatMessages.push({
					from: msg.from ?? 0,
					name: msg.name ?? "",
					text: msg.text ?? "",
					ts: Date.now(),
				})
				if (state.chatMessages.length > 100) {
					state.chatMessages.splice(0, state.chatMessages.length - 100)
				}
				break
			case "error":
				state.error = msg.message || "操作失败"
				if (toastTimer) window.clearTimeout(toastTimer)
				toastTimer = window.setTimeout(() => {
					state.error = ""
				}, 3000)
				break
			case "pong":
				break
		}
	}

	/**
	 * 应用游戏状态
	 * @param msg 游戏状态消息对象
	 */
	const applyGameState = (msg: any) => {
		const PREV_FRAME = state.view?.frame ?? -1
		const PREV_VIEW = state.view
		const EVENTS: GameEvent[] = msg.events || []
		const HAS_REVEAL = !!msg.reveal
		state.view = msg.view
		state.zones = msg.zones
		state.legal = msg.legal || []
		state.yourTurn = msg.yourTurn
		state.youWon = !!msg.gameOver && !!msg.youWon
		state.paused = !!msg.paused
		state.connectionLost = false
		// 以服务端视角校准玩家身份, 防止公网延迟/重连等场景下 playerIndex 漂移
		// 导致横幅胜负判断与实际结果不一致
		if (msg.view?.me?.index != null) {
			state.playerIndex = msg.view.me.index
		}
		// 对局开始/重连恢复/再来一局: 进入对局界面
		state.screen = "game"
		// 质疑翻开动画
		if (msg.reveal) startReveal(msg.reveal as RevealMsg)

		// 事件驱动动画与音效
		for (const EVT of EVENTS) {
			processEvent(EVT, PREV_VIEW, HAS_REVEAL)
		}

		// 抽牌音效
		if (state.zones && PREV_FRAME !== -1 && state.zones.playerHand.length > (state.view?.me.hand.length ?? 0)) {
			playSfx("cardDraw")
		}

		if (state.view?.gameEnded) {
			playSfx("victory")
			if (resultsTimer) window.clearTimeout(resultsTimer)
			const delay = HAS_REVEAL ? 6800 : 3000
			resultsTimer = window.setTimeout(() => {
				state.screen = "results"
			}, delay)
		} else {
			resetCountdown()
		}
	}

	/**
	 * 应用观战状态(不含任何一方手牌)
	 * @param msg 观战状态消息
	 */
	const applySpectatorState = (msg: any) => {
		if (!msg.spectatorView) return
		state.spectator = true
		state.spectatorView = msg.spectatorView as SpectatorViewMsg
		state.zones = msg.zones as Zones
		state.paused = !!msg.spectatorView.paused
		state.connectionLost = false
		state.screen = "spectate"
		if (msg.reveal) startReveal(msg.reveal as RevealMsg)
	}

	/**
	 * 重置回合倒计时
	 */
	const resetCountdown = () => {
		const TIMEOUT = state.view?.config?.turnTimeoutSeconds ?? 0
		if (TIMEOUT <= 0) {
			state.turnRemaining = 0
			return
		}
		state.turnRemaining = TIMEOUT
		if (countdownTimer) window.clearInterval(countdownTimer)
		countdownTimer = window.setInterval(() => {
			if (state.turnRemaining > 0) state.turnRemaining--
		}, 1000)
	}

	/**
	 * 处理游戏事件
	 * @param evt 游戏事件对象
	 * @param prevView 上一游戏视图对象或 null
	 * @param hasReveal 是否有确认牌对象
	 */
	const processEvent = (evt: GameEvent, prevView: GameView | null, hasReveal: boolean) => {
		switch (evt.type) {
			case "claim_made":
				playSfx("cardPlay")
				showBanner({
					id: evt.id,
					kind: "claim",
					claim: evt.claim,
					cardCount: evt.cardNames?.length ?? 0,
					isMine: evt.player === state.playerIndex,
				})
				break
			case "pass_made":
				playSfx("hoof")
				// 接受声明后, 根据之前声明的卡牌给出效果提示
				if (prevView) {
					const ATK = prevView.attackingClaim
					const BLK = prevView.blockingClaim
					if (ATK) {
						const TIP = attackEffectTip(ATK.claim)
						if (TIP) showToast(TIP)
					}
					if (BLK) {
						const TIP = blockEffectTip(BLK.claim)
						if (TIP) showToast(TIP)
					}
				}
				break
			case "challenge_made":
				break
			case "bout_started":
				// 质疑后等翻牌动画结束再展示下一局
				if (hasReveal) {
					scheduleBanner({
						id: evt.id,
						kind: "bout_start",
						boutNumber: state.view?.roundNumber ?? 1,
					} as BannerMsg, 5500)
				} else {
					showBanner({
						id: evt.id,
						kind: "bout_start",
						boutNumber: state.view?.roundNumber ?? 1,
					} as BannerMsg)
				}
				break
			case "bout_ended": {
				// 最后一局直接采用服务端下发的 youWon(权威结果, 不依赖本地身份/时序),
				// 中间局用事件 winner 与本端视角比较
				const FINAL = state.view?.gameEnded != null
				const meWon = FINAL ? !!state.youWon : (evt.winner ?? -1) === state.playerIndex
				playSfx("roundResult")
				const banner = {
					id: evt.id,
					kind: "bout_end",
					victory: meWon,
					reason: FINAL ? resultReason(meWon) : deriveReason(evt),
					player: evt.winner,
				} as BannerMsg
				if (hasReveal) {
					scheduleBanner(banner, 2900)
				} else {
					showBanner(banner)
				}
				break
			}
			case "wolfy_taunt":
				playSfx("wolfTaunt")
				state.wolfyTaunt++
				if (tauntTimer) window.clearTimeout(tauntTimer)
				tauntTimer = window.setTimeout(() => {
					state.wolfyTaunt = 0
				}, 1600)
				break
			case "cakes_transferred":
				playSfx("cakeTransfer")
				break
			case "hand_revealed":
				if (evt.player !== undefined && evt.player !== null) {
					state.handReveal = {
						player: evt.player,
						cards: evt.cardNames ?? [],
					}
					if (handRevealTimer) window.clearTimeout(handRevealTimer)
					handRevealTimer = window.setTimeout(() => {
						state.handReveal = null
					}, 3600)
				}
				break
			case "pick_made":
				if (evt.picks?.length) {
					showToast(`选牌: ${evt.picks.join(", ")}`)
				}
				break
			case "concede_made":
				showToast("玩家认输")
				break
		}
	}

	/**
	 * 安时显示 banner
	 * @param banner
	 * @param delay 延迟时间, 单位毫秒
	 */
	const scheduleBanner = (banner: BannerMsg, delay: number) => {
		const TIMER = window.setTimeout(() => {
			pendingBanners.delete(TIMER)
			showBanner(banner)
		}, delay)
		pendingBanners.add(TIMER)
	}

	/**
	 * 获取攻击效果提示
	 * @param claim 攻击声明
	 */
	const attackEffectTip = (claim: string): string => {
		switch (claim) {
			case "quartermaster":
				return "军需官: 手牌上限 +1"
			case "oracle":
				return "神谕师: 查看了对手手牌, 再行动一回合"
			case "scout":
				return "斥候: 再行动一回合"
			case "assassin":
				return "刺客: 抢走 5 个蛋糕"
			case "wizard":
				return "法师: 抢走 2 个蛋糕"
			case "summoner":
				return "召唤师: 选一张牌, 查看对手手牌"
			default:
				return ""
		}
	}

	/**
	 * 获取防御效果提示
	 * @param claim 防御声明
	 */
	const blockEffectTip = (claim: string): string => {
		switch (claim) {
			case "priest":
				return "牧师: 挡住攻击并收入一张牌"
			case "angel":
				return "天使: 挡下所有攻击, 对手再行动一回合"
			default:
				return ""
		}
	}

	/**
	 * 获取游戏结束原因
	 * @param evt 游戏事件对象
	 */
	const deriveReason = (evt: GameEvent): string => {
		const ME = state.playerIndex
		const WINNER = evt.winner ?? -1
		const MY_NAME = state.players.find((p) => p.index === ME)?.name || "你"
		// 按己方视角返回文案: 胜者看到"我赢", 败者看到"我输", 两侧不会读到同一句
		return WINNER === ME ? `${MY_NAME} 赢得了本局` : `${MY_NAME} 输掉了本局`
	}

	/**
	 * 整场对局结束时的结果文案(按己方视角, 避免两边读到同一句话)
	 * @param victory 本端是否获胜
	 */
	const resultReason = (victory: boolean): string => {
		const OPPONENT = 1 - state.playerIndex
		const OPP_NAME = state.players.find((p) => p.index === OPPONENT)?.name || "对手"
		return victory ? `击败了 ${OPP_NAME}` : `输给了 ${OPP_NAME}`
	}

	/**
	 * 创建房间
	 * @param name 房间名称
	 * @param mode 房间模式
	 * @param deckConfig 牌组配置
	 */
	const createRoom = (name: string, mode: "private" | "random", deckConfig?: Record<string, number> | null) => {
		connect(() => {
			send({type: "create_room", name, mode, deckConfig: deckConfig ?? undefined})
		})
	}

	/**
	 * 加入房间
	 * @param code 房间代码
	 * @param name 房间名称
	 */
	const joinRoom = (code: string, name: string) => {
		connect(() => {
			send({type: "join_room", code, name})
		})
	}

	/**
	 * 开始游戏
	 */
	const startGame = () => {
		send({type: "start_game"})
	}

	/**
	 * 执行操作
	 * @param action 操作对象
	 */
	const act = (action: Record<string, unknown>) => {
		send({type: "action", action})
	}

	/**
	 * 重新开始游戏
	 */
	const rematch = () => {
		send({type: "rematch"})
	}

	/**
	 * 离开房间
	 */
	const leave = () => {
		send({type: "leave"})
		clearResume()
		resetRoom()
		state.screen = "start"
	}

	/**
	 * 重置房间状态
	 */
	const resetRoom = () => {
		for (const timer of pendingBanners) {
			window.clearTimeout(timer)
		}
		pendingBanners.clear()
		if (countdownTimer) window.clearInterval(countdownTimer)
		countdownTimer = undefined
		if (reconnectTimer) window.clearTimeout(reconnectTimer)
		reconnectTimer = undefined
		state.turnRemaining = 0
		state.roomCode = ""
		state.mode = ""
		state.playerIndex = 0
		state.players = []
		state.view = null
		state.zones = null
		state.legal = []
		state.youWon = false
		state.reveal = null
		state.pendingReveal = null
		state.banner = null
		state.wolfyTaunt = 0
		state.matching = false
		state.handReveal = null
		state.paused = false
		state.connectionLost = false
		state.opponentStatus = "online"
		state.rematchVotes = [false, false]
		state.deckConfig = null
		state.spectator = false
		state.spectatorView = null
	}

	/**
	 * 清除手牌显示
	 */
	const clearHandReveal = () => {
		state.handReveal = null
		if (handRevealTimer) window.clearTimeout(handRevealTimer)
	}

	/**
	 * 发送聊天消息
	 * @param text 聊天消息
	 */
	const sendChat = (text: string) => {
		const TRIMMED = text.trim()
		if (!TRIMMED) return
		send({type: "chat", text: TRIMMED.slice(0, 200)})
	}

	return {
		state: readonly(state),
		createRoom,
		joinRoom,
		startGame,
		act,
		rematch,
		leave,
		connect,
		clearHandReveal,
		sendChat,
		clearError: () => (state.error = ""),
		clearMessage: () => (state.message = ""),
	}
})

export const game = useGame(pinia)
