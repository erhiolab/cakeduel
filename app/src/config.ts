// 蛋糕对决前端配置
//
// wsUrl 说明:
//   - 留空(推荐): 生产环境自动使用当前站点同源(后端托管 dist 时无需任何修改);
//     开发环境(vite dev)自动使用 ws://127.0.0.1:8080
//   - 手动指定: 填写完整地址, 例如 'ws://192.168.1.100:8080' 或 'wss://game.example.com'
export const appConfig = {
	wsUrl: "",
}
