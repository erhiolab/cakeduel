<script setup lang="ts">
import {nextTick, ref, watch} from "vue"
import {game} from "../composables/useGame"

// 单条消息最大长度
const MAX_MESSAGE_LENGTH = 200

// 未读角标上限
const MAX_UNREAD_DISPLAY = 99

// 游戏状态
const {state, sendChat} = game

// 面板是否展开
const open = ref(false)

// 输入内容
const draft = ref("")

// 消息列表元素
const listEl = ref<HTMLElement | null>(null)

// 未读数
const unread = ref(0)

// 新消息: 未展开时累计未读, 展开时滚动到底部
watch(() => state.chatMessages.length, async () => {
	if (!open.value) unread.value++
	await nextTick()
	if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight
})

// 展开/收起面板
const toggle = () => {
	open.value = !open.value
	if (open.value) unread.value = 0
}

// 发送消息
const submit = () => {
	const text = draft.value
	if (!text.trim()) return
	sendChat(text)
	draft.value = ""
	void nextTick().then(() => {
		if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight
	})
}
</script>

<template>
	<div class="chat">
		<Transition name="pop">
			<div v-if="open" class="chat-panel glass">
				<div class="chat-head">
					<span>局内对话</span>
					<button class="chat-close" @click="toggle">—</button>
				</div>
				<div ref="listEl" class="chat-list">
					<div v-if="state.chatMessages.length === 0" class="chat-empty">还没有消息，打个招呼吧</div>
					<div
						v-for="(m, i) in state.chatMessages"
						:key="i"
						class="chat-msg"
						:class="{ mine: m.from === state.playerIndex }"
					>
						<span class="chat-name">{{ m.from === state.playerIndex ? "我" : m.name }}</span>
						<span class="chat-bubble">{{ m.text }}</span>
					</div>
				</div>
				<form class="chat-input" @submit.prevent="submit">
					<input v-model="draft" :maxlength="MAX_MESSAGE_LENGTH" placeholder="输入消息…"/>
					<button type="submit" :disabled="!draft.trim()">发送</button>
				</form>
			</div>
		</Transition>
		<button class="chat-toggle" @click="toggle">
			<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"
				 stroke-linecap="round" stroke-linejoin="round">
				<path d="M21 11.5a8.38 8.38 0 0 1-8.5 8.5 8.5 8.5 0 0 1-3.8-.9L3 21l1.9-5.7a8.5 8.5 0 1 1 16.1-3.8z"/>
			</svg>
			<span v-if="unread > 0" class="unread">{{ unread > MAX_UNREAD_DISPLAY ? "99+" : unread }}</span>
		</button>
	</div>
</template>

<style scoped>
.chat {
	position: absolute;
	right: 1rem;
	bottom: 0.8rem;
	z-index: 85;
	display: flex;
	flex-direction: column;
	align-items: flex-end;
}

.chat-toggle {
	width: 3rem;
	height: 3rem;
	border-radius: 50%;
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	color: #3a2c1f;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
	position: relative;
	transition: transform 0.15s;
}

.chat-toggle:hover {
	transform: scale(1.06);
}

.unread {
	position: absolute;
	top: -0.25rem;
	right: -0.25rem;
	min-width: 1.2rem;
	height: 1.2rem;
	border-radius: 1rem;
	background: #dc2626;
	color: #fff;
	font-size: 0.68rem;
	font-weight: 800;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 0.3rem;
}

.chat-panel {
	width: 20rem;
	max-width: calc(100vw - 2rem);
	max-height: 18rem;
	margin-bottom: 0.6rem;
	border-radius: 0.9rem;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.chat-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.55rem 0.8rem;
	font-size: 0.8rem;
	font-weight: 800;
	color: #3a2c1f;
	border-bottom: 1px solid rgba(107, 84, 56, 0.12);
}

.chat-close {
	width: 1.6rem;
	height: 1.6rem;
	border-radius: 50%;
	font-size: 0.9rem;
	color: #9a7a55;
	transition: background 0.2s;
}

.chat-close:hover {
	background: rgba(0, 0, 0, 0.06);
}

.chat-list {
	flex: 1;
	min-height: 6rem;
	overflow-y: auto;
	padding: 0.55rem 0.7rem;
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
}

.chat-empty {
	text-align: center;
	color: #a8947b;
	font-size: 0.75rem;
	padding: 1.2rem 0;
}

.chat-msg {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: 0.1rem;
	max-width: 90%;
}

.chat-msg.mine {
	align-self: flex-end;
	align-items: flex-end;
}

.chat-name {
	font-size: 0.62rem;
	font-weight: 700;
	color: #a8947b;
}

.chat-bubble {
	font-size: 0.8rem;
	font-weight: 600;
	color: #3a2c1f;
	background: rgba(255, 255, 255, 0.65);
	border: 1px solid rgba(255, 255, 255, 0.7);
	border-radius: 0.6rem;
	padding: 0.35rem 0.6rem;
	word-break: break-word;
}

.chat-msg.mine .chat-bubble {
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	border-color: rgba(232, 162, 58, 0.5);
}

.chat-input {
	display: flex;
	gap: 0.4rem;
	padding: 0.55rem 0.7rem;
	border-top: 1px solid rgba(107, 84, 56, 0.12);
}

.chat-input input {
	flex: 1;
	min-width: 0;
	border-radius: 0.6rem;
	border: 1px solid rgba(107, 84, 56, 0.25);
	background: rgba(255, 255, 255, 0.6);
	padding: 0.4rem 0.6rem;
	font-size: 0.8rem;
	color: #3a2c1f;
	outline: none;
}

.chat-input input:focus {
	border-color: rgba(245, 197, 24, 0.8);
}

.chat-input button {
	border-radius: 0.6rem;
	padding: 0 0.8rem;
	background: linear-gradient(135deg, #f5c54a, #e8a23a);
	color: #3a2c1f;
	font-size: 0.78rem;
	font-weight: 800;
}

.chat-input button:disabled {
	opacity: 0.5;
}
</style>
