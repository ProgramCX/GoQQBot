# 发送按钮

用 `message.Markdown(...).Buttons(...)` 在消息下方显示按钮。下面的 `menu` 可注册给群聊或单聊命令。

```go
package guide

import (
	"github.com/ProgramCX/GoQQBot/pkg/message"
	"github.com/ProgramCX/GoQQBot/pkg/message/button"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func menu(ctx *types.Context) error {
	return ctx.Reply(message.Markdown("## 操作菜单").Buttons(
		button.Command("帮助", "/help").AutoSend(),
		button.Callback("刷新", "refresh-data").ID("refresh").Primary(),
		button.Link("官网", "https://example.com"),
	))
}
```

命令按钮默认预填输入框。回调按钮需要注册 [按钮处理函数](button-callbacks.md)。链接按钮只在客户端跳转。

## 布局与标识

使用自动布局，或用 `Row` 自定义每行。

| 配置 | 含义 |
| --- | --- |
| `Buttons(...)` | 默认两列，按从左到右、从上到下排列 |
| `Columns(n)` | 自动布局列数，范围 1..5 |
| `Row(...)` | 显式追加一行，适合每行数量不同的布局 |
| `Keyboard(raw)` | 使用原始 `interaction.Keyboard` 快照 |
| `KeyboardTemplate(id)` | 引用平台侧已配置的模板键盘 |
| `ID(id)` | 指定稳定按钮 ID，便于回调匹配 |

省略 ID 时，SDK 在单个键盘内生成不冲突的布局 ID。要长期识别业务动作，可以显式设置 ID。

在上面的处理函数中，可用以下表达式替换消息构造部分。

```text
message.Markdown("选择操作").
    Row(button.Command("帮助", "/help"), button.Command("绑定", "/bind ")).
    Row(button.Link("官网", "https://example.com"))
```

构造表达式不发送消息。将表达式传给 `ctx.Reply` 即可发送。`/bind ` 的末尾空格会保留。

## 样式、权限和二次确认

按操作用途选择样式、适用用户和确认方式。

| 方法 | 用途 |
| --- | --- |
| `Primary()`、`Style(style)` | 设置展示样式 |
| `Visited(label)` | 设置点击后的展示文字 |
| `Users(ids...)` | 允许指定平台用户使用 |
| `AdminOnly()`、`All()` | 管理员可用或所有人可用 |
| `Roles(ids...)` | 指定频道身份组 |
| `Group(id)` | 将回调按钮放进互斥分组 |
| `Confirm(text)` | 使用平台默认按钮标题的确认弹窗 |
| `Confirmation(modal)` | 自定义 `interaction.ButtonModal` 的确认及取消标题 |
| `UnsupportedTips(text)` | 设置客户端无法展示按钮时的提示 |
| `ClickLimit(n)` | 设置旧平台字段 |

## 跨场景使用

为命令按钮选择操作方式。

- `AutoSend()` 请求点击后自动发送。
- `Prefill()` 请求预填输入框。
- `Reply()` 让发送的命令带原消息引用。
- `PickImage()` 请求单聊选图器。
- `ChooseChannel()` 请求频道中的频道选择器。

需要场景不支持时返回错误，可添加 `Outgoing.Strict()`。

```text
message.Markdown("选择图片").Strict().Buttons(
    button.Command("选择图片", "/upload ").PickImage(),
)
```

链式方法返回独立值，构造器保持不变。应用可复用菜单模板。高级字段可通过 `button.Raw`、`Keyboard` 或 [`WithField("keyboard", value)`](../../../advanced/raw-api.md) 传入。

## 注意事项

- 键盘最多五行，每行最多五个按钮。默认两列最多十个（5 行 × 2 列）。十四个按钮需要 `Columns(3)`（5 行 × 3 列最多十五个）。
- 布局超限时，`Build` 或发送入口返回错误。SDK 不自动加列、丢弃按钮或拆成多条消息。
- `Buttons`、`Row`、原始键盘和模板键盘只能选一种。按钮显示文字最多十个字符。
- 客户端决定实际显示颜色。`Users` 接收平台用户标识，QQ 号码不适用。
- `Roles` 仅支持频道和频道私信。带频道身份组限制的按钮在群聊中返回错误。SDK 不会改成所有人可用。
- `Group` 仅支持回调按钮。命令和链接按钮不支持该配置。
- 确认正文最多四十个字符，包含 HTTP(S) URL 会超出支持范围。确认和取消标题分别最多四个字符。
- 展示权限只控制按钮展示。服务端仍要校验操作者和目标资源权限。
- `ClickLimit` 已弃用。业务幂等需要独立实现。
- 上述五种操作只适用于命令动作。上例的 `AutoSend` 在单聊保留自动发送，在群聊和频道退回预填。
- 场景不支持时，SDK 保留预填命令。自动发送、选图器或频道选择器提示会移除。
- 原始键盘中的已知提示字段也受场景规则约束。SDK 保留未知扩展字段，并校验已知的不合法组合。

## 另请参阅

[按钮回调与 ACK](button-callbacks.md) · [文本交互](text.md) · [QQ 官方：单聊发送 keyboard 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html) · [QQ 官方：群聊发送 keyboard 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
