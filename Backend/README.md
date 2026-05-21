# Backend

这是直播竞拍系统的后端骨架。

当前已经搭好的内容：

- `NestJS` 风格目录结构
- 按业务模块拆分的 `src/modules`
- `Prisma` 和 `Redis` 基础占位
- `WebSocket` 网关骨架
- 公共层 `common`

后续建议顺序：

1. 安装依赖
2. 配置 `.env`
3. 补 `prisma/schema.prisma`
4. 先实现直播间和拍品读取接口
5. 再实现出价主流程
