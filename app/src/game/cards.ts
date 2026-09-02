/**
 * 卡牌元数据
 */
export interface CardMeta {
	key: string
	name: string
	color: string
	// sword | fire | shield | sheep | special
	type: string
	desc: string
}

/**
 * 卡牌元数据
 */
export const CARDS: Record<string, CardMeta> = {
	soldier: {
		key: "soldier",
		name: "士兵",
		color: "#C8872A",
		type: "sword",
		desc: "⚔️ · 抢走 1 个蛋糕。",
	},
	archer: {
		key: "archer",
		name: "弓箭手",
		color: "#8FAE5A",
		type: "sword",
		desc: "⚔️ · 抢走 1 个蛋糕。",
	},
	wizard: {
		key: "wizard",
		name: "法师",
		color: "#E07A6A",
		type: "fire",
		desc: "🔥 · 抢走 2 个蛋糕。",
	},
	defender: {
		key: "defender",
		name: "盾卫",
		color: "#7C9FBF",
		type: "shield",
		desc: "🛡️ · 挡住一个来抢蛋糕的⚔️。",
	},
	scientist: {
		key: "scientist",
		name: "科学家",
		color: "#6FA8A4",
		type: "shield",
		desc: "🛡️ · 挡住一个来抢蛋糕的🔥。",
	},
	wolfy: {
		key: "wolfy",
		name: "狼爵士",
		color: "#B7C0C3",
		type: "sheep",
		desc: "🐑 · 不能声明狼爵士。撒谎时可混进打出的牌里，一旦被质疑就必输。回合结束时翻开亮出来。",
	},
	assassin: {
		key: "assassin",
		name: "刺客",
		color: "#5B4A6A",
		type: "sword",
		desc: "⚔️ · 抢走 5 个蛋糕。",
	},
	scout: {
		key: "scout",
		name: "斥候",
		color: "#D9A441",
		type: "sword",
		desc: "⚔️ · 抢 1 个蛋糕，然后再行动一回合。",
	},
	summoner: {
		key: "summoner",
		name: "召唤师",
		color: "#C96BA8",
		type: "fire",
		desc: "🔥 · 指定一张卡牌名称，查看对手手牌，每有一张同名牌就抢 2 块蛋糕。",
	},
	quartermaster: {
		key: "quartermaster",
		name: "军需官",
		color: "#6B8E5A",
		type: "special",
		desc: "⭐ · 本局剩余时间手牌上限 +1。",
	},
	oracle: {
		key: "oracle",
		name: "神谕师",
		color: "#7A7FD1",
		type: "special",
		desc: "⭐ · 查看对手手牌，然后再行动一回合。",
	},
	priest: {
		key: "priest",
		name: "牧师",
		color: "#E5C86B",
		type: "shield",
		desc: "🛡️ · 挡住一张⚔️/🔥/⭐收入你的手牌，取消其效果。",
	},
	angel: {
		key: "angel",
		name: "天使",
		color: "#F2D98C",
		type: "shield",
		desc: "🛡️ · 挡住所有的牌并取消全部效果，然后对手再行动一回合。",
	},
	baacrates: {
		key: "baacrates",
		name: "咩格拉底",
		color: "#D8B57A",
		type: "sheep",
		desc: "🐑 · 咩格拉底教授和狼爵士在本回合变为科学家。",
	},
	agent_u: {
		key: "agent_u",
		name: "特工U",
		color: "#E26DA2",
		type: "sheep",
		desc: "🐑 · 变成对手手牌中的任意一张牌。",
	},
	pierrot: {
		key: "pierrot",
		name: "绵顿",
		color: "#B79BC8",
		type: "sheep",
		desc: "🐑 · 变成你上回合打出的那张⚔️/🔥/⭐的副本。",
	},
}

/**
 * 卡牌键名
 */
export const CARD_KEYS = Object.keys(CARDS)

// 特殊卡列表(可自定义数量)
export const SPECIAL_CARD_NAMES = CARD_KEYS.filter((k) => !["soldier", "archer", "wizard", "defender", "scientist", "wolfy"].includes(k))

// 默认经典卡组(特殊卡数量, 对应后端默认)
export const DEFAULT_DECK: Record<string, number> = {
	assassin: 1,
	scout: 1,
	summoner: 1,
	quartermaster: 1,
	oracle: 1,
	priest: 2,
	angel: 1,
	baacrates: 1,
	agent_u: 1,
	pierrot: 1,
}

/**
 * 获取卡牌图片 URL
 * @param name 卡牌键名
 * @returns 卡牌图片 URL
 */
export const cardImage = (name: string): string => {
	return `/cakeduel/cards-hd/zh-CN/${name}.jpg`
}

/**
 * 获取卡牌小图片 URL
 * @param name 卡牌键名
 * @returns 卡牌小图片 URL
 */
export const cardSmallImage = (name: string): string => {
	return `/cakeduel/cards/zh-CN/${name}.jpg`
}

/**
 * 卡牌背面图片 URL
 */
export const CARD_BACK = "/cakeduel/card-back-hd.jpg"

/**
 * 卡牌动画帧
 */
export const WAGGLE_FRAMES = [
	"/cakeduel/waggle/waggle0.png",
	"/cakeduel/waggle/waggle1.png",
	"/cakeduel/waggle/waggle2.png",
	"/cakeduel/waggle/waggle3.png",
]
