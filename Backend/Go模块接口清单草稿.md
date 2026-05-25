# Go 模块/接口清单草稿

这份清单基于当前 Go 后端骨架整理，目标是让你们后面实现时，直接按 package 和 handler 分工。

核心口径不变：

- 一个直播间会展示多个拍品
- 但同一时刻只允许一个当前激活竞拍场次
- 所以后端核心仍然是 `拍品队列 + 场次状态机 + 实时出价`

---

## 1. 建议目录结构

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
```

分层职责：

- `handler`：接收 HTTP 请求，做参数解析和响应组装
- `service`：承载业务规则和状态流转
- `repository`：封装 MySQL / Redis 访问
- `ws`：直播间订阅、广播、定向推送

---

## 2. 包职责清单

## 2.1 `config`

职责：

- 读取环境变量
- 组装运行配置
- 为 `MySQL / Redis / JWT / WebSocket` 预留统一入口

建议提供：

- `type Config struct`
- `func Load() Config`

## 2.2 `app`

职责：

- 装配配置、service、handler、router
- 统一管理 HTTP Server 生命周期

建议提供：

- `type App struct`
- `func New(cfg config.Config) *App`
- `func (a *App) Run() error`

## 2.3 `model`

职责：

- 定义领域模型和 DTO 草稿
- 统一直播间、拍品、场次、出价、排行榜、结果等结构

建议至少保留：

- `User`
- `LiveRoom`
- `AuctionItem`
- `AuctionSession`
- `Bid`
- `RankingEntry`
- `AuctionResult`

## 2.4 `repository`

职责：

- 抽象持久化接口，避免 service 直接依赖数据库实现
- 后续可以分别落成 `mysql` 与 `redis` 实现

建议至少拆成：

- `RoomRepository`
- `ItemRepository`
- `SessionRepository`
- `BidRepository`
- `RankingRepository`
- `ResultRepository`

## 2.5 `service`

职责：

- 承载全部核心业务规则
- 保证出价校验、自动延时、封顶成交、队列推进的正确性

建议拆分：

- `RoomService`
- `ItemService`
- `SessionService`
- `BidService`
- `AdminService`

规则要内化在 service 里：

- 支持 `0 元起拍`
- 封顶价按拍品单独设置，不设置则无封顶
- 每次出价增量必须是 `incrementStep` 的整数倍
- 最后 `30s` 内有人出价，则在当前结束时间基础上 `+30s`
- 延时次数无限
- 达到封顶价直接成交
- 一个直播间同一时刻只能拍一个拍品

## 2.6 `http/handler`

职责：

- 对外暴露 REST API
- 做参数解析、状态码处理、错误返回

建议 handler：

- `AuthHandler`
- `HealthHandler`
- `RoomHandler`
- `ItemHandler`
- `SessionHandler`
- `BidHandler`
- `AdminHandler`

## 2.7 `ws`

职责：

- 房间订阅与退订
- 实时广播当前价、排行、延时、结束、切场
- 给客户端下发“被超越提醒”和“振动信号”

建议事件名：

- `auction_price_updated`
- `auction_ranking_updated`
- `auction_extended`
- `auction_session_ended`
- `auction_session_activated`
- `room_item_queue_updated`
- `device_vibrate_signal`

---

## 3. REST 接口草稿

## 3.1 用户端

- `POST /auth/login`
- `GET /users/me`
- `GET /health`
- `GET /rooms`
- `GET /rooms/{roomId}`
- `GET /rooms/{roomId}/items`
- `GET /rooms/{roomId}/items/{itemId}`
- `GET /rooms/{roomId}/current-session`
- `GET /sessions/{sessionId}/ranking`
- `GET /users/me/bids`
- `POST /bids`

### 3.1.1 登录接口

建议第一版统一为手机号密码登录：

- `POST /auth/login`

请求体：

```json
{
  "phone": "13800138000",
  "password": "123456"
}
```

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "jwt-token-demo",
    "user": {
      "id": 1001,
      "nickname": "玉友小周",
      "avatar": "https://example.com/avatar.png",
      "phone": "13800138000"
    }
  }
}
```

失败响应建议：

```json
{
  "code": 10001,
  "message": "手机号或密码错误"
}
```

补充约定：

- 前端登录成功后保存 `token`
- 后续用户端接口统一通过 `Authorization: Bearer <token>` 传递登录态
- 如果当前版本后端尚未完成鉴权，也建议先保留该接口返回结构，方便前端先接 mock

### 3.1.2 当前用户信息接口

- `GET /users/me`

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1001,
    "nickname": "玉友小周",
    "avatar": "https://example.com/avatar.png",
    "phone": "13800138000"
  }
}
```

### 3.1.3 直播间拍品列表接口

- `GET /rooms/{roomId}/items`

说明：

- 返回该直播间全部拍品队列
- 用于前端渲染拍品抽屉、详情入口和状态标签
- 同一时刻允许多个拍品可见，但只能有一个拍品对应当前激活场次

每个拍品建议至少包含：

- `itemId`
- `sessionId`
- `title`
- `coverUrl`
- `startPrice`
- `currentPrice`
- `incrementStep`
- `ceilingPrice`
- `status`
- `countdownSeconds`
- `isCurrent`

其中 `status` 建议统一枚举：

- `pending`：待上架
- `not_started`：未开始
- `upcoming`：即将开始
- `auctioning`：竞拍中
- `sold`：已成交
- `failed`：已流拍
- `cancelled`：已取消
- `ended`：已结束

前端按钮映射约定：

- `auctioning`：显示 `立即出价`
- `sold` / `failed` / `cancelled` / `ended`：显示 `查看拍卖结果`
- `pending` / `not_started` / `upcoming`：显示 `查看倒计时`

### 3.1.4 当前激活场次接口

- `GET /rooms/{roomId}/current-session`

说明：

- 返回当前直播间真正处于实时竞拍驱动中的那个场次
- 前端直播页中的悬浮卡、当前价、倒计时、排行榜、我的出价状态都以它为准

成功响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "sessionId": 9001,
    "roomId": 101,
    "itemId": 3001,
    "status": "auctioning",
    "title": "金镶玉平安扣和田玉吊坠项链首饰",
    "coverUrl": "https://example.com/item-cover.png",
    "startPrice": 0,
    "currentPrice": 850,
    "incrementStep": 50,
    "ceilingPrice": 3000,
    "countdownSeconds": 560,
    "endTime": "2026-05-25T18:30:00+08:00",
    "isExtended": false,
    "top3": [
      { "rank": 1, "userId": 1001, "nickname": "玉友小周", "bidPrice": 850 },
      { "rank": 2, "userId": 1008, "nickname": "珠珠", "bidPrice": 800 },
      { "rank": 3, "userId": 1016, "nickname": "阿莹", "bidPrice": 750 }
    ],
    "myBid": {
      "userId": 1001,
      "highestBid": 850,
      "rank": 1,
      "isLeading": true
    }
  }
}
```

如果当前没有激活场次，建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## 3.2 主播/后台

- `POST /admin/rooms/{roomId}/items`
- `PATCH /admin/items/{itemId}`
- `POST /admin/rooms/{roomId}/queue/reorder`
- `POST /admin/rooms/{roomId}/queue/next`
- `POST /admin/sessions/{sessionId}/start`
- `POST /admin/sessions/{sessionId}/cancel`
- `GET /admin/rooms/{roomId}/sessions`
- `GET /admin/orders`
- `GET /admin/stats/overview`
- `GET /admin/stats/timeline`

---

## 4. 重点实现建议

## 4.1 第一优先级

- 跑通直播间列表
- 跑通拍品列表与详情
- 跑通当前场次快照
- 跑通 `POST /bids`

## 4.2 第二优先级

- Redis 当前场次状态
- 排行榜缓存
- 自动延时
- 封顶成交

## 4.3 第三优先级

- 队列推进
- 主播手动切换
- 结果生成
- 模拟订单

## 4.4 第四优先级

- WebSocket 广播
- 被超越提醒
- 设备振动信号
- 智能跟随 / 自动加价扩展位

---

## 5. 和 Prisma 草稿的关系

`prisma/schema.prisma` 当前不再作为 Go 运行时 ORM 配置，而是作为数据库建模参考文档保留。

也就是说：

- Go 真正落地时，你们可以用 `GORM` / `sqlx`
- 但表结构设计仍然可以先参考这份 `Prisma` 草稿
- 等数据库设计完全稳定后，再决定是否保留 Prisma 文档
