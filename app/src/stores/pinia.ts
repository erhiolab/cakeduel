import {createPinia} from "pinia"

// 全局唯一的 Pinia 实例(业务模块在 setup 外用 store 时显式传入)
export const pinia = createPinia()
