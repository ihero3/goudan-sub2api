#!/usr/bin/env bash
# =============================================================================
# ThreeRouter 一键重启脚本 (Windows Git Bash)
#   - 干净地停止旧的后端/前端循环 (pidfile + 进程镜像 + 端口)
#   - 后端必须以 DATA_DIR 启动，否则会误判为首次运行进入安装向导
#   - 后端崩溃自动重启 (while true)；前端直接 npm run dev
#   - 用法: bash restart.sh
# =============================================================================
set -u

BACKEND_DIR='/d/li/threerouter-sub2api/backend'
FRONTEND_DIR='/d/li/threerouter-sub2api/frontend'
DATA_DIR='D:\li\threerouter-sub2api\backend\.ata'
LOG_DIR='/d/li/threerouter-sub2api'

TS=$(date +'%Y-%m-%d %H:%M:%S')
echo "[restart] $TS begin"

# --- 1. 停止上一次由本脚本启动的后端循环 (pidfile) ---
if [ -f "$BACKEND_DIR/backend.loop.pid" ]; then
  OPID=$(cat "$BACKEND_DIR/backend.loop.pid")
  echo "[restart] kill old backend loop pid=$OPID"
  kill -9 "$OPID" 2>/dev/null || true
  rm -f "$BACKEND_DIR/backend.loop.pid"
fi

# --- 2. 停止上一次由本脚本启动的前端循环 (pidfile) ---
if [ -f "$FRONTEND_DIR/frontend.loop.pid" ]; then
  OPID=$(cat "$FRONTEND_DIR/frontend.loop.pid")
  echo "[restart] kill old frontend loop pid=$OPID"
  kill -9 "$OPID" 2>/dev/null || true
  rm -f "$FRONTEND_DIR/frontend.loop.pid"
fi

# --- 3. 兜底: 杀掉任何残留的 server.exe 与 3000 端口占用 ---
taskkill /F /IM server.exe 2>/dev/null || true
P3=$(netstat -ano 2>/dev/null | grep ':3000' | grep LISTENING | awk '{print $5}' | head -1)
if [ -n "$P3" ]; then
  echo "[restart] kill old frontend port-3000 pid=$P3"
  taskkill /F /PID "$P3" 2>/dev/null || true
fi
sleep 2

# --- 4. 启动后端循环 (脱离终端, 崩溃自动重启) ---
cd "$BACKEND_DIR"
DATA_DIR="$DATA_DIR" nohup bash -c 'while true; do ./server.exe; echo "[loop] server.exe exited $? $(date)"; sleep 1; done' >> "$LOG_DIR/backend.log" 2>&1 &
echo $! > "$BACKEND_DIR/backend.loop.pid"
echo "[restart] backend loop started pid=$(cat $BACKEND_DIR/backend.loop.pid) (DATA_DIR=$DATA_DIR)"

# --- 5. 启动前端 dev server (脱离终端) ---
cd "$FRONTEND_DIR"
nohup npm run dev >> "$LOG_DIR/frontend.log" 2>&1 &
echo $! > "$FRONTEND_DIR/frontend.loop.pid"
echo "[restart] frontend started pid=$(cat $FRONTEND_DIR/frontend.loop.pid)"

# --- 6. 健康检查 ---
sleep 9
echo "[restart] port check:"
if netstat -ano 2>/dev/null | grep ':8080' | grep LISTENING >/dev/null; then
  echo "  backend  8080 LISTENING"
else
  echo "  backend  8080 DOWN  (see $LOG_DIR/backend.log)"
fi
if netstat -ano 2>/dev/null | grep ':3000' | grep LISTENING >/dev/null; then
  echo "  frontend 3000 LISTENING"
else
  echo "  frontend 3000 DOWN  (see $LOG_DIR/frontend.log)"
fi

# --- 7. 预热前端: vite 首次会跑依赖预构建(optimizeDeps), 期间会 hold 住所有请求
#        直到预构建完成才放行。主动触发并在脚本结束前等到真正 200, 避免用户打开时卡在转圈。
echo "[restart] warming up frontend (trigger optimizeDeps)..."
for i in $(seq 1 30); do
  CODE=$(curl -s -m 20 -o /dev/null -w '%{http_code}' --noproxy '*' http://localhost:3000/ 2>/dev/null)
  if [ "$CODE" = "200" ]; then
    echo "  frontend / -> 200 (ready after ${i} tries)"
    break
  fi
  echo "  frontend / -> $CODE (optimizing... try $i/30)"
  sleep 3
done

echo "[restart] done"
