# 频道内容管理

用 `channel.NewClient(client)` 管理频道公告、精华消息、日程和论坛帖子，也可控制语音子频道的播放及麦克风状态。频道客户端复用原有客户端的认证与 HTTP 配置。

## 配置事件订阅

论坛和音频按实际功能选择，示例中的四项无需同时启用。

### 使用 WebSocket

将所需条目合并到 `config.yaml` 现有的 `connection.websocket.intents` 列表，保留其他业务订阅。保存后重启机器人。

```yaml
connection:
  mode: websocket
  websocket:
    intents:
      - FORUMS_EVENT
      - OPEN_FORUM_EVENT
      - AUDIO_ACTION
      - AUDIO_OR_LIVE_CHANNEL_MEMBER
```

[全部 Intent 与下属事件](../../getting-started/events/intents.md)

### 使用 Webhook

在 QQ 机器人运营与管理后台选择 **开发设置 → 事件订阅与回调**。将接入方式设为 **Webhook**。在 **接收事件配置** 选择 **部分接收**，按事件标识勾选下文事件表中要处理的具体事件。例如发帖勾选 `FORUM_THREAD_CREATE`，监听音频开始勾选 `AUDIO_START`。保存设置。已选 **全部接收** 时，无需逐项勾选。

[QQ 机器人运营与管理后台](https://q.qq.com/qqbot/dashboard/) · [配置 HTTPS 回调](../../getting-started/events/webhook.md)

## 官方能力索引

通过频道客户端的方法调用相应接口。

| 模块 | 官方入口 | SDK 方法 |
| --- | --- | --- |
| 公告 | [创建频道公告](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/post_guild_announces.html) | `CreateAnnouncement`、`DeleteAnnouncement` |
| 精华消息 | [添加精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/put_pins_message.html)、[获取精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/get_pins_message.html) | `AddPin`、`ListPins`、`DeletePin` |
| 日程 | [创建日程](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/post_schedule.html)、[日程列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/get_schedules.html) | `CreateSchedule`、`ListSchedules`、`GetSchedule`、`UpdateSchedule`、`DeleteSchedule` |
| 论坛 | [发表帖子](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/put_thread.html)、[帖子列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/get_threads_list.html) | `CreateThread`、`ListThreads`、`GetThread`、`DeleteThread` |
| 音频与上下麦 | [音频控制](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/audio_control.html)、[机器人上麦](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/put_mic.html) | `ControlAudio`、`JoinMic`、`LeaveMic` |

## 接入已有业务

将频道客户端传给负责内容管理的服务。下面的示例把一条已有消息设为精华，并返回更新后的消息 ID 列表。消息处理函数可调用该服务，再根据操作结果回复用户。参见[业务服务接入](../../advanced/business-integration.md)。

```go
package example

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
)

func pinMessage(ctx context.Context, client *api.Client, channelID, messageID string) ([]string, error) {
	pins, err := channel.NewClient(client).AddPin(ctx, channelID, messageID)
	if err != nil {
		return nil, err
	}
	return pins.MessageIDs, nil
}
```

查询当天的日程时，调用 `ListSchedules(ctx, channelID, 0)`。按结束时间筛选时，将起始时间作为 Unix 毫秒传入。

## 接收论坛和音频事件

用 `OnEvent` 注册需要处理的事件，再从上下文读取对应的数据。WebSocket 连接使用下表中的 intent，Webhook 则在开放平台选择事件。

| 事件 | WebSocket intent | 读取方法 |
| --- | --- | --- |
| 论坛主题、评论、回复及审核 | `FORUMS_EVENT` | `ctx.ForumEvent()` |
| 开放论坛主题、评论及回复 | `OPEN_FORUM_EVENT` | `ctx.ForumEvent()` |
| 音频播放及机器人上下麦 | `AUDIO_ACTION` | `ctx.AudioAction()` |
| 用户进出音视频或直播子频道 | `AUDIO_OR_LIVE_CHANNEL_MEMBER` | `ctx.AudioLiveMember()` |

用事件类型区分操作，例如 `message.FORUM_THREAD_CREATE` 表示创建主题。开放论坛只提供频道、子频道和作者标识。普通论坛还可能提供主题、评论、回复或审核数据。标题和正文保留收到的原始 JSON，以兼容平台的字符串及富文本数组。参见[论坛事件](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/forum.html)、[开放论坛事件](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/open_forum.html)及[事件订阅](../../getting-started/events/README.md)。

## 注意事项

- **公告。** 创建消息公告时需同时提供消息 ID 和子频道 ID。推荐子频道列表会整批替换，最多 3 项，且各子频道需对全体成员可见。消息公告和推荐子频道公告会互相替换。
- **删除范围。** `DeleteAnnouncement` 和 `DeletePin` 接受平台保留值 `all`。使用这个值会删除整个频道的全局公告或指定子频道的全部精华。删除推荐子频道公告也使用 `all`。
- **精华与日程。** 子频道最多保留 20 条精华，只有可见消息可设为精华。创建、修改和删除日程需要管理员权限。创建日程时，平台限制每位管理员每天 10 次，每个频道每天 100 次。
- **论坛。** 发帖接口返回的是任务 ID。查询详情时使用帖子 ID。平台当前只向私域机器人开放这些接口，开通后需从频道移除机器人并重新添加。
- **音频。** 这些接口要求平台授予音频权限。使用 `AudioStart` 时传入音频 URL，暂停、继续和停止只传对应操作。
- **旧接口。** 频道公告使用当前的全局公告接口。已废弃的子频道公告接口仍需按历史业务需求评估。
- 频道小程序能力运行在 JS 客户端，超出 Go 服务端 HTTP API 的范围。SDK 不提供客户端环境或页面渲染。

## 另请参阅

[原始 API 扩展](../../advanced/raw-api.md) · [频道消息收发](../messages/channel/messages.md) · [SDK 覆盖范围](../../reference/coverage.md)
