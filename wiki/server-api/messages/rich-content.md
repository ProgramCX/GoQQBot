# 读取图片、表情、卡片与转发

`ctx.Message()` 将平台消息解码为业务可持有的快照，不要求业务依赖 WebSocket、Webhook 或内部路由。SDK 不自动下载附件、追踪网页、上传素材或补造缺失字段。

## 将完整消息交给业务服务

下面注册函数可用于群 @和单聊消息。业务回调只接收 Go 上下文和数据模型，不必认识 SDK 的处理上下文或传输协议。

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

业务服务可将 SDK 模型转换成自己的消息模型再持久化，避免把协议细节扩散到领域代码。返回切片和解码数据属于此次快照，不与原始 SDK 事件共享可变状态。

## 选择数据入口

| 内容 | 入口 | 边界 |
| --- | --- | --- |
| 原始文字、Unicode 表情或平台文本标签 | `m.Content` | 原样保留，不自动把标签转成图片 |
| 当前消息附件 | `m.Attachments` | URL、文件名、MIME、字节数、宽高等，字段依平台下发 |
| 所有层级附件 | `m.AllAttachments()` | 按原消息及递归条目顺序展开，失去独立消息归属 |
| 所有层级图片 | `m.Images()` | 按 MIME 筛选，不根据文件名或 URL 猜测 |
| 语音结果 | 附件的 `VoiceWAVURL`、`ASRReferText` | 仅在平台提供时可用 |
| 卡片、小程序、位置等 | `m.ARK` | 开放 JSON 字段，不保证含跳转或图片 URL |
| 引用、并列或转发条目 | `m.MsgElements` | 递归保留作者、消息索引、正文、附件和 ARK |
| 场景扩展 | `m.MessageScene.Lookup(key)` | 按 `key=value` 提取，值中的等号会保留 |
| 未来未建模字段 | `m.Raw`、条目的 `Raw` | 原始 JSON；可能有敏感授权数据 |

`message.MessageAttachment` 和 `message.MessageScene` 是相应接收模型的兼容别名。旧 `ArkData`、`Elements` 原始字段也保留，新业务优先使用类型化 `ARK` 和 `MsgElements`。

## 读取可选卡片字段

以下纯数据函数不联网。`jump_url` 只是按实际载荷读取的字段名示例，不代表每一种卡片都承诺有此字段。

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

字段缺失返回 `false, nil`；存在但不是预期 JSON 类型时返回解码错误。对于对象或数组，传入对应业务结构，而不是强行当成字符串。

## 图片表情与引用的限制

图片形式的表情包仅在平台实际下发图片附件时可由 `Images()` 读取。QQ 内置表情、小程序或转发内容没有链接时，SDK 没有足够信息生成真实图片地址。合并转发若只下发文字摘要，也无法恢复未提供的原始消息节点。

需要原作者和层级归属时遍历 `MsgElements`，不要只保存扁平附件列表。`msg_idx/ref_msg_idx` 是展示引用上下文，可能出现临时值或关联不稳定，不能保证可靠溯源或作为撤回 ID。

## 另请参阅

[接收单聊](c2c/receive.md) · [接收群聊](group/receive.md) · [引用与撤回](reference-and-recall.md) · [官方单聊接收模型](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/c2c_message_create.html) · [官方群聊接收模型](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_at_message_create.html)
