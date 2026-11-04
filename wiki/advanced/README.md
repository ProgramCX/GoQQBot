# 业务接入与扩展

按场景选择 SDK 入口。

- 常规回复使用 `Context`。
- 独立后台服务使用 `api.Client`。
- 只构造消息时使用 `message`。

三种入口可以组合使用，无需导入 `internal`。

- [将机器人接入业务服务](business-integration.md)。精简处理函数，把业务逻辑和持久化交给自己的服务。
- [调用未封装的官方接口](raw-api.md)。配置认证、URL 编码、自定义响应模型和消息字段。
- [测试消息发送](testing.md)。注入本地 HTTP 替身，让单元测试离线运行。
- [示例程序与验收](../examples/README.md)。用真实机器人验证客户端行为。

## 注意事项

扩展入口由调用方显式使用。SDK 尚未完整封装所有官方模块，已支持的接口见[覆盖范围](../reference/coverage.md)。

## 另请参阅

[API 调用指南](../getting-started/api-calls.md)
