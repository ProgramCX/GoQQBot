# 使用 ARK 模板

已有平台 ARK 模板的业务可以通过 `message.Ark` 发送。旧模板由 `template_id` 和 `kv` 组成。

## 传入完整模板配置

以下函数接收业务已准备好的模板对象。SDK 在构造时保存快照，发送时自动选取路径。

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

`markdownTypes.Ark` 包含 `TemplateID` 和 `KV []ArkKV`。每个 KV 可带字符串 `Value` 或对象列表 `Obj []ArkObj`。对象中的键值由 `ObjKV []ArkObjKV` 表达。

## 接收卡片与发送模板分开处理

`ctx.Message()` 的 `ARK *message.ARKData` 保存接收卡片。字段包括 `Prompt`、`Type`、`Name` 和任意 JSON 的 `Fields`。读取小程序跳转、预览或其他字段时，用 `ARK.Field` 按实际类型解码。见[读取富文本卡片](../rich-content.md)。

## 注意事项

- 构造器可用不代表取得平台权限。新业务需确认目标场景支持所选模板。
- 键名和嵌套结构由平台模板决定。SDK 不推断模板内容。
- 接收到的 `ARKData` 不同于 `markdownTypes.Ark`。直接传给发送模板构造器会类型不匹配。
- 平台未提供的字段保持缺失。显示文案无法推出真实 URL。
- 消息 `Raw` 保留未知字段，也可能含敏感数据。

## 另请参阅

[读取富文本卡片](../rich-content.md) · [图文卡片](cards.md) · [官方频道 ARK 消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html)
