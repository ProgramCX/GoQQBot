# 发送按钮

使用 `message.Markdown(...).Buttons(...)` 在消息下方显示操作按钮。下面的 `menu` 是完整处理函数，可注册给群聊或单聊命令。

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

命令按钮默认预填输入框。上例的 `AutoSend` 在单聊中保留自动发送，在群聊和频道中默认降级为预填。回调按钮需要另行注册 [按钮处理函数](button-callbacks.md)，链接按钮只在客户端跳转。

## 布局与标识

| 配置 | 含义 |
| --- | --- |
| `Buttons(...)` | 默认两列，按从左到右、从上到下排列 |
| `Columns(n)` | 自动布局列数，范围 1..5 |
| `Row(...)` | 显式追加一行，适合每行数量不同的布局 |
| `Keyboard(raw)` | 使用原始 `interaction.Keyboard` 快照 |
| `KeyboardTemplate(id)` | 引用平台侧已配置的模板键盘 |
| `ID(id)` | 指定稳定按钮 ID，便于回调匹配 |

最多五行，每行最多五个按钮。默认两列最多十个按钮；十四个按钮需要 `Columns(3)`。超限会在 `Build` 或发送前返回错误，不会自动加列、丢弃按钮或拆成多条消息。

`Buttons`、`Row` 和原始或模板键盘是互斥配置。省略 ID 时在单个键盘内生成不冲突的布局 ID；需要长期识别业务动作时应显式设置 ID。显示文字最多十个字符。

在上面的处理函数中，可用以下表达式替换消息构造部分：

```text
message.Markdown("选择操作").
    Row(button.Command("帮助", "/help"), button.Command("绑定", "/bind ")).
    Row(button.Link("官网", "https://example.com"))
```

这是构造表达式，需传入 `ctx.Reply`，不会单独发送。`/bind ` 末尾空格会保留。

## 样式、权限和二次确认

| 方法 | 用途与限制 |
| --- | --- |
| `Primary()`、`Style(style)` | 设置展示样式，实际颜色由客户端决定 |
| `Visited(label)` | 设置点击后的展示文字 |
| `Users(ids...)` | 仅指定平台用户标识可用，不是 QQ 号码 |
| `AdminOnly()`、`All()` | 管理员可用或所有人可用 |
| `Roles(ids...)` | 指定频道身份组，仅频道及频道私信场景可用 |
| `Group(id)` | 回调按钮互斥分组，命令或链接按钮不能设置 |
| `Confirm(text)` | 使用平台默认按钮标题的确认弹窗 |
| `Confirmation(modal)` | 自定义 `interaction.ButtonModal`，包括确认及取消标题 |
| `UnsupportedTips(text)` | 客户端不能展示按钮时的提示 |
| `ClickLimit(n)` | 旧平台字段，已弃用；业务幂等不能依赖它 |

确认正文最多四十个字符、不能包含 HTTP(S) URL；确认和取消标题分别最多四个字符。按钮的展示权限不能替代业务授权，服务端仍应校验操作者和目标资源权限。

## 跨场景使用

`AutoSend`、`Prefill`、`Reply`、`PickImage`、`ChooseChannel` 只适用于命令动作。`Reply()` 表示发送命令时带原消息引用；`PickImage()` 请求单聊选图器；`ChooseChannel()` 请求频道中的频道选择器。

不支持的场景默认移除自动发送、选图器或频道选择器提示，保留预填命令。添加 `Outgoing.Strict()` 会让不支持的配置返回错误，例如：

```text
message.Markdown("选择图片").Strict().Buttons(
    button.Command("选择图片", "/upload ").PickImage(),
)
```

权限不会降级：带频道身份组限制的按钮在群聊中仍返回错误，不会自动变成所有人可用。原始键盘中的已知提示字段也遵循同一场景规则；未知扩展字段会保留。

构造器不可变，链式方法返回独立值；可以复用菜单模板。通过 `button.Raw`、`Keyboard` 或 `WithField("keyboard", value)` 接入高级协议字段时，仍会校验已知的不合法组合。

## 另请参阅

- [按钮回调与 ACK](button-callbacks.md)
- [文本交互](text.md)
- [QQ 官方：单聊发送 keyboard 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html)
- [QQ 官方：群聊发送 keyboard 字段](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html)
