# 运行示例程序

在 SDK 仓库根目录运行示例，可以用 QQ 客户端验证消息和按钮的效果。示例配置使用自己的机器人凭证。若要在独立项目中接入 SDK，请参考[快速开始](../getting-started/quick-start.md)。

## 选择示例

根据要测试的功能选择示例。

| 示例 | 启动命令 | 交互入口 |
| --- | --- | --- |
| 基础单聊回复 | `go run ./examples/echo -config config.yaml` | 单聊 `/hello` |
| Markdown 菜单 | `go run ./examples/menu -config config.yaml` | `/menu` |
| 消息交互 | `go run ./examples/interactions -config config.yaml` | `/interact`，按钮见示例菜单 |
| 富媒体 | `go run ./examples/media -config config.yaml` | `/media` |
| 图片素材 | `go run ./examples/materials -config config.yaml -image-file ./test/image.jpg` | `/materials` |

测试群聊基础回复可用[快速开始中的完整程序](../getting-started/quick-start.md)或 `menu` 示例。

运行 `go run ./examples/interactions -help` 可查看当前启动参数。

## 本地富媒体完整参数

用下面的 PowerShell 命令启动示例。文件名替换成实际素材路径。

```powershell
go run ./examples/media -config config.yaml -debug-events `
  -image-file "./test/image.jpg" `
  -video-file "./test/video.mp4" `
  -voice-file "./test/voice.silk" `
  -file-file "./test/document.pdf" `
  -card-url "https://example.com/page" `
  -card-image "https://example.com/image.jpg"
```



| 参数 | 用途 |
| --- | --- |
| `-config` | 机器人 YAML 配置 |
| `-debug-events` | 输出事件类型和路由诊断，不输出完整消息载荷 |
| `-image-file`、`-video-file`、`-voice-file`、`-file-file` | 本地文件分片上传测试 |
| `-image-url`、`-video-url`、`-voice-url`、`-file-url` | 测试 QQ 从公网 URL 下载并上传素材 |
| `-card-url`、`-card-image` | 图文卡片目标页和图片 HTTPS 地址 |

启动后在 QQ 发送 `/media`，再选择操作。

- “图片 URL”使用 `-image-url`。
- “图片分片”使用 `-image-file`。
- 缺少参数时，示例会提示具体参数。

## 记录结果

分别记录单聊、群聊、电脑端和手机端的表现。



## 注意事项

- 把真实配置提交到版本库会泄露凭证。
- `echo` 的群事件注册与快速开始不同。群 @ 命令使用上文列出的示例。只键入 `@机器人` 文本不会触发实际艾特。
- 未知命令和图片等消息需注册[普通消息处理函数](../getting-started/events/routing.md)。命令注册只处理匹配的命令。
- 交互示例需要 `INTERACTION` intent 和平台权限。QQ 菜单中的命令与程序启动参数分别使用。
- PowerShell 续行反引号后出现空格时，续行会失败。
- `example.com` 用来展示格式。测试卡片时换成可访问地址。语音格式、编码和大小以 QQ 当前要求为准。改名为 `.silk` 不会转码。
- `-image-url` 与 `-image-file` 分别配置。设置其中一个不会补齐另一个。
- **测试范围。** 模拟 HTTP 测试只覆盖请求逻辑，显示效果需在 QQ 客户端验证。URL 上传由 QQ 下载文件，本地分片上传则由 SDK 发送文件。前一种方式失败时，仍可验证后一种方式。
- 撤回示例只撤回刚发送的机器人回复。删除素材会影响后续使用，请使用专门的测试素材。示例只在当前进程保存素材引用，不提供业务素材数据库。

## 另请参阅

[富媒体上传](../server-api/messages/rich-media/README.md) · [素材管理](../server-api/messages/material-management.md) · [消息交互测试明细](../../docs/interactions-acceptance.md) · [测试消息发送](../advanced/testing.md)
