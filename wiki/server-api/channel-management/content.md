# 频道内容管理

此分类对应官方的公告、精华消息、日程、论坛帖子和音频控制。当前 SDK 没有这些能力的领域封装；消息发送接口不能替代内容管理接口。

## 官方能力索引

| 模块 | 官方入口 | SDK 状态 |
| --- | --- | --- |
| 公告 | [创建频道公告](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/post_guild_announces.html) | 尚未封装 |
| 精华消息 | [添加精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/put_pins_message.html)、[获取精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/get_pins_message.html) | 尚未封装 |
| 日程 | [创建日程](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/post_schedule.html)、[日程列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/get_schedules.html) | 尚未封装 |
| 论坛 | [发表帖子](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/put_thread.html)、[帖子列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/get_threads_list.html) | 尚未封装 |
| 音频与上下麦 | [音频控制](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/audio_control.html)、[机器人上麦](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/put_mic.html) | 尚未封装 |

## 接入已有业务

将客户端注入一个内容服务，由服务管理请求模型、分页、修改权限和取消上下文。服务返回业务结果，handler 仅负责调用服务和回复用户。具体写法见[业务服务接入](../../advanced/business-integration.md)。

官方页面标为已弃用的接口不应作为新业务入口。客户端 JS 的频道小程序能力也不是 Go 服务端 HTTP API；此 SDK 不提供客户端环境或页面渲染能力。

## 另请参阅

- [原始 API 扩展](../../advanced/raw-api.md)
- [频道消息收发](../messages/channel/messages.md)
- [SDK 覆盖范围](../../reference/coverage.md)
