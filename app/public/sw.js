// 蛋糕对决资源缓存 Service Worker
// CACHE_NAME 需与 app/src/game/assets.ts 的 ASSET_CACHE_NAME 保持一致(版本变更时同步)
const CACHE_NAME = "cakeduel-assets-v0.4.2"

// 只缓存同源的静态资源(图片/音频/样式/脚本/字体), 接口请求永远走网络
const CACHEABLE = new Set(["image", "audio", "script", "style", "font"])

self.addEventListener("install", (event) => {
	// 新版本立即接管页面, 避免旧缓存文件长期生效
	self.skipWaiting()
})

self.addEventListener("activate", (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(
					keys
						.filter((key) => key.startsWith("cakeduel-assets-") && key !== CACHE_NAME)
						.map((key) => caches.delete(key)),
				),
			)
			.then(() => self.clients.claim()),
	)
})

self.addEventListener("fetch", (event) => {
	const REQ = event.request
	if (REQ.method !== "GET" || !REQ.url.startsWith(self.location.origin)) return
	if (!CACHEABLE.has(REQ.destination) && !/^https?:[^/]+\/cakeduel\//.test(REQ.url)) return

	event.respondWith(
		caches.match(REQ).then((cached) => {
			if (cached) return cached
			return fetch(REQ).then((response) => {
				if (response && response.ok) {
					const COPY = response.clone()
					caches.open(CACHE_NAME).then((cache) => {
						cache.put(REQ, COPY)
					})
				}
				return response
			})
		}),
	)
})
