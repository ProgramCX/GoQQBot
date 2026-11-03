# 业务接入与扩展

常规回复使用 `Context`，独立后台服务使用 `api.Client`，纯消息构造使用 `message`。这三种入口可以组合，不需要业务导入 `internal`。

- [将机器人接入业务服务](business-integration.md)：handler 保持简短，把业务逻辑和持久化交给自己的服务。
- [调用未封装的官方接口](raw-api.md)：认证、URL 编码、自定义响应模型及消息字段扩展。
- [测试消息发送](testing.md)：注入本地 HTTP 替身，避免单元测试访问 QQ。
- [示例程序与验收](../examples/README.md)：使用真实机器人验证客户端行为。

扩展点是显式能力，不表示 SDK 已完整封装所有官方模块。接口覆盖见[覆盖范围](../reference/coverage.md)。
