// 移动端自动全屏
//
// 浏览器要求全屏必须由用户手势触发, 因此这里在首次点击/触摸/按键时尝试进入全屏;
// 若被拒绝(未处于手势中/被浏览器策略拦截)则等下一次手势重试, 成功后停止监听。
// 桌面端、已作为独立应用运行(无浏览器界面)以及不支持全屏 API 的浏览器不做处理。

// 视为移动端的最大宽度(px)
export const MOBILE_BREAKPOINT = 900

// 全屏状态与请求(含 Safari 的 webkit 前缀实现)
type FullscreenDocument = Document & {
	webkitFullscreenElement?: Element | null
}

type FullscreenTarget = HTMLElement & {
	webkitRequestFullscreen?: (options?: FullscreenOptions) => Promise<void> | void
}

// 是否为移动设备
export const isMobileDevice = (): boolean => {
	if (window.matchMedia("(pointer: coarse)").matches) return true
	if (/android|iphone|ipad|ipod|mobile/i.test(navigator.userAgent)) return true
	return window.innerWidth < MOBILE_BREAKPOINT
}

// 是否已作为独立应用运行(iOS 添加到主屏幕 / PWA 安装后本身就无浏览器界面)
export const isStandaloneDisplay = (): boolean => {
	if (window.matchMedia("(display-mode: standalone)").matches) return true
	return (navigator as Navigator & {standalone?: boolean}).standalone === true
}

// 当前是否处于全屏
export const isFullscreenActive = (): boolean => {
	const DOC = document as FullscreenDocument
	return !!(DOC.fullscreenElement ?? DOC.webkitFullscreenElement)
}

// 浏览器是否支持元素全屏(iOS Safari 不支持, 只能依赖添加到主屏幕)
export const isFullscreenSupported = (): boolean => {
	const TARGET = document.documentElement as FullscreenTarget
	return typeof (TARGET.requestFullscreen ?? TARGET.webkitRequestFullscreen) === "function"
}

// 全屏后尝试锁定横屏(仅部分移动端浏览器支持, 失败不影响全屏)
const lockLandscape = async () => {
	const LOCK = screen.orientation?.lock
	if (typeof LOCK !== "function") return
	try {
		await LOCK.call(screen.orientation, "landscape")
	} catch {
		// 忽略: 部分浏览器仅在特定条件下允许锁定方向
	}
}

// 请求全屏(必须在用户手势回调中同步调用, 否则会被浏览器拒绝)
export const enterFullscreen = async (): Promise<boolean> => {
	if (isFullscreenActive()) return true
	const TARGET = document.documentElement as FullscreenTarget
	const REQUEST = TARGET.requestFullscreen ?? TARGET.webkitRequestFullscreen
	if (typeof REQUEST !== "function") return false
	try {
		// navigationUI 用于隐藏部分浏览器残留的导航栏, 不支持该参数的浏览器会忽略
		await REQUEST.call(TARGET, {navigationUI: "hide"})
	} catch {
		try {
			await REQUEST.call(TARGET)
		} catch {
			return false
		}
	}
	await lockLandscape()
	return true
}

// 自动全屏: 监听用户手势并尝试进入全屏, 返回清理函数
export const setupAutoFullscreen = (): (() => void) => {
	if (!isMobileDevice() || isStandaloneDisplay() || !isFullscreenSupported()) return () => {}
	// 触发时机: 首次点击/触摸/按键(浏览器只认可真实用户手势)
	const EVENTS = ["pointerdown", "touchend", "keydown"]
	// 是否正在请求全屏(避免同一手势重复触发)
	let pending = false
	const cleanup = () => {
		for (const EVENT of EVENTS) window.removeEventListener(EVENT, onGesture, true)
	}
	const onGesture = () => {
		if (pending || isFullscreenActive()) return
		pending = true
		void enterFullscreen().then((OK) => {
			pending = false
			// 进入全屏后不再监听: 用户主动退出的全屏不强推回去
			if (OK) cleanup()
		})
	}
	for (const EVENT of EVENTS) window.addEventListener(EVENT, onGesture, {capture: true, passive: true})
	return cleanup
}
