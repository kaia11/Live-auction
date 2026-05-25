# NestJS 模块/接口清单草稿

这份清单基于当前项目骨架和修正后的业务模型整理，目标是让你们可以直接按模块分工实现。

核心口径只有一句话：

- 一个直播间展示多个拍品
- 但任一时刻只维护一个当前激活竞拍场次
- 所以后端要围绕 `拍品队列 + 当前场次状态机 + 实时出价与排名` 来拆模块

---

## 1. 建议模块结构

建议最终模块结构如下：

```text
src/
  app.module.ts
  common/
    dto/
    decorators/
    filters/
    guards/
    interceptors/
  prisma/
  redis/
  modules/
    auth/
    users/
    live-rooms/
    auction-items/
    item-queue/
    auction-sessions/
    bids/
    ranking/
    results/
    orders/
    admin/
    websocket/
    logs/
```

说明：

- 你们当前工程里已有 `users / live-rooms / auction-items / auction-sessions / bids / results / websocket / logs`
- 建议继续补上 `auth / item-queue / ranking / orders / admin`

---

## 2. 模块职责清单

## 2.1 `auth`

职责：

- 处理伪登录或 mock 登录
- 解析当前用户
- 提供角色校验能力

建议提供：

- `MockAuthGuard`
- `@CurrentUser()` 装饰器
- `RolesGuard`

## 2.2 `users`

职责：

- 获取当前用户信息
- 获取用户竞拍历史摘要
- 获取用户订单列表

依赖：

- `PrismaService`
- `ResultsService`
- `OrdersService`

## 2.3 `live-rooms`

职责：

- 获取直播间列表
- 获取直播间详情
- 返回当前激活场次摘要
- 返回在线人数、当前拍品、当前状态

依赖：

- `PrismaService`
- `RedisService`
- `AuctionSessionsService`
- `ItemQueueService`

## 2.4 `auction-items`

职责：

- 管理拍品静态信息
- 获取房间内拍品基础信息
- 获取拍品详情信息
- 修改未开始拍品规则

依赖：

- `PrismaService`

## 2.5 `item-queue`

职责：

- 管理某直播间拍品顺序
- 标记 `queued / upcoming / active / finished / cancelled`
- 场次结束后推进到下一个拍品
- 支持主播手动切换到下一件拍品

依赖：

- `PrismaService`
- `RedisService`

## 2.6 `auction-sessions`

职责：

- 创建场次
- 启动场次
- 获取当前场次快照
- 自动延时
- 正常结束
- 封顶成交
- 异常取消

规则约定：

- 允许 `0 元起拍`
- 封顶价按拍品单独设置，不设置则无封顶
- 每次加价必须是该拍品 `incrementStep` 的整数倍
- 最后 `30s` 内有人出价，则在原结束时间基础上 `+30s`
- 延时次数无限
- 达到封顶价直接成交

依赖：

- `PrismaService`
- `RedisService`
- `ItemQueueService`
- `ResultsService`
- `WebsocketGateway`

## 2.7 `bids`

职责：

- 接收正式出价请求
- 幂等校验
- 原子更新当前价和领先者
- 触发排名变化和实时通知
- 为代理出价 / 智能跟随预留扩展点

依赖：

- `PrismaService`
- `RedisService`
- `AuctionSessionsService`
- `RankingService`
- `WebsocketGateway`
- `LogsService`

## 2.8 `ranking`

职责：

- 维护当前场次排行榜
- 提供前 N 名查询
- 提供“我的名次 / 我的最高出价 / 我是否领先”

当前版本建议固定支持：

- 前 `3` 名排行榜
- 我的当前名次
- 我的最高有效出价

依赖：

- `RedisService`
- `PrismaService`

## 2.9 `results`

职责：

- 生成场次结果
- 标记成交 / 流拍 / 取消
- 提供历史结果查询

依赖：

- `PrismaService`
- `OrdersService`

## 2.10 `orders`

职责：

- 根据成交结果生成模拟订单
- 查询用户订单和后台订单列表

当前版本建议先做：

- 模拟订单
- 成交金额
- 买家信息
- 订单状态

真实物流履约先不作为第一版必做能力。

依赖：

- `PrismaService`

## 2.11 `admin`

职责：

- 给主播/商家后台提供统一入口
- 创建拍品
- 调整拍品顺序
- 启动场次
- 修改未开始规则
- 取消异常竞拍
- 手动切换下一件拍品
- 查看后台静态统计
- 查看成交统计卡片
- 查看拍品出价时间线

依赖：

- `AuctionItemsService`
- `ItemQueueService`
- `AuctionSessionsService`
- `OrdersService`

## 2.12 `websocket`

职责：

- 连接管理
- `join_room / leave_room`
- 广播场次变化
- 广播价格变化
- 广播排名变化
- 广播“你被超越了”

依赖：

- `RedisService`

## 2.13 `logs`

职责：

- 记录关键业务操作日志
- 记录异常日志
- 给答辩和排障留证据

依赖：

- `PrismaService`

---

## 3. 推荐控制器清单

## 3.1 用户端控制器

### `UsersController`

建议路由：

- `GET /users/me`
- `GET /users/me/bids`
- `GET /users/me/orders`

### `LiveRoomsController`

建议路由：

- `GET /rooms`
- `GET /rooms/:roomId`
- `GET /rooms/:roomId/online-count`

### `AuctionItemsController`

建议路由：

- `GET /rooms/:roomId/items`
- `GET /rooms/:roomId/items/:itemId`

### `ItemQueueController`

建议路由：

- `GET /rooms/:roomId/queue`
- `POST /admin/rooms/:roomId/queue/next`

### `AuctionSessionsController`

建议路由：

- `GET /rooms/:roomId/current-session`
- `GET /sessions/:sessionId`
- `GET /sessions/:sessionId/countdown`

### `RankingController`

建议路由：

- `GET /sessions/:sessionId/ranking`
- `GET /sessions/:sessionId/ranking/me`

返回建议固定包含：

- 前 3 名榜单
- 我的名次
- 我的最高有效出价

### `BidsController`

建议路由：

- `POST /bids`

### `ResultsController`

建议路由：

- `GET /results/:resultId`
- `GET /rooms/:roomId/results`

### `OrdersController`

建议路由：

- `GET /orders/:orderId`
- `GET /users/me/orders`

## 3.2 主播/后台控制器

### `AdminItemsController`

建议路由：

- `POST /admin/rooms/:roomId/items`
- `PATCH /admin/items/:itemId`
- `DELETE /admin/items/:itemId`

### `AdminQueueController`

建议路由：

- `POST /admin/rooms/:roomId/queue/reorder`
- `POST /admin/rooms/:roomId/queue/:queueId/upcoming`

### `AdminSessionsController`

建议路由：

- `POST /admin/rooms/:roomId/sessions`
- `POST /admin/sessions/:sessionId/start`
- `POST /admin/sessions/:sessionId/cancel`
- `POST /admin/sessions/:sessionId/end`

### `AdminOrdersController`

建议路由：

- `GET /admin/orders`
- `GET /admin/orders/:orderId`

---

## 4. 核心 DTO 建议

下面这些 DTO 建议优先补齐。

## 4.1 出价相关

### `CreateBidDto`

字段建议：

- `roomId: string`
- `itemId: string`
- `sessionId: string`
- `requestId: string`
- `bidPrice: number`

说明：

- 这个 DTO 你们现在已经有了，建议后面补 `@IsUUID()` 或统一 ID 规则校验

### `CreateBidResponseDto`

字段建议：

- `success: boolean`
- `requestId: string`
- `sessionId: string`
- `currentPrice: number`
- `leaderUserId: string | null`
- `myRank: number | null`
- `isLeading: boolean`
- `endTime: string`
- `version: number`
- `message: string`
- `isCeilingReached: boolean`

## 4.2 场次快照相关

### `CurrentSessionSnapshotDto`

字段建议：

- `roomId`
- `sessionId`
- `itemId`
- `itemTitle`
- `itemCoverImage`
- `status`
- `currentPrice`
- `incrementStep`
- `ceilingPrice`
- `leaderUserId`
- `participantCount`
- `endTime`
- `version`
- `myHighestBid`
- `myRank`
- `isLeading`

## 4.3 排名相关

### `RankingEntryDto`

字段建议：

- `userId`
- `nickname`
- `avatar`
- `highestBid`
- `rank`

### `SessionRankingDto`

字段建议：

- `sessionId`
- `topRanks: RankingEntryDto[]`
- `myRank`
- `myHighestBid`
- `isLeading`
- `version`

约定：

- `topRanks` 默认只返回前 `3` 名

## 4.4 后台相关

### `CreateAuctionItemDto`

字段建议：

- `title`
- `coverImage`
- `description`
- `startPrice`
- `incrementStep`
- `ceilingPrice`
- `durationSeconds`
- `extensionSeconds`
- `extensionTriggerSeconds`

规则约定：

- `startPrice` 允许为 `0`
- `ceilingPrice` 为空时表示无封顶
- `incrementStep` 为每个拍品单独配置
- 前端与后端都必须校验：每次出价增量只能是 `incrementStep` 的整数倍

### `ReorderQueueDto`

字段建议：

- `roomId`
- `items: { queueId: string; sortOrder: number }[]`

### `StartSessionDto`

字段建议：

- `sessionId`
- `startTime`

### `SwitchNextQueueItemDto`

字段建议：

- `roomId`
- `currentQueueId`
- `operatorId`
- `reason`

### `ProxyBidConfigDto`

这是后续智能跟随功能的预留 DTO。

字段建议：

- `sessionId`
- `enabled`
- `maxAutoBidPrice`
- `incrementStrategy`

### `CancelSessionDto`

字段建议：

- `reason`

---

## 5. WebSocket 事件清单

## 5.1 客户端发起

- `join_room`
- `leave_room`
- `subscribe_session`
- `unsubscribe_session`

## 5.2 服务端广播

- `room_joined`
- `room_online_count_updated`
- `auction_price_updated`
- `auction_detail_updated`
- `auction_ranking_updated`
- `auction_bid_success`
- `auction_overtaken`
- `auction_extended`
- `auction_session_ended`
- `auction_session_activated`
- `room_item_queue_updated`
- `device_vibrate_signal`

## 5.3 推荐消息结构

所有广播建议至少带：

- `roomId`
- `sessionId`
- `itemId`
- `version`
- `serverTime`

---

## 6. 当前工程与建议结构的映射

你们当前已有：

- `LiveRoomsController`：可保留
- `AuctionSessionsController`：建议把路由从 `GET /rooms/:roomId/current` 调整成 `GET /rooms/:roomId/current-session`
- `BidsController`：可保留为正式出价入口

建议新增：

- `ItemQueueController`
- `RankingController`
- `OrdersController`
- `AdminItemsController`
- `AdminQueueController`
- `AdminSessionsController`
- `AdminOrdersController`
- `AdminStatsController`

---

## 7. 实现优先级

### P0

- `LiveRoomsModule`
- `AuctionItemsModule`
- `AuctionSessionsModule`
- `BidsModule`
- `WebsocketModule`
- `PrismaModule`
- `RedisModule`

### P1

- `ItemQueueModule`
- `RankingModule`
- `ResultsModule`
- `OrdersModule`

### P2

- `AuthModule`
- `AdminModule`
- `LogsModule`

---

## 8. 最小可跑闭环

如果你们现在要最快进入实现，建议先把下面这一条链跑通：

1. `GET /rooms`
2. `GET /rooms/:roomId`
3. `GET /rooms/:roomId/items`
4. `GET /rooms/:roomId/items/:itemId`
5. `GET /rooms/:roomId/current-session`
6. `POST /bids`
7. WebSocket 推 `auction_price_updated`

然后第二步再补：

- `GET /sessions/:sessionId/ranking`
- `GET /admin/stats/overview`
- `GET /admin/stats/timeline`
- `auction_ranking_updated`
- `auction_session_ended`
- `auction_session_activated`
- `room_item_queue_updated`
- `device_vibrate_signal`

---

## 9. 一句话结论

NestJS 这一版不要按“页面”拆，而要按“业务状态流”拆：

- 静态信息归 `rooms/items`
- 队列推进归 `item-queue`
- 当前赛况归 `auction-sessions`
- 并发写入口归 `bids`
- 实时同步归 `websocket`
- 名次能力归 `ranking`

同时要提前预留：

- 智能跟随 / 代理出价
- 设备振动提醒
- 后台统计页

这样后面你们不管是继续写接口，还是让 AI 生成模块代码，都会顺很多。
