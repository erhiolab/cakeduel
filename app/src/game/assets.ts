import {CARD_KEYS} from "./cards"
import {SFX} from "./audio"

// 缓存版本: 资源/路径有变化时递增, 同时同步 public/sw.js 里的 CACHE_NAME
const CACHE_VERSION = "v0.4.3"

// 浏览器 Cache Storage 名称
export const ASSET_CACHE_NAME = `cakeduel-assets-${CACHE_VERSION}`

// 本地“已缓存”标记(带版本, 升级后会自动重新缓存)
const CACHED_FLAG_KEY = `cakeduel_assets_cached_${CACHE_VERSION}`

// 背景图/图标等常驻资源
const STATIC_ASSETS = [
	"/cakeduel/playmat.jpg",
	"/cakeduel/card-back-hd.jpg",
	"/cakeduel/card-back.jpg",
	"/cakeduel/token-cake.png",
	"/cakeduel/token-trophy.png",
	"/cakeduel/trophy.png",
	"/cakeduel/cake.png",
	"/cakeduel/waggle/waggle0.png",
	"/cakeduel/waggle/waggle1.png",
	"/cakeduel/waggle/waggle2.png",
	"/cakeduel/waggle/waggle3.png",
]

// 音频资源(音效 + 背景音乐)
const AUDIO_ASSETS = [
	"/audio/cakeduel/bgm.m4a",
	...Object.values(SFX),
]

/**
 * 需要预缓存的资源地址列表(小图与高清卡图都缓存, 防止空白卡)
 */
export const buildAssetUrls = (): string[] => {
	const CARDS: string[] = []
	for (const KEY of CARD_KEYS) {
		CARDS.push(`/cakeduel/cards/zh-CN/${KEY}.jpg`)
		CARDS.push(`/cakeduel/cards-hd/zh-CN/${KEY}.jpg`)
	}
	return [...new Set([...STATIC_ASSETS, ...CARDS, ...AUDIO_ASSETS])]
}

/**
 * 是否首次启动/版本升级需要重新预缓存
 */
export const shouldPreloadAssets = (): boolean => {
	if (typeof caches === "undefined" || typeof localStorage === "undefined") return false
	return localStorage.getItem(CACHED_FLAG_KEY) !== "1"
}

/**
 * 预缓存资源到浏览器 Cache Storage
 * @param onProgress 进度回调(已缓存数, 总数)
 */
export const preloadAssets = async (onProgress?: (done: number, total: number) => void): Promise<boolean> => {
	if (typeof caches === "undefined") return false
	try {
		const CACHE = await caches.open(ASSET_CACHE_NAME)
		const URLS = buildAssetUrls()
		let done = 0
		await Promise.all(
			URLS.map(async (url) => {
				try {
					const REQ = new Request(url, {cache: "no-store"})
					const RES = await fetch(REQ)
					if (RES.ok) await CACHE.put(url, RES)
				} catch {
					// 单个资源失败不阻塞整体, 交给 Service Worker 后续按需补缓存
				} finally {
					done++
					onProgress?.(done, URLS.length)
				}
			}),
		)
		localStorage.setItem(CACHED_FLAG_KEY, "1")
		return true
	} catch {
		return false
	}
}
