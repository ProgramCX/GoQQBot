# 群聊与频道实机验收

要把新增接口按顺序验收并准备发版，见[发版验收清单](RELEASE-CHECK.md)。

从仓库根目录执行以下 PowerShell 命令。脚本默认读取仓库根目录的 `config.yaml`，也可通过 `-Config` 指定另一份配置。

```powershell
.\examples\management\run.ps1 -List
.\examples\management\run.ps1 -Write -ImageFile .\test\image.jpg
```

第一条列出所有单次操作、位置参数及 JSON 类型。第二条启动监听并开放下列 QQ 测试命令。省略 `-Write` 时只记录接收的事件。执行需要机器人在线的单次操作时，保留监听终端，在第二个终端运行命令。

## 配置事件订阅

WebSocket 模式下，在现有配置的 `connection.websocket.intents` 中按实际测试项目添加下列名称。只配置机器人已获准订阅的事件。Webhook 模式使用 QQ 开放平台的事件订阅配置。

| 验收项目 | 订阅 |
| --- | --- |
| 群里 @机器人并获取群和成员 OpenID | `GROUP_AND_C2C_EVENT` |
| 群成员加入、退出、申请入群 | `GROUP_MEMBER_EVENT` |
| 频道、子频道变更 | `GUILDS` |
| 频道成员变更 | `GUILD_MEMBERS` |
| 子频道 @机器人、频道私信 | `PUBLIC_GUILD_MESSAGES`、`DIRECT_MESSAGE` |
| 频道表态、消息审核 | `GUILD_MESSAGE_REACTIONS`、`MESSAGE_AUDIT` |
| 论坛、开放论坛 | `FORUMS_EVENT`、`OPEN_FORUM_EVENT` |
| 音频及音视频成员进出 | `AUDIO_ACTION`、`AUDIO_OR_LIVE_CHANNEL_MEMBER` |

## 在 QQ 中测试频道消息

先在子频道发送一条普通的 @机器人 消息。终端应显示 `AT_MESSAGE_CREATE`，并打印 `guild_id`、`channel_id`、`user_id`、`message_id`。在频道私信中发送普通消息时，终端应显示 `DIRECT_MESSAGE_CREATE`。私信日志中的 `guild_id` 是私信会话 ID。

| 命令 | 预期结果 |
| --- | --- |
| `/sdk-ping` | 返回文本 |
| `/sdk-legacy` | 使用兼容入口回复，终端打印 Unix 秒时间戳 |
| `/sdk-image` | 使用本地文件发送图片，确认 QQ 客户端能显示 |
| `/sdk-recall` | 发送一条机器人测试回复，3 秒后尝试撤回，并回复成功或已识别的失败原因 |

子频道中发送命令时先通过 QQ 客户端 @机器人。频道私信中可直接发命令。用测试成员给一条频道消息添加或移除表情，终端应分别显示 `MESSAGE_REACTION_ADD` 或 `MESSAGE_REACTION_REMOVE`。

## 单次调用管理与消息 API

用实际收到的字符串 ID 替换下列占位值。群 OpenID、频道 ID、子频道 ID 和频道私信会话 ID 分别使用对应的值。

```powershell
$run = '.\examples\management\run.ps1'
$group = '替换为 group_openid'
$member = '替换为 member_openid'
$guild = '替换为频道 guild_id'
$channel = '替换为子频道 channel_id'
$user = '替换为频道成员 user_id'

# 群聊资料、成员和待审批列表
& $run -Operation group-info -Ids @($group)
& $run -Operation group-member -Ids @($group, $member)
& $run -Operation group-join-requests -Ids @($group) -Body @{limit = 10}

# 频道列表、子频道和权限
& $run -Operation guilds -Body @{Limit = 10}
& $run -Operation channels -Ids @($guild)
& $run -Operation channel-member-permissions -Ids @($channel, $user)

# 发送一条子频道测试消息，并测试消息表态
$sent = & $run -Operation channel-send -Ids @($channel) -Write -Body @{content = '频道测试消息'} | ConvertFrom-Json
if ($LASTEXITCODE -eq 0 -and $sent.id) {
    & $run -Operation channel-reaction-add -Ids @($channel, $sent.id, '203') -Write
    & $run -Operation channel-reaction-users -Ids @($channel, $sent.id, '203')
    & $run -Operation channel-reaction-remove -Ids @($channel, $sent.id, '203') -Write
}

# 创建频道私信会话，再主动发一条测试消息
$dm = & $run -Operation dm-create -Ids @($guild, $user) -Write | ConvertFrom-Json
if ($LASTEXITCODE -eq 0 -and $dm.guild_id) {
    & $run -Operation dm-send -Ids @($dm.guild_id) -Write -Body @{content = '频道私信测试'}
}
```

频道资料或内容写入可以按创建、更新、删除的顺序验收。每次只操作专用的测试资源，并用上一步响应中的 ID。

```powershell
$created = & $run -Operation channel-create -Ids @($guild) -Write -Body @{
    name = 'GoQQBot 验收子频道'
    type = 0
} | ConvertFrom-Json
if ($LASTEXITCODE -eq 0 -and $created.id) {
    & $run -Operation channel-update -Ids @($created.id) -Write -Body @{name = 'GoQQBot 验收改名'}
    & $run -Operation channel -Ids @($created.id)
    & $run -Operation channel-delete -Ids @($created.id) -Write
}
```

群管理写入示例。请使用专用测试群和测试成员。

```powershell
# 对测试成员禁言一分钟，再解除禁言
$until = [DateTimeOffset]::UtcNow.AddMinutes(1).ToString('yyyy-MM-ddTHH:mm:ssZ')
& $run -Operation group-set-mutes -Ids @($group) -Write -Body @(
    @{op = 'add'; member_openid = $member; mute_expire_at = $until}
)
& $run -Operation group-set-mutes -Ids @($group) -Write -Body @(
    @{op = 'del'; member_openid = $member}
)

# 使用监听或申请列表返回的申请 ID 审批一条测试申请
$request = '替换为 join_request_id'
& $run -Operation group-decide-join -Ids @($group, $member) -Write -Body @{
    op = 'approve'
    join_request_id = $request
}
```

## 其他操作与参数

`-Body` 接受 PowerShell 哈希表、数组或标量。`-JsonFile` 可以传入现成的 JSON 文件。两种输入使用 `-List` 所列 Go 类型的 JSON 字段。带 `json:` 标签的字段使用标签名，其余使用 Go 字段名。在仓库根目录可用 `go doc ./pkg/api/group CreateStrategyRequest` 或 `go doc ./pkg/api/channel ScheduleInput` 查看参数。日程的起止时间使用字符串形式的 Unix 毫秒。

| 要测的项目 | 对应操作 |
| --- | --- |
| 群禁言、黑名单、审批策略 | `group-mutes`、`group-blacklist`、`group-strategies`，以及相应写入操作 |
| 频道身份组和禁言 | `guild-roles`、`guild-role-members`、`guild-member-mute`、`guild-members-mute` |
| 频道公告和精华消息 | `guild-create-announcement`、`channel-add-pin`、`channel-pins`，以及相应删除操作 |
| 日程、论坛及音频 | `channel-schedules`、`channel-threads`、`channel-audio`，以及相应创建、修改和删除操作 |
| 图片 URL 或本地图片 | `channel-send` 或 `dm-send`，在参数中填 `image` 或 `image_file` |
| 撤回指定测试消息 | `channel-recall` 或 `dm-recall`，可传 `@{HideTip = $true}` |

也可以直接调用 Go 程序。`--` 后的位置参数按 `-List` 显示的顺序填写。

```powershell
go run ./examples/management -list
go run ./examples/management -config config.yaml -op group-info -- $group
```

## 注意事项

- 群管理和频道管理能力由 QQ 平台授权。遇到拒绝时记录操作名、HTTP 状态、平台错误码及 TraceID。审核响应还可能带有审核 ID。
- 频道撤回返回 `40012010` 时，测试会回复应用缺少接口权限。SDK 可通过 `errors.Is(err, api.ErrAPIAccessDenied)` 识别，原始错误码和 TraceID 仍可通过 `errors.As` 读取。
- 查询结果及事件日志会包含测试成员和会话标识。对外分享前先隐去这些信息。
- 批量操作可能部分成功。检查返回值中的成功或失败列表，再到 QQ 客户端核对。写入或发送结果不明确时先确认平台状态，再决定是否重试。
- 本地图片测试工具限制为 10 MiB。图片格式、权限和展示效果仍以 QQ 客户端实际结果为准。
