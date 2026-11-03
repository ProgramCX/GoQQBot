# 使用 ARK 模板

旧 ARK 模板以 `template_id` 和 `kv` 组成。已经使用平台 ARK 模板的业务可以通过 `message.Ark` 发送；新业务应先确认目标场景是否支持该模板，不要根据构造器存在推断平台权限。

## 传入完整模板配置

以下函数接收业务已准备好的模板对象，SDK 在构造时保存快照，业务无需重复拼发送路径。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	markdownTypes "github.com/ProgramCX/GoQQBot/pkg/message/types"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func ReplyARK(ctx *types.Context, template markdownTypes.Ark) error {
	return ctx.Reply(message.Ark(template))
}
```

`markdownTypes.Ark` 包含 `TemplateID` 和 `KV []ArkKV`。每个 KV 可以是字符串 `Value`，也可以携带 `Obj []ArkObj`；对象中的键值通过 `ObjKV []ArkObjKV` 表达。具体键名及嵌套结构取决于所使用的平台模板，SDK 不推断模板内容。

## 接收卡片与发送模板分开处理

`ctx.Message()` 中的 `ARK *message.ARKData` 是接收到的卡片，含 `Prompt`、`Type`、`Name` 和任意 JSON 的 `Fields`。它不是 `markdownTypes.Ark`，不能直接交给发送模板构造器。

读取小程序跳转、预览或其他业务字段时，用 `ARK.Field` 按实际类型解码。平台没有提供字段时保持缺失，不应从显示文案虚构 URL。未知字段仍可在消息 `Raw` 中检查，注意其中的敏感数据。

## 另请参阅

[读取富文本卡片](../rich-content.md) · [图文卡片](cards.md) · [官方频道 ARK 消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html)
