import {ref} from "vue"
import {defineStore} from "pinia"
import {pinia} from "../stores/pinia"

/**
 * 共享的出牌/选牌选择状态(GameScreen 与 ActionBar 共用)
 */
export const useSelectionStore = defineStore("selection", () => {
	// 选中的手牌实体 ID
	const ids = ref<Set<number>>(new Set())

	// 选中的声明牌名
	const claim = ref("")

	// 选中的召唤师选项索引
	const pick = ref<number | null>(null)

	/**
	 * 切换卡牌选择状态
	 * @param entityId 卡牌 ID
	 */
	const toggleCardSelection = (entityId: number) => {
		const NEXT = new Set(ids.value)
		if (NEXT.has(entityId)) NEXT.delete(entityId)
		else NEXT.add(entityId)
		ids.value = NEXT
	}

	/**
	 * 重置选择状态
	 */
	const resetSelection = () => {
		ids.value = new Set()
		claim.value = ""
		pick.value = null
	}

	return {ids, claim, pick, toggleCardSelection, resetSelection}
})

// 单例实例(组件直接使用)
export const selection = useSelectionStore(pinia)

// 兼容原有解构导入: store action 为闭包函数, 可直接解构调用
export const {toggleCardSelection, resetSelection} = selection
