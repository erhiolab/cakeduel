<script setup lang="ts">
import {onMounted, onUnmounted} from "vue"

// 版本号: 与 app/package.json、backend/internal/version/version.go 保持一致
// (由 scripts/bump-version.mjs 统一更新)
const VERSION = "0.2.0"

// GitHub 仓库地址
const REPO_URL = "https://github.com/erhiolab/cakeduel"

// 原版设计与作者
const ORIGIN_AUTHOR = "I_Nori"

// 原版视频地址
const ORIGIN_VIDEO_URL = "https://www.bilibili.com/video/BV1h8uq6SECj"

// 开源协议与版权信息(与 LICENSE / README 一致)
const LICENSE_NAME = "MIT License"

const LICENSE_YEAR = "2026"

const LICENSE_HOLDER = "Cake Duel Contributors"

defineProps<{
	open: boolean
}>()

const emit = defineEmits<{
	close: []
}>()

// 按 Esc 关闭
const onKeydown = (e: KeyboardEvent) => {
	if (e.key === "Escape") emit("close")
}

onMounted(() => {
	window.addEventListener("keydown", onKeydown)
})

onUnmounted(() => {
	window.removeEventListener("keydown", onKeydown)
})
</script>

<template>
	<div v-if="open" class="overlay" @click.self="emit('close')">
		<div class="sheet">
			<div class="head">
				<h2 class="title-font">关于</h2>
				<button class="close" aria-label="关闭" @click="emit('close')">✕</button>
			</div>
			<div class="body">
				<div class="hero">
					<img src="/cakeduel/trophy.png" alt="" draggable="false"/>
					<div>
						<h3>蛋糕对决 Cake Duel</h3>
						<p>在线双人暗牌对决 · 类骗子酒馆</p>
					</div>
				</div>

				<dl>
					<dt>版本</dt>
					<dd>v{{ VERSION }}（前端 / 后端同版本）</dd>

					<dt>仓库地址</dt>
					<dd>
						<a :href="REPO_URL" target="_blank" rel="noopener noreferrer">{{ REPO_URL.replace("https://", "") }}</a>
					</dd>

					<dt>玩法与设计</dt>
					<dd>
						来自 <b>{{ ORIGIN_AUTHOR }}</b> 的原版「蛋糕对决」：
						<a :href="ORIGIN_VIDEO_URL" target="_blank" rel="noopener noreferrer">BV1h8uq6SECj</a>
					</dd>

					<dt>技术栈</dt>
					<dd>Vue 3 + TypeScript + Vite · Go + Gorilla WebSocket</dd>

					<dt>开源协议</dt>
					<dd>{{ LICENSE_NAME }} · © {{ LICENSE_YEAR }} {{ LICENSE_HOLDER }}</dd>
				</dl>

				<p class="notice">
					本项目为独立复刻的开源爱好者作品：规则、素材与音效来自 {{ ORIGIN_AUTHOR }} 的原版游戏，
					代码基于 MIT License 发布，完整说明见仓库 README。
				</p>
			</div>
		</div>
	</div>
</template>

<style scoped>
.overlay {
	position: fixed;
	inset: 0;
	z-index: 90;
	display: flex;
	align-items: safe center;
	justify-content: center;
	padding: 1rem;
	background: rgba(48, 34, 22, 0.45);
	backdrop-filter: blur(10px) saturate(120%);
}

.sheet {
	width: 100%;
	max-width: 26rem;
	max-height: min(84vh, 38rem);
	overflow-y: auto;
	display: flex;
	flex-direction: column;
	border-radius: 1rem;
	background: linear-gradient(180deg, #fdf6e9, #f3e4cd);
	border: 1px solid rgba(255, 255, 255, 0.75);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9), 0 18px 48px rgba(40, 24, 10, 0.45);
}

.head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem 1.4rem 0.5rem;
	position: sticky;
	top: 0;
	z-index: 2;
	background: linear-gradient(180deg, #fdf6e9 82%, rgba(253, 246, 233, 0));
}

.head h2 {
	font-size: 1.4rem;
	color: #6b4a2b;
}

.close {
	width: 2.4rem;
	height: 2.4rem;
	border-radius: 50%;
	font-size: 1rem;
	color: #9a7a55;
	background: rgba(0, 0, 0, 0.05);
	flex-shrink: 0;
	transition: background 0.2s;
}

.close:hover {
	background: rgba(0, 0, 0, 0.1);
}

.body {
	padding: 0.2rem 1.4rem 1.4rem;
	flex: 1;
}

.hero {
	display: flex;
	align-items: center;
	gap: 0.9rem;
	padding: 0.8rem 0.9rem;
	margin-bottom: 0.8rem;
	border-radius: 0.9rem;
	background: rgba(255, 255, 255, 0.55);
	border: 1px solid rgba(107, 84, 56, 0.18);
}

.hero img {
	width: 3rem;
	height: 3rem;
	object-fit: contain;
	flex-shrink: 0;
	filter: drop-shadow(0 3px 6px rgba(0, 0, 0, 0.25));
}

.hero h3 {
	font-size: 1.05rem;
	font-weight: 900;
	color: #6b4a2b;
}

.hero p {
	margin-top: 0.2rem;
	font-size: 0.72rem;
	font-weight: 600;
	color: #9a7a55;
}

dl {
	display: flex;
	flex-direction: column;
	gap: 0.55rem;
}

dt {
	font-size: 0.62rem;
	font-weight: 900;
	letter-spacing: 0.12em;
	text-transform: uppercase;
	color: #c4756a;
	margin-bottom: 0.12rem;
}

dd {
	font-size: 0.82rem;
	font-weight: 600;
	line-height: 1.5;
	color: #6b4a2b;
	overflow-wrap: anywhere;
}

dd a {
	color: #b45309;
	font-weight: 800;
	text-decoration: underline;
	text-underline-offset: 0.15rem;
}

.notice {
	margin-top: 0.8rem;
	padding-top: 0.7rem;
	border-top: 1px dashed rgba(107, 84, 56, 0.25);
	font-size: 0.68rem;
	line-height: 1.55;
	font-weight: 600;
	color: #9a7a55;
}

/* 小屏适配 */
@media (max-width: 480px) {
	.overlay {
		padding: 0.5rem;
	}

	.sheet {
		max-height: 88vh;
	}

	.body {
		padding: 0.1rem 1rem 1rem;
	}

	.head {
		padding: 0.8rem 1rem 0.4rem;
	}

	.hero img {
		width: 2.5rem;
		height: 2.5rem;
	}

	.hero h3 {
		font-size: 0.95rem;
	}

	dd {
		font-size: 0.76rem;
	}
}
</style>
