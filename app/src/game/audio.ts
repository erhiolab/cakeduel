// 音频管理
const BASE = "/audio/cakeduel"

/**
 * 音效管理
 */
export const SFX = {
	hoof: `${BASE}/hoof.wav`,
	cardPlay: `${BASE}/card-play.wav`,
	cardDraw: `${BASE}/card-draw.wav`,
	challenge: `${BASE}/challenge.wav`,
	cakeTransfer: `${BASE}/cake-transfer.wav`,
	roundResult: `${BASE}/round-result.wav`,
	victory: `${BASE}/victory.wav`,
	wolfTaunt: `${BASE}/wolf-taunt.wav`,
}

// 背景音乐管理
let bgm: HTMLAudioElement | null = null

// 启用音频管理
let enabled = false

/**
 * 启用音频
 */
export const enableAudio = () => {
	enabled = true
	playBgm()
}

/**
 * 播放音效
 */
export const playSfx = (key: keyof typeof SFX) => {
	if (!enabled) return
	try {
		const AUDIO = new Audio(SFX[key])
		AUDIO.volume = key === "cardDraw" ? 0.7 : 0.9
		void AUDIO.play().catch(() => {
		})
	} catch {
		// 忽略音频错误
	}
}

/**
 * 播放背景音乐
 */
export const playBgm = () => {
	if (!enabled || bgm) return
	try {
		bgm = new Audio(`${BASE}/bgm.m4a`)
		bgm.loop = true
		bgm.volume = 0.55
		void bgm.play().catch(() => {
		})
	} catch {
		// 忽略音频错误
	}
}

/**
 * 停止背景音乐
 */
export const stopBgm = () => {
	if (bgm) {
		bgm.pause()
		bgm = null
	}
}
