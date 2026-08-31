<script setup lang="ts">
import {computed, ref, watch} from "vue"
import {game} from "../composables/useGame"
import type {LegalAction} from "../composables/useGame"
import {CARDS} from "../game/cards"
import {playSfx} from "../game/audio"
import {resetSelection, selection} from "../composables/useSelection"

// 过牌二次确认的自动取消时间(ms)
const PASS_CONFIRM_TIMEOUT_MS = 3000

// 游戏状态与动作
const {state, act} = game

// 是否已确认过牌
const passConfirm = ref(false)

// 过牌确认自动取消定时器
let passTimer: number | undefined

// 选中的手牌实体 ID
const selectedIds = computed(() => selection.ids)

// 选中的声明
const selectedClaim = ref(selection.claim)

// 选中的召唤师选项
const selectedPick = ref(selection.pick)

// 出牌/选牌类合法动作
const claimAction = computed<LegalAction | undefined>(() => state.legal.find((l) => l.type === "claim"))

// 召唤师类合法动作
const pickAction = computed<LegalAction | undefined>(() => state.legal.find((l) => l.type === "pick"))

// 放行/质疑是否可用
const canPass = computed(() => state.legal.some((l) => l.type === "pass"))

// 质疑是否可用
const canChallenge = computed(() => state.legal.some((l) => l.type === "challenge"))

// 手牌实体
const handCards = computed(() => state.zones?.playerHand ?? [])

// 手牌牌名
const handNames = computed(() => state.view?.me.hand ?? [])

// 选中的牌在 hand 中的索引(服务端要求的 handIndices)
const selectedHandIndices = computed(() => {
	const INDICES: number[] = []
	handCards.value.forEach((card, idx) => {
		if (selectedIds.value.has(card.entityId)) INDICES.push(idx)
	})
	return INDICES
})

// 选中的牌名(全部相同时)
const selectedNames = computed(() => {
	return selectedHandIndices.value.map((i) => handNames.value[i]).filter(Boolean)
})

// 行动栏模式
const mode = computed<"claim" | "block_response" | "response" | "pick" | "attack_pass" | "paused" | "hidden">(() => {
	if (!state.view || !state.yourTurn) return "hidden"
	if (state.paused) return "paused"
	if (pickAction.value) return "pick"
	if (claimAction.value && canChallenge.value) return "block_response"
	if (claimAction.value) {
		return selectedHandIndices.value.length === 0 && canPass.value ? "attack_pass" : "claim"
	}
	if (canPass.value || canChallenge.value) return "response"
	return "hidden"
})

// 自动选择声明: 选中的牌同名且可声明时自动选中
watch([selectedNames, selectedHandIndices], () => {
	if (!claimAction.value) return
	if (selectedHandIndices.value.length === 0) {
		selectedClaim.value = ""
		return
	}
	const names = selectedNames.value
	if (names.length && names.every((n) => n === names[0])) {
		const n = names[0]
		if (claimAction.value.claimFrom?.includes(n)) {
			selectedClaim.value = n
			return
		}
	}
	if (!selectedClaim.value || !claimAction.value.claimFrom?.includes(selectedClaim.value)) {
		selectedClaim.value = claimAction.value.claimFrom?.[0] ?? ""
	}
	selection.claim = selectedClaim.value
}, {immediate: true})

// 每帧(服务端状态更新)后重置本地选择
watch(() => state.view?.frame, () => {
	resetSelection()
	passConfirm.value = false
})

// 提交出牌声明
const submitClaim = () => {
	if (selectedHandIndices.value.length === 0 || !selectedClaim.value) return
	act({type: "claim", handIndices: selectedHandIndices.value, claim: selectedClaim.value})
}

// 提交召唤师选牌
const submitPick = () => {
	if (selectedPick.value == null) return
	act({type: "pick", pickIndices: [selectedPick.value]})
}

// 过牌(进攻过牌需要二次确认)
const doPass = () => {
	if (mode.value === "attack_pass") {
		if (!passConfirm.value) {
			passConfirm.value = true
			playSfx("hoof")
			if (passTimer) window.clearTimeout(passTimer)
			passTimer = window.setTimeout(() => (passConfirm.value = false), PASS_CONFIRM_TIMEOUT_MS)
			return
		}
	}
	playSfx("hoof")
	act({type: "pass"})
}

// 质疑
const doChallenge = () => {
	playSfx("hoof")
	act({type: "challenge"})
}

// 取卡牌主题色
const claimColor = (claim: string) => {
	return CARDS[claim]?.color ?? "#B7C0C3"
}

// 十六进制颜色转 rgba
const rgba = (hex: string, a: number) => {
	const R = parseInt(hex.slice(1, 3), 16)
	const G = parseInt(hex.slice(3, 5), 16)
	const B = parseInt(hex.slice(5, 7), 16)
	return `rgba(${R},${G},${B},${a})`
}

// 声明选项胶囊样式(选中/未选中)
const pillStyle = (claim: string, on: boolean) => {
	const C = claimColor(claim)
	return on
		? {
			background: C,
			color: "#1c1917",
			boxShadow: `0 2px 10px ${rgba(C, 0.35)}, inset 0 1px 0 rgba(255,255,255,0.35)`,
			border: `1.5px solid ${C}`
		}
		: {background: rgba(C, 0.6), color: "#78716c", border: `1.5px solid ${C}`}
}

// 出牌按钮是否可提交
const ready = computed(() => selectedClaim.value !== "" && selectedHandIndices.value.length > 0)

// 等待对手行动
const waiting = computed(() => state.view != null && !state.yourTurn && !state.view.gameEnded)
</script>

<template>
	<div class="action-bar">
		<div v-if="mode === 'paused'" class="panel glass paused-panel">
			<span>对局已暂停，等待对方重连…</span>
		</div>
		<div v-if="waiting" class="waiting glass">
			<span>等待对手…</span>
		</div>
		<div v-if="mode === 'claim' || mode === 'block_response'" class="panel glass">
			<template v-if="mode === 'block_response'">
				<div class="hint">回应对方的声明</div>
				<div class="btn-row">
					<button v-if="canPass" class="btn muted" @click="doPass">
						<span class="btn-label">放行</span>
						<span class="btn-sub">接受对方声明</span>
					</button>
					<button v-if="canChallenge" class="btn danger" @click="doChallenge">
						<span class="btn-label">质疑</span>
						<span class="btn-sub">翻开牌面</span>
					</button>
				</div>
				<div class="or-row">
					<svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
						<path d="M12 0 L15 9 L24 12 L15 15 L12 24 L9 15 L0 12 L9 9 Z"/>
					</svg>
					或从手牌中出牌
					<svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
						<path d="M12 0 L15 9 L24 12 L15 15 L12 24 L9 15 L0 12 L9 9 Z"/>
					</svg>
				</div>
			</template>
			<div v-if="claimAction!.claimFrom!.length > 1" class="claim-from">
				<span class="mini-label">声明为</span>
				<button
					v-for="c in claimAction!.claimFrom!"
					:key="c"
					class="pill"
					:class="{ on: selectedClaim === c }"
					:style="pillStyle(c, selectedClaim === c)"
					@click="selectedClaim = c"
				>
					{{ CARDS[c]?.name ?? c }}
				</button>
			</div>
			<div class="submit-row">
				<button class="btn primary" :class="{ pulse: ready }" :disabled="!ready" @click="submitClaim">
					<span class="btn-label">
						出 {{ selectedHandIndices.length }}× {{ CARDS[selectedClaim]?.name ?? selectedClaim }}
					</span>
					<span class="btn-sub">扣下并声明</span>
				</button>
			</div>
		</div>
		<div v-if="mode === 'attack_pass'" class="panel glass">
			<button class="btn" :class="passConfirm ? 'danger' : 'muted'" @click="doPass">
				<span class="btn-label">{{ passConfirm ? "确认过牌？" : "过牌" }}</span>
				<span class="btn-sub">放弃本次进攻</span>
			</button>
			<div class="or-row">
				<svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
					<path d="M12 0 L15 9 L24 12 L15 15 L12 24 L9 15 L0 12 L9 9 Z"/>
				</svg>
				或从手牌中出牌
				<svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
					<path d="M12 0 L15 9 L24 12 L15 15 L12 24 L9 15 L0 12 L9 9 Z"/>
				</svg>
			</div>
		</div>
		<div v-if="mode === 'response'" class="panel glass">
			<div class="hint">回应对方的声明</div>
			<div class="btn-row">
				<button v-if="canPass" class="btn muted" @click="doPass">
					<span class="btn-label">放行</span>
					<span class="btn-sub">接受对方声明</span>
				</button>
				<button v-if="canChallenge" class="btn danger" @click="doChallenge">
					<span class="btn-label">质疑</span>
					<span class="btn-sub">翻开牌面</span>
				</button>
			</div>
		</div>
		<div v-if="mode === 'pick'" class="panel glass">
			<div class="hint">选一张牌</div>
			<div class="pick-row">
				<button
					v-for="(c, i) in pickAction!.pickFrom!"
					:key="c"
					class="pill"
					:class="{ on: selectedPick === i }"
					:style="pillStyle(c, selectedPick === i)"
					@click="selectedPick = i"
				>
					{{ CARDS[c]?.name ?? c }}
				</button>
			</div>
			<div class="submit-row">
				<button class="btn primary" :disabled="selectedPick == null" @click="submitPick">
					<span class="btn-label">确认选择</span>
				</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
.action-bar {
	position: relative;
	display: flex;
	justify-content: center;
	z-index: 30;
	padding-bottom: 0.3rem;
}

.panel {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	min-width: 18rem;
	max-width: 38rem;
	padding: 0.8rem 1rem;
	border-radius: 1rem;
}

.hint {
	text-align: center;
	font-size: 0.65rem;
	font-weight: 700;
	letter-spacing: 0.12em;
	text-transform: uppercase;
	color: #a8a29e;
}

.btn-row {
	display: flex;
	justify-content: center;
	gap: 0.7rem;
}

.btn {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.1rem;
	border-radius: 0.8rem;
	padding: 0.55rem 1rem;
	font-weight: 800;
	font-size: 0.85rem;
	transition: transform 0.15s, box-shadow 0.2s, opacity 0.2s;
}

.btn:hover:not(:disabled) {
	transform: scale(1.03);
}

.btn:active:not(:disabled) {
	transform: scale(0.97);
}

.btn:disabled {
	opacity: 0.5;
}

.btn-label {
	white-space: nowrap;
}

.btn-sub {
	font-size: 0.6rem;
	font-weight: 600;
	opacity: 0.7;
	white-space: nowrap;
}

.btn.primary {
	background: linear-gradient(to bottom, #f5c518, #d9a20c);
	color: #1c1917;
	border: 1px solid rgba(245, 197, 24, 0.5);
	box-shadow: 0 4px 14px rgba(245, 197, 24, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.btn.muted {
	background: rgba(255, 255, 255, 0.45);
	color: #44403c;
	border: 1px solid rgba(255, 255, 255, 0.5);
	box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6);
}

.btn.danger {
	background: linear-gradient(to bottom, #f49187, #e07a6a);
	color: #1c1917;
	border: 1px solid rgba(244, 145, 135, 0.55);
	box-shadow: 0 4px 14px rgba(244, 145, 135, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.or-row {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.4rem;
	font-size: 0.6rem;
	font-weight: 700;
	letter-spacing: 0.1em;
	text-transform: uppercase;
	color: #a8a29e;
}

.claim-from {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.4rem;
	flex-wrap: wrap;
}

.mini-label {
	font-size: 0.62rem;
	font-weight: 700;
	letter-spacing: 0.08em;
	text-transform: uppercase;
	color: #a8a29e;
	margin-right: 0.3rem;
}

.pill {
	padding: 0.35rem 0.8rem;
	border-radius: 2rem;
	font-size: 0.78rem;
	font-weight: 800;
	transition: transform 0.15s, background 0.2s;
}

.pill:hover {
	transform: scale(1.06);
}

.pill.on {
	color: #1c1917 !important;
}

.pick-row {
	display: flex;
	justify-content: center;
	gap: 0.4rem;
	flex-wrap: wrap;
}

.submit-row {
	display: flex;
	justify-content: center;
}

.waiting {
	border-radius: 1rem;
	padding: 0.7rem 1.6rem;
	font-size: 0.85rem;
	font-weight: 700;
	color: rgba(255, 220, 160, 0.7);
}

.paused-panel {
	align-items: center;
	color: #b45309;
	font-weight: 800;
	font-size: 0.9rem;
}

.pulse {
	animation: pulse 1.3s ease-in-out infinite;
}

@keyframes pulse {
	0%,
	100% {
		box-shadow: 0 0 0 0 rgba(232, 199, 58, 0.55), 0 4px 14px rgba(245, 197, 24, 0.2);
	}
	50% {
		box-shadow: 0 0 0 9px rgba(232, 199, 58, 0), 0 4px 14px rgba(245, 197, 24, 0.2);
	}
}
</style>
