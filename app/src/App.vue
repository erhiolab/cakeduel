<script setup lang="ts">
import {onMounted} from "vue"
import {game} from "./composables/useGame"
import {enableAudio} from "./game/audio"
import StartScreen from "./screens/StartScreen.vue"
import LobbyScreen from "./screens/LobbyScreen.vue"
import GameScreen from "./screens/GameScreen.vue"
import ResultsScreen from "./screens/ResultsScreen.vue"
import LandscapePrompt from "./components/LandscapePrompt.vue"

const {state} = game

const handleFirstClick = () => {
	enableAudio()
}

onMounted(() => {
	window.addEventListener("pointerdown", handleFirstClick, {once: true})
	// 自动连接: 刷新页面后按 token 恢复房间/对局
	game.connect()
})
</script>

<template>
	<div class="app-root" @pointerdown="handleFirstClick" @contextmenu.prevent>
		<StartScreen v-if="state.screen === 'start'" key="start"/>
		<LobbyScreen v-else-if="state.screen === 'lobby'" key="lobby"/>
		<GameScreen v-else-if="state.screen === 'game'" key="game"/>
		<ResultsScreen v-else-if="state.screen === 'results'" key="results"/>
		<LandscapePrompt/>
	</div>
</template>

<style scoped>
.app-root {
	width: 100%;
	height: 100%;
	position: relative;
	overflow: hidden;
}

</style>
