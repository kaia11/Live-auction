# Backend

这是直播竞拍系统的 Go 后端原型。

当前仓库的目标是先把后端主链路、竞拍规则和商家端基础管理能力跑通，再逐步接入 `MySQL`、`Redis` 和真实 `WebSocket`。

## 当前技术栈

- 语言：`Go 1.25.3`
- HTTP：标准库 `net/http`
- 架构：`handler -> service -> repository`
- 当前数据层：内存态原型
- 后续数据层：`MySQL + Redis`

## 当前已实现

- 统一配置读取：`internal/config`
- 统一响应结构：`code / message / data / serverTime`
- 统一错误码：`internal/http/error_codes.go`
- 健康检查：`GET /health`
- 用户端核心接口：
  - 直播间列表 / 详情
  - 拍品列表 / 详情
  - 当前竞拍场次
  - 排行榜
  - 我的竞拍状态
  - 我的出价记录
- 出价主链路原型：
  - 房间、拍品、场次归属校验
  - `requestId` 去重
  - 当前价、领先者、参与人数更新
  - 出价记录写入
- 竞拍规则原型：
  - `0 元起拍`
  - 每次加价必须是步长整数倍
  - 最后 `30s` 自动延时 `30s`
  - 达到封顶价直接成交
- 商家端基础管理原型：
  - 创建拍品
  - 修改未开始拍品
  - 队列重排
  - 手动切换下一件
  - 手动开始 / 取消场次

## 当前目录结构

```text
Backend/
  cmd/
    server/
      main.go
  internal/
    app/
    config/
    http/
      handler/
    model/
    repository/
    service/
    ws/
  prisma/
    schema.prisma
  .env.example
  go.mod
```

## Go 安装位置

当前机器上的 Go 已安装在：

```text
D:\Tools\Go\go
```

可执行文件路径：

```text
D:\Tools\Go\go\bin\go.exe
```

当前用户环境变量已经配置：

- `GOROOT=D:\Tools\Go\go`
- `Path` 已追加 `D:\Tools\Go\go\bin`

如果新开的终端还找不到 `go`，关闭终端重新打开一次即可。

## 启动方式

在项目根目录进入 `Backend/` 后执行：

```bash
go mod tidy
go fmt ./...
go run ./cmd/server
```

如果当前终端还没有刷新环境变量，也可以直接用绝对路径运行：

```powershell
D:\Tools\Go\go\bin\go.exe mod tidy
D:\Tools\Go\go\bin\go.exe fmt ./...
D:\Tools\Go\go\bin\go.exe run ./cmd/server
```

## 环境变量

参考 [Backend/.env.example](/abs/path/d:/Learning/HKU/sem2/字节竞赛/Backend/.env.example)：

```env
APP_ENV=development
APP_PORT=8080
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/auction_live?charset=utf8mb4&parseTime=True&loc=Local
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
JWT_SECRET=replace_me
WS_ALLOWED_ORIGIN=*
```

说明：

- 当前版本启动时依赖 `Redis`
- 如果配置了 `MYSQL_DSN` 且 MySQL 可连通，查询和关键业务写入会优先走 MySQL
- 鉴权使用 `Authorization: Bearer <token>`，密钥由 `JWT_SECRET` 控制

## 接口调试建议

建议先按下面顺序测试：

```text
GET  /health
GET  /rooms
GET  /rooms/room-001
GET  /rooms/room-001/items
GET  /rooms/room-001/current-session
GET  /sessions/session-001/ranking
GET  /sessions/session-001/my-status?userId=user-001
POST /bids
GET  /users/me/bids?userId=user-001
GET  /rooms/room-001/events
```

商家端建议再测：

```text
POST  /admin/rooms/room-001/items
PATCH /admin/items/{itemId}
POST  /admin/rooms/room-001/queue/reorder
POST  /admin/rooms/room-001/queue/next
POST  /admin/sessions/{sessionId}/start
POST  /admin/sessions/{sessionId}/cancel
GET   /admin/rooms/room-001/sessions
GET   /admin/orders
GET   /admin/stats/overview
GET   /admin/stats/timeline
```

## 当前版本说明

- 当前是“后端原型版”，重点是验证业务规则和联调流程
- 当前不是正式高并发版本
- 当前已经具备 `Redis + MySQL + WebSocket + JWT` 的第一版主链路
- 当前仍然缺少压测、监控、CI 等生产闭环能力

## 接口文档

- 正式接口文档：`Backend/openapi.yaml`
- 权限矩阵：`Backend/接口权限矩阵.md`

## 观测与压测

- 指标端点：`GET /metrics`
- Prometheus：`docker compose up -d prometheus`
- 基础压测脚本：
  - `Backend/scripts/load_bid_test.sh`
  - `Backend/scripts/load_bid_test.ps1`

## 演示数据重置

- WSL / Linux:

```bash
sh ./Backend/scripts/reset_demo_data.sh
```

- Windows PowerShell（调用 WSL 内脚本）:

```powershell
.\Backend\scripts\reset_demo_data.ps1
```

## 后续计划

1. 补集成测试与压测
2. 补监控、告警和部署闭环
3. 继续增强高并发生产化能力
