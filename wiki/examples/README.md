# 示例程序与人工验收

本页命令均在 SDK 仓库根目录执行；自己的业务项目按[快速开始](../getting-started/quick-start.md)导入 SDK。配置文件需填写你自己的凭证，不提交到版本库。

## 选择示例

| 示例 | 启动命令 | 交互入口 |
| --- | --- | --- |
| 基础单聊回复 | `go run ./examples/echo -config config.yaml` | 单聊 `/hello` |
| Markdown 菜单 | `go run ./examples/menu -config config.yaml` | `/menu` |
| 消息交互 | `go run ./examples/interactions -config config.yaml` | `/interact`，按钮见示例菜单 |
| 富媒体 | `go run ./examples/media -config config.yaml` | `/media` |
| 图片素材 | `go run ./examples/materials -config config.yaml -image-file ./test/image.jpg` | `/materials` |

群聊基础回复请使用[快速开始中的完整程序](../getting-started/quick-start.md)或 `menu` 示例；当前 echo 示例的群事件注册与该快速开始不同，不用于群 @命令验收。群里需要实际选择机器人进行 @，不能只输入普通文本形式的 `@机器人`。未知命令、图片等接收测试需要对应的 `OnMessage`，不是注册命令后自动监听所有消息。

交互示例需要 `INTERACTION` intent 和平台相应权限。可运行 `go run ./examples/interactions -help` 等命令查看示例当前支持的参数，不把菜单中的命令与操作系统启动参数混淆。

## 本地富媒体完整参数

下面是 PowerShell 命令，文件名按你实际准备的素材修改。反引号必须是该行最后一个字符，后面不要留空格。

```powershell
go run ./examples/media -config config.yaml -debug-events `
  -image-file "./test/image.jpg" `
  -video-file "./test/video.mp4" `
  -voice-file "./test/voice.silk" `
  -file-file "./test/document.pdf" `
  -card-url "https://example.com/page" `
  -card-image "https://example.com/image.jpg"
```

`example.com` 仅作参数格式展示，测试卡片需要替换成实际可访问地址。语音素材的格式、编码及大小须符合 QQ 当前要求；不能靠把音频扩展名改成 `.silk` 完成转码。

| 参数 | 用途 |
| --- | --- |
| `-config` | 机器人 YAML 配置 |
| `-debug-events` | 输出事件类型和路由诊断，不输出完整消息载荷 |
| `-image-file`、`-video-file`、`-voice-file`、`-file-file` | 本地文件分片上传测试 |
| `-image-url`、`-video-url`、`-voice-url`、`-file-url` | 平台从公网 URL 拉取的上传测试 |
| `-card-url`、`-card-image` | 图文卡片目标页和图片 HTTPS 地址 |

启动后在 QQ 发送 `/media` 并选择操作。点击“图片 URL”使用 `-image-url`，点击“图片分片”使用 `-image-file`；配置了其中一个不会自动补齐另一个。未配置的项目会说明缺少哪个参数。

## 记录结果

分别记录单聊、群聊、电脑端和手机端的表现。模拟 HTTP 测试通过不能替代客户端验收。URL 下载失败和本地分片成功也可能同时发生，原因是平台拉取链路与本地上传链路不同。

撤回示例只针对刚刚发送的机器人回复；素材删除会改变该素材后续可用性，验收时不要使用生产素材。素材示例中的当前素材引用只保存在当前进程，不是业务素材数据库。

## 另请参阅

- [富媒体上传](../server-api/messages/rich-media/README.md)
- [素材管理](../server-api/messages/material-management.md)
- [消息交互验收明细](../../docs/interactions-acceptance.md)
- [测试消息发送](../advanced/testing.md)
