# 频道内容管理

频道内容管理使用官方的专用接口。SDK 尚未封装公告、精华消息、日程、论坛帖子及音频控制。

## 官方能力索引

按任务查找官方接口和封装状态。

| 模块 | 官方入口 | SDK 状态 |
| --- | --- | --- |
| 公告 | [创建频道公告](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/post_guild_announces.html) | 尚未封装 |
| 精华消息 | [添加精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/put_pins_message.html)、[获取精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/get_pins_message.html) | 尚未封装 |
| 日程 | [创建日程](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/post_schedule.html)、[日程列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/get_schedules.html) | 尚未封装 |
| 论坛 | [发表帖子](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/put_thread.html)、[帖子列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/get_threads_list.html) | 尚未封装 |
| 音频与上下麦 | [音频控制](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/audio_control.html)、[机器人上麦](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/put_mic.html) | 尚未封装 |

## 接入已有业务

将客户端注入内容服务。服务负责请求模型、分页、修改权限和取消上下文。处理函数调用服务，再根据业务结果回复用户。具体写法见[业务服务接入](../../advanced/business-integration.md)。



## 注意事项

- 发送消息无法替代内容管理接口。新业务应选择官方仍支持的接口。
- 频道小程序能力运行在 JS 客户端，超出 Go 服务端 HTTP API 的范围。SDK 不提供客户端环境或页面渲染。

## 另请参阅

[原始 API 扩展](../../advanced/raw-api.md) · [频道消息收发](../messages/channel/messages.md) · [SDK 覆盖范围](../../reference/coverage.md)
