import {reactive} from "vue"

/**
 * 共享的出牌/选牌选择状态(GameScreen 与 ActionBar 共用)
 */
export const selection = reactive<{
	ids: Set<number>
	claim: string
	pick: number | null
}>({
	ids: new Set(),
	claim: "",
	pick: null,
})

/**
 * 切换卡牌选择状态
 * @param entityId 卡牌 ID
 * @returns 卡牌选择状态
 */
export const toggleCardSelection = (entityId: number) => {
	const NEXT = new Set(selection.ids)
	if (NEXT.has(entityId)) NEXT.delete(entityId)
	else NEXT.add(entityId)
	selection.ids = NEXT
}

/**
 * 重置选择状态
 */
export const resetSelection = () => {
	selection.ids = new Set()
	selection.claim = ""
	selection.pick = null
}
