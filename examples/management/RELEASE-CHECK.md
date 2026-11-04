# 本轮 SDK 发版验收

## 当前实机进度

机器人管理新增接口已向 QQ 发起实测。资料、分享链接、菜单读取和空菜单写回均成功。面板列表在四个场景返回成功。指定测试群的临时面板完成创建、详情读取、修改、关联对象增删和删除。空菜单写回只验证写入接口，菜单展示和点击仍需客户端测试。

测试群的停用自动审批策略已完成创建、修改和删除。测试频道的临时身份组已完成创建、改名和删除。以普通成员复测后，设置禁言和解除禁言均成功，查询也确认了状态变化。修改禁言时长的请求返回成功，但随后查询到的到期时间与请求值不一致，仍需复测。原管理员测试对象返回 `50105002`。

本轮已通过 20 项查询、群入群审批的同意和拒绝，以及频道私信图片发送。群成员列表、详情、黑名单、频道成员列表和频道私信撤回返回 `40012010`，成功路径待有权限的应用复测。

本轮跳过子频道实机测试。依赖子频道的发送、撤回、表态、日程、论坛写入和音频操作均按跳过记录。已取得的子频道相关查询结果仍有效。

剩余可按测试资源安排的项目包括群禁言与解除、群成员移除及黑名单写入、自动审批策略全流程，以及频道身份组、频道成员管理等写入。涉及对象权限、发帖或收发消息的事件，应另记事件是否实际投递，HTTP 成功不能代替事件验收。

先运行只读检查。管理写入和 QQ 客户端动作按下面的流程逐项核对。每项记录为通过、权限受限、缺少测试资源或失败。测试群和频道应由你管理。禁言、移除和审批使用专门的测试账号。

## 一次跑完只读接口

保持原来的验收监听窗口。另开 PowerShell，在仓库根目录运行以下命令，把事件日志或查询结果中的 ID 填进去。未知值可留空，之后补填再执行。

```powershell
$run = '.\examples\management\run.ps1'
& $run -Operation guilds

$ids = @{
    group_openid = $group
    member_openid = $member
    guild_id = ''
    channel_id = ''
    user_id = ''
    role_id = ''
    message_id = ''
    schedule_channel_id = ''
    schedule_id = ''
    forum_channel_id = ''
    thread_id = ''
    audio_channel_id = ''
}

$report = & $run -Check -Body $ids | ConvertFrom-Json
$report | Where-Object status -ne 'manual' | Format-Table -AutoSize
$report | Group-Object status | Select-Object Name, Count
```

- `pass` 表示 QQ 返回成功，SDK 成功解析。
- `permission_blocked` 表示 QQ 拒绝接口或对象权限。保留错误码和 TraceID。
- `missing_fixture` 表示缺少该操作所需的 ID。填写后重跑。
- `fail` 表示需要检查请求、平台状态或响应解析。
- `manual` 表示写入操作，执行时要单独加 `-Write`。

取 ID 的顺序如下：

1. 在 QQ 群 @机器人发一句话。监听日志给出 `group_openid` 和 `member_openid`。
2. `guilds` 给出 `guild_id`。运行 `channels`，传入频道 ID，按名称区分文字、日程、论坛和音视频子频道。
3. 在文字子频道 @机器人发普通消息。监听日志给出 `channel_id`、`user_id` 和 `message_id`。
4. `guild-roles` 给出 `role_id`。日程和帖子 ID 可从 `channel-schedules`、`channel-threads` 取得。

频道私信的 `guild_id` 是私信会话 ID，单独用于 `dm-send` 和 `dm-recall`。

## 群聊管理写入

先看对应查询能否通过。群成员详情已实测返回 `40012010`，记为权限受限。其他接口分别记录结果。

| 测试顺序 | 操作 | QQ 或响应核对 |
| --- | --- | --- |
| 1 | `group-set-mutes` 提交 `add` → `group-mutes` → `group-set-mutes` 提交 `update`、`del` | 测试账号发言受限，解除后恢复。查询可看到到期时间。 |
| 2 | 测试账号申请入群 → `group-join-requests` → `group-decide-join` | 列表的 `join_request_id` 与收到的申请一致。用两次申请分别验证 `approve` 和 `decline`。 |
| 3 | `group-remove-members` → `group-update-blacklist` 的 `add` → `group-blacklist` → 黑名单 `del` | 用专门账号核对移除和黑名单。检查批量结果中的失败项。 |
| 4 | `group-create-strategy` → `group-strategies` → `group-update-strategy` → `group-strategy-whitelist` → `group-execute-strategy` → `group-delete-strategy` | 只关联测试群。白名单加删均要验证。执行成功表示 QQ 已受理，最终审批需查看群内状态及事件。 |

```powershell
$g = $ids.group_openid
$m = $ids.member_openid

# 只在专用测试群和测试账号执行
& $run -Operation group-remove-members -Ids @($g) -Write -Body @{member_openids = @($m)}
& $run -Operation group-update-blacklist -Ids @($g) -Write -Body @{op = 'add'; member_openids = @($m)}
& $run -Operation group-blacklist -Ids @($g)
& $run -Operation group-update-blacklist -Ids @($g) -Write -Body @{op = 'del'; member_openids = @($m)}
```

禁言和审批的完整命令见 [基础验收命令](README.md#单次调用管理与消息-api)。审批策略使用 `group_openids`、`strategy_id` 和白名单测试账号的 QQ 号。参数可用 `go doc ./pkg/api/group CreateStrategyRequest` 等命令查看。

创建策略时先设为停用。用返回的策略 ID 验证修改和白名单。只有准备好一条允许自动通过的测试申请后，才启用并执行策略。

```powershell
$strategy = & $run -Operation group-create-strategy -Write -Body @{
    group_openids = @($g)
    is_enable = 'off'
    remark = 'GoQQBot release test'
} | ConvertFrom-Json
$strategyId = $strategy.strategy_id
if ($strategyId) {
    & $run -Operation group-update-strategy -Ids @($strategyId) -Write -Body @{remark = 'GoQQBot release test updated'}
    & $run -Operation group-strategies
}

# 将 $testQQ 设置为测试账号的 QQ 号（字符串）
& $run -Operation group-strategy-whitelist -Ids @($strategyId) -Write -Body @{op = 'add'; whitelist_users = @($testQQ)}
# 准备好对应申请后再运行以下两条
& $run -Operation group-update-strategy -Ids @($strategyId) -Write -Body @{is_enable = 'on'}
& $run -Operation group-execute-strategy -Ids @($strategyId) -Write
# 在 QQ 核对审批后清理
& $run -Operation group-strategy-whitelist -Ids @($strategyId) -Write -Body @{op = 'del'; whitelist_users = @($testQQ)}
& $run -Operation group-delete-strategy -Ids @($strategyId) -Write
```

## 频道管理和消息写入

| 测试顺序 | 操作 | QQ 或响应核对 |
| --- | --- | --- |
| 1 | `channel-create` → `channel-update` → `channel` → `channel-delete` | 使用新建的测试子频道。观察创建、修改和删除事件。 |
| 2 | `guild-create-role` → `guild-update-role` → `guild-add-member-role` → `guild-role-members` → `guild-remove-member-role` → `guild-delete-role` | 仅操作新身份组。确认成员列表和客户端显示。 |
| 3 | `channel-update-member-permissions`、`channel-update-role-permissions` | 写入前保存测试对象的权限。写入后查询验证，再恢复原值。 |
| 4 | `guild-api-permissions` → `guild-request-api-permission` | 使用列表中可申请的接口，核对授权消息与后续授权状态。 |
| 5 | `guild-member-mute`、`guild-members-mute`、`guild-mute` | 只在测试频道禁言。先测试短时禁言，再传 `mute_seconds = '0'` 解除。批量禁言核对成功用户列表。 |
| 6 | `channel-send` → `channel-add-pin` → `channel-pins` → `channel-delete-pin`，以及 `guild-create-announcement` → `guild-delete-announcement` | 使用刚发送的测试消息，记录其 ID 并核对客户端展示。 |
| 7 | `channel-create-schedule` → `channel-schedule` → `channel-update-schedule` → `channel-delete-schedule` | 在日程子频道核对名称、时间和清理结果。 |
| 8 | `channel-create-thread` → `channel-threads` → `channel-thread` → `channel-delete-thread` | 发帖返回任务 ID。列表或事件返回的帖子 ID 用于查询和删除。 |
| 9 | `channel-join-mic` → `channel-audio` → `channel-leave-mic` | 在音视频子频道验证播放、暂停、继续、停止和上下麦事件。 |
| 10 | `guild-remove-member` | 使用专门的频道测试成员，核对成员移除事件。 |
| 11 | `channel-send`、`dm-create`、`dm-send`、表态添加/查询/删除、`channel-recall`、`dm-recall` | 核对返回的消息 ID、客户端内容、表态、审核和撤回状态。频道私信图片已实测成功，当前机器人的频道私信撤回已确认权限受限。 |

另外在 QQ 里验证普通消息、`/sdk-ping`、`/sdk-image`、`/sdk-legacy` 和 `/sdk-recall`。论坛及音频事件需要对应的子频道和订阅权限，具体订阅见 [基础验收说明](README.md#配置事件订阅)。

## 发版前核对

```powershell
go test ./...
go test -race ./...
go vet ./...
python wiki/_tools/check_docs.py
npm run docs:build
```

将权限受限和缺少测试资源的条目单独记录。需要对外声称实机通过的功能，应有真实 QQ 通过记录。未解决的 SDK 失败应先修复，再确定版本和发布。
