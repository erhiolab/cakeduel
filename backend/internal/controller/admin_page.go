package controller

// adminPageHTML 管理员页面(独立轻量页面, 不依赖前端构建产物)
const adminPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>蛋糕对决 · 管理后台</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background:#11161d; color:#e8e2d6; font-family: system-ui, "Microsoft YaHei", sans-serif; min-height:100vh; }
.wrap { max-width: 1100px; margin: 0 auto; padding: 1rem; }
header { display:flex; align-items:center; justify-content:space-between; gap:.6rem; padding:.5rem 0 1rem; }
h1 { font-size:1.15rem; letter-spacing:.05em; }
.pill { display:inline-flex; align-items:center; gap:.4rem; padding:.3rem .7rem; border-radius:2rem; font-size:.72rem; font-weight:700; }
.pill.ok { background:#14532d; color:#86efac; border:1px solid #22c55e55; }
.pill.bad { background:#7f1d1d; color:#fca5a5; border:1px solid #ef444455; }
.pill.gold { background:#713f12; color:#fcd34d; border:1px solid #f59e0b55; }
.card { background:#1b2430; border:1px solid #2c3947; border-radius:.8rem; padding:.8rem .9rem; margin-bottom:.8rem; }
.muted { color:#8d99a6; font-size:.72rem; }
.grid2 { display:grid; grid-template-columns:1fr 1fr; gap:.8rem; }
@media (max-width:760px){ .grid2{ grid-template-columns:1fr; } }
label { display:block; font-size:.78rem; margin-bottom:.35rem; }
input { width:100%; padding:.6rem .7rem; border-radius:.5rem; border:1px solid #3a4a5c; background:#0f1720; color:#fff; font-size:1rem; letter-spacing:.2em; text-align:center; }
button { cursor:pointer; border:0; border-radius:.5rem; padding:.55rem .9rem; font-weight:800; }
.btn-main { background:#b45309; color:#fff; width:100%; margin-top:.6rem; }
.btn-ghost { background:#2c3947; color:#cfd8e0; }
.btn-danger { background:#7f1d1d; color:#fecaca; }
.btn-mini { padding:.28rem .55rem; font-size:.7rem; }
.banner { background:#713f12; border:1px solid #f59e0b66; border-radius:.6rem; padding:.7rem .9rem; margin-bottom:.8rem; text-align:center; font-weight:800; color:#fde68a; }
.password { font-size:1.5rem; letter-spacing:.35em; color:#fff; }
.row { display:flex; flex-wrap:wrap; gap:.45rem; align-items:center; }
.chip { background:#243140; border:1px solid #35485b; border-radius:2rem; padding:.25rem .6rem; font-size:.7rem; }
.chip b { color:#fcd34d; }
.player { border:1px solid #33465a; background:#121a24; border-radius:.6rem; padding:.45rem .6rem; margin-top:.45rem; }
.player .head { display:flex; align-items:center; gap:.5rem; font-weight:800; }
.hand { display:flex; gap:.25rem; flex-wrap:wrap; margin-top:.35rem; }
.hand img { height:3.1rem; border-radius:.3rem; background:#0a0f16; }
.hand .back { width:2.1rem; height:3.1rem; background:#1f2b39; border-radius:.25rem; display:inline-flex; align-items:center; justify-content:center; font-size:.6rem; color:#8d99a6; }
.claims span { color:#93c5fd; }
.chat { margin-top:.5rem; border-top:1px dashed #33465a; padding-top:.45rem; max-height:9rem; overflow:auto; font-size:.75rem; line-height:1.7; }
.chat .from { color:#fcd34d; font-weight:800; }
.refresh { margin-left:auto; }
.empty { color:#8d99a6; text-align:center; padding:1.5rem; }
</style>
</head>
<body>
<div class="wrap">
<header>
	<h1>🎂 蛋糕对决 · 管理后台</h1>
	<div id="headRight"></div>
</header>
<div id="login">
	<div class="card">
		<h2 style="font-size:.95rem;margin-bottom:.6rem;">访问密码</h2>
		<p class="muted" style="margin-bottom:.7rem;">访问后台会在 Redis 生成一个 1 分钟有效的一次性密码（不回显），请管理员自行读取后输入。</p>
		<div class="banner" id="pwBanner" style="display:none;">
			<div>🔐 一次性密码已生成，请从 Redis 读取</div>
			<div class="muted" style="margin-top:.25rem;">Redis 键：<b>cakeduel:admin:password</b> · 剩余 <span id="pwLeft">60</span> 秒</div>
		</div>
		<label for="pwInput">输入密码</label>
		<input id="pwInput" maxlength="8" autocomplete="off" style="letter-spacing:.3em;" placeholder="输入 Redis 中的密码"/>
		<button class="btn-main" id="loginBtn">进入后台</button>
		<div id="loginErr" class="muted" style="color:#fca5a5;margin-top:.5rem;text-align:center;"></div>
	</div>
</div>
<div id="dashboard" style="display:none;">
	<div class="grid2">
		<div class="card">
			<div class="row" style="justify-content:space-between;">
				<b>在线用户 <span id="onlineCount" class="pill ok">0</span></b>
				<span class="muted">房间 <span id="roomCountTop">0</span></span>
			</div>
			<div id="onlineList" style="margin-top:.5rem;"></div>
		</div>
		<div class="card">
			<b>房间列表</b>
			<div id="roomList" style="margin-top:.5rem;"></div>
		</div>
	</div>
</div>
</div>
<script>
var TOKEN = sessionStorage.getItem("cd_admin_token") || "";
var state = { online: [], rooms: [] };
function esc(s){ return String(s == null ? "" : s).replace(/[&<>"']/g, function(m){ return {"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[m]; }); }
function api(path, opt){
	var headers = Object.assign({}, (opt && opt.headers) || {});
	headers["Authorization"] = "Bearer " + TOKEN;
	headers["Content-Type"] = "application/json";
	return fetch(path, Object.assign({}, opt, {headers: headers})).then(function(res){
		return res.json().then(function(data){
			if (res.status === 401 && TOKEN){ TOKEN = ""; sessionStorage.removeItem("cd_admin_token"); showLogin(); }
			if (data.error){ throw new Error(data.message || "请求失败"); }
			return data.body;
		});
	});
}
function showLogin(){ document.getElementById("login").style.display=""; document.getElementById("dashboard").style.display="none"; challenge(); }
function showDash(){ document.getElementById("login").style.display="none"; document.getElementById("dashboard").style.display=""; refreshAll(); }
function challenge(){
	fetch("/api/admin/challenge").then(function(res){ return res.json(); }).then(function(d){
		if (!d.error){
			document.getElementById("pwBanner").style.display="";
			var left = 60;
			document.getElementById("pwLeft").textContent = left;
			var t = setInterval(function(){ left--; document.getElementById("pwLeft").textContent = Math.max(0,left); if (left<=0) clearInterval(t); }, 1000);
		} else {
			document.getElementById("pwBanner").style.display="";
			document.getElementById("pwBanner").textContent = "密码生成失败：Redis 可能不可用";
		}
	}).catch(function(){});
}
document.getElementById("loginBtn").addEventListener("click", function(){
	var pw = document.getElementById("pwInput").value.trim();
	if (!pw) return;
	fetch("/api/admin/verify", {method:"POST", headers:{"Content-Type":"application/json"}, body: JSON.stringify({password: pw})})
		.then(function(res){ return res.json(); })
		.then(function(d){
			if (d.error){ document.getElementById("loginErr").textContent = d.message; return; }
			TOKEN = d.body.token;
			sessionStorage.setItem("cd_admin_token", TOKEN);
			showDash();
		}).catch(function(e){ document.getElementById("loginErr").textContent = String(e); });
});
document.getElementById("pwInput").addEventListener("keydown", function(e){ if (e.key === "Enter") document.getElementById("loginBtn").click(); });
function renderOnline(){
	var el = document.getElementById("onlineList");
	document.getElementById("onlineCount").textContent = state.online.length;
	el.innerHTML = state.online.length ? "" : '<div class="empty">暂无在线用户</div>';
	state.online.forEach(function(u){
		var d = document.createElement("div");
		d.className = "row";
		d.style.cssText = "margin-top:.4rem;background:#121a24;padding:.4rem .6rem;border-radius:.5rem;";
		d.innerHTML = '<span class="chip"><b>' + esc(u.name || "未命名") + "</b></span>" +
			'<span class="chip">' + esc(roleText(u.role)) + "</span>" +
			(u.roomCode ? '<span class="chip">🏠 ' + esc(u.roomCode) + "</span>" : "") +
			'<span class="muted" style="margin-left:auto;">#' + esc(u.token || u.id || "") + "</span>";
		el.appendChild(d);
	});
}
function roleText(role){ return {player:"玩家", spectator:"观战", waiting_spectator:"观战等待", matching:"匹配中", online:"在线"}[role] || role; }
function renderRooms(){
	var el = document.getElementById("roomList");
	document.getElementById("roomCountTop").textContent = state.rooms.length;
	el.innerHTML = state.rooms.length ? "" : '<div class="empty">暂无房间</div>';
	state.rooms.forEach(function(r){
		var box = document.createElement("div");
		box.className = "player";
		var playersHtml = (r.players || []).map(function(p){
			var hand = (p.hand || []).map(function(n){ return '<img src="/cakeduel/cards/zh-CN/' + encodeURIComponent(n) + '.jpg" alt="' + esc(n) + '" title="' + esc(n) + '"/>'; }).join("");
			return '<div class="player"><div class="head">' +
				(p.index === r.attackerIndex ? "⚔️" : "🛡️") + " " + esc(p.name || "空位") +
				'<span class="pill ' + (p.connected ? "ok" : "bad") + '">' + (p.connected ? "在线" : "离线") + "</span>" +
				"<span class='pill gold'>🍰 " + p.cakes + "</span>" +
				"<span class='muted' style='margin-left:auto;'>" + p.handCount + " 张</span>" +
				'</div><div class="hand">' + (hand || '<span class="back">无牌</span>') + "</div></div>";
		}).join("");
		var claims = (r.claims || []).map(function(c){ return "<span>" + esc(c) + "</span>"; }).join(" ");
		var chats = (r.chatHistory || []).map(function(m){ return '<div><span class="from">' + esc(m.name) + ":</span> " + esc(m.text) + "</div>"; }).join("");
		box.innerHTML =
			'<div class="row" style="justify-content:space-between;"><b>🏠 ' + esc(r.code) + "</b>" +
			'<span class="pill gold">' + esc(r.mode) + "</span>" +
			"<span class='pill " + (r.paused ? "bad" : "ok") + "'>" + statusText(r) + "</span></div>" +
			'<div class="row" style="margin-top:.35rem;"><span class="chip">👁 观战 ' + (r.spectatorCount || 0) + "</span>" +
			"<span class='chip'>第 " + r.roundNumber + " 回合</span>" +
			"<span class='chip'>比分 " + scoreText(r) + "</span></div>" +
			'<div class="claims" style="margin-top:.4rem;">' + (claims || '<span class="muted">暂无声明</span>') + "</div>" +
			playersHtml +
			'<div class="chat">' + (chats || '<div class="muted">暂无聊天记录</div>') + "</div>" +
			'<div style="margin-top:.55rem;"><button class="btn-danger btn-mini" onclick="dismiss(\'' + esc(r.code) + '\')">🗑 强制解散</button></div>';
		el.appendChild(box);
	});
}
function statusText(r){
	if (r.gameOver) return "已结束";
	if (r.status === "playing") return r.paused ? "暂停中" : "对局中";
	return "等待中";
}
function scoreText(r){
	var a=0,b=0; (r.boutWinners || []).forEach(function(w){ if (w===0) a++; else b++; });
	return a + " : " + b;
}
function dismiss(code){
	if (!window.confirm("确认解散房间 " + code + " ? 所有玩家将被移出。")) return;
	api("/api/admin/rooms/dismiss", {method:"POST", body: JSON.stringify({code: code})}).then(refreshRooms).catch(function(e){ alert(e.message); });
}
function refreshOnline(){ api("/api/admin/overview").then(function(b){ state.online = b.online || []; renderOnline(); }).catch(function(){}); }
function refreshRooms(){ api("/api/admin/rooms").then(function(b){ state.rooms = b.rooms || []; renderRooms(); }).catch(function(){}); }
function refreshAll(){ refreshOnline(); refreshRooms(); }
var autoTimer = null;
document.getElementById("headRight").innerHTML = '<button class="btn-ghost btn-mini" id="logoutBtn">退出</button>';
document.getElementById("logoutBtn").addEventListener("click", function(){ TOKEN=""; sessionStorage.removeItem("cd_admin_token"); showLogin(); });
if (TOKEN){ showDash(); } else { showLogin(); }
if (autoTimer) clearInterval(autoTimer);
autoTimer = setInterval(function(){ if (TOKEN) refreshAll(); }, 3000);
</script>
</body>
</html>`
