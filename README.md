# Live Auction

## 在线访问

- 用户端：`http://47.96.37.164/user/`
- 商家端：`http://47.96.37.164/anchor/`
- 后端接口基地址：`http://47.96.37.164:8080`

一个面向直播竞拍场景的全栈项目，包含：

- `Frontend_user`：用户端前端，负责看直播、出价、查看竞拍状态
- `Frontend_anchor`：商家/主播端前端，负责开播、上架商品、管理场次
- `Backend`：Go 后端，提供鉴权、直播间、商品、场次、出价、结算、WebSocket 实时能力

项目当前以 Go 后端为主，前端通过 Vite 启动或打包部署。

## 依赖环境

建议环境：

- Go `1.22`
- Node.js `18+`
- npm `9+`
- MySQL `8.x`
- Redis `7.x`
- Docker + Docker Compose（推荐，用于一键启动后端依赖）

## 启动步骤

### 方式一：推荐，使用 Docker Compose 启动后端

在项目根目录执行：

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛"
docker-compose up -d --build
```

启动后：

- 后端接口：`http://127.0.0.1:8080`
- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`

查看状态：

```powershell
docker-compose ps
docker-compose logs -f backend
```

### 方式二：本地分别启动后端

1. 先准备 MySQL 和 Redis
2. 如需手动初始化数据库，执行：

```powershell
mysql -uroot -p < .\Backend\mysql\schema.sql
mysql -uroot -p auction_live < .\Backend\mysql\seed.sql
```

3. 启动后端：

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛\Backend"
$env:APP_PORT="8080"
$env:MYSQL_DSN="root:password@tcp(127.0.0.1:3306)/auction_live?charset=utf8mb4&parseTime=True&loc=Local"
$env:REDIS_ADDR="127.0.0.1:6379"
$env:JWT_SECRET="replace_me"
$env:REQUIRE_PERSISTENT_LEDGER="true"
go run ./cmd/server
```

如果只是本地快速演示，没有可用 MySQL，可以切到内存模式：

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛\Backend"
$env:REQUIRE_PERSISTENT_LEDGER="false"
go run ./cmd/server
```

### 启动用户端前端

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛\Frontend_user"
npm install
npm run dev
```

默认开发地址一般为：`http://127.0.0.1:5173`

### 启动商家端前端

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛\Frontend_anchor"
npm install
npm run dev
```

默认开发地址一般为：`http://127.0.0.1:5174`

## 目录结构

```text
.
├─Backend/                 Go 后端
│  ├─cmd/server/           程序入口
│  ├─internal/             核心业务代码
│  ├─mysql/                数据库建表与种子数据
│  ├─scripts/              压测、重置数据等脚本
│  └─deploy/               监控等部署配置
├─Frontend_user/           用户端前端
├─Frontend_anchor/         商家/主播端前端
├─java_backend/            历史目录/预留目录
└─docker-compose.yml       本地容器编排
```

## 配置说明

### 后端环境变量

后端主要读取以下配置：

- `APP_PORT`：服务端口，默认 `8080`
- `MYSQL_DSN`：MySQL 连接串
- `REQUIRE_PERSISTENT_LEDGER`：是否强制使用 MySQL 持久化，默认 `true`
- `REDIS_ADDR`：Redis 地址，默认 `127.0.0.1:6379`
- `REDIS_PASSWORD`：Redis 密码，默认空
- `REDIS_DB`：Redis 数据库编号，默认 `0`
- `JWT_SECRET`：JWT 签名密钥，默认 `replace_me`

### 前端接口地址

两个前端都支持通过 `VITE_API_BASE_URL` 指定后端地址；如果不传，会使用代码里的默认地址。

本地开发示例：

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛\Frontend_user"
$env:VITE_API_BASE_URL="http://127.0.0.1:8080"
npm run dev
```

```powershell
cd "D:\Learning\HKU\sem2\字节竞赛\Frontend_anchor"
$env:VITE_API_BASE_URL="http://127.0.0.1:8080"
npm run dev
```

## 补充说明

- 后端健康检查接口：`GET /health`
- Docker Compose 会自动挂载 `Backend/mysql/schema.sql` 和 `Backend/mysql/seed.sql`
- 如果前端页面能打开但接口失败，优先检查后端是否已启动，以及 `VITE_API_BASE_URL` 是否指向正确地址
