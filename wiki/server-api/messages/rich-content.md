# 读取图片、表情、卡片与转发

`ctx.Message()` 把平台消息解码成业务可持有的快照。业务无需依赖 WebSocket、Webhook 或内部路由。

## 将完整消息交给业务服务

下面的注册函数可用于群 @和单聊消息。业务回调只接收 Go 上下文和数据模型。回调无需处理 SDK 上下文或传输协议。

```go
package guide

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/goqqrobot"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func RegisterRichReceiver(bot *goqqrobot.Bot, consume func(context.Context, *message.Received) error) {
	handler := func(ctx *types.Context) error {
		m, err := ctx.Message()
		if err != nil {
			return err
		}
		return consume(ctx.RequestContext(), m)
	}
	bot.OnMessage(message.C2C_MESSAGE_CREATE, handler)
	bot.OnMessage(message.GROUP_AT_MESSAGE_CREATE, handler)
}
```

业务服务可以先把 SDK 模型转为自己的消息模型，再持久化。返回的切片和解码数据归这次快照所有。快照不与原始 SDK 事件共享可变状态。

## 选择数据入口

| 内容 | 入口 | 可读数据 |
| --- | --- | --- |
| 原始文字、Unicode 表情或平台文本标签 | `m.Content` | 原样保留，不自动把标签转成图片 |
| 当前消息附件 | `m.Attachments` | 平台下发的 URL、文件名、MIME、字节数、宽高等 |
| 所有层级附件 | `m.AllAttachments()` | 按原消息及递归条目顺序展开 |
| 所有层级图片 | `m.Images()` | 按 MIME 筛选 |
| 语音结果 | 附件的 `VoiceWAVURL`、`ASRReferText` | 平台提供的语音字段 |
| 卡片、小程序、位置等 | `m.ARK` | 开放的 JSON 字段 |
| 引用、并列或转发条目 | `m.MsgElements` | 递归保留作者、消息索引、正文、附件和 ARK |
| 场景扩展 | `m.MessageScene.Lookup(key)` | 按 `key=value` 提取，保留值里的等号 |
| 尚未建模的字段 | `m.Raw`、条目的 `Raw` | 原始 JSON |

`message.MessageAttachment` 和 `message.MessageScene` 是对应接收模型的兼容别名。SDK 仍保留旧的原始字段 `ArkData`、`Elements`。新业务优先使用类型化的 `ARK` 和 `MsgElements`。

## 读取可选卡片字段

下面的函数只读数据，不联网。`jump_url` 是按实际载荷读取的字段示例。

```go
package guide

import "github.com/ProgramCX/GoQQBot/pkg/message"

func CardJumpURL(m *message.Received) (string, bool, error) {
	if m == nil || m.ARK == nil {
		return "", false, nil
	}
	var target string
	present, err := m.ARK.Field("jump_url", &target)
	return target, present, err
}
```

对象或数组应解码到相应的业务结构。发送用的 [ARK 模板](types/ark.md)另有独立结构。

## 图片表情与引用的限制

需要保留作者和层级归属时，遍历 `MsgElements`。

## 注意事项

- SDK 不自动下载附件、追踪网页、上传素材或补造字段。
- 扁平附件列表丢失消息归属。`Images()` 不从文件名或 URL 推测图片类型。
- 语音字段仅在平台提供时可用。卡片不保证含跳转或图片 URL。
- 字段缺失时返回 `false, nil`。实际 JSON 类型与目标类型不符时返回解码错误。
- 原始 JSON 可能含敏感授权数据。
- 图片表情包只有随图片附件下发时才可用 `Images()` 读取。内置表情、小程序或转发缺少链接时，SDK 无法生成真实图片地址。
- 合并转发只下发文字摘要时，SDK 无法恢复原消息节点。
- `msg_idx/ref_msg_idx` 是展示引用上下文。临时值或不稳定关联无法保证可靠溯源。这两个索引也不用于撤回。

## 另请参阅

[接收单聊](c2c/receive.md) · [接收群聊](group/receive.md) · [引用与撤回](reference-and-recall.md) · [官方单聊接收模型](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html) · [官方群聊接收模型](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html)
