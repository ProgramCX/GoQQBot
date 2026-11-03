# 官方 API 覆盖核对

核对日期：2026-11-04。依据当前 [API v2 文档](https://bot.q.qq.com/wiki/develop/api-v2/)及其导航，读取服务端接口、事件等 160 个页面，对照当前工作区源码。本次只作覆盖审查，不新增 API，不将平台权限未开放误判为 SDK 缺失。

判定口径：有明确公开方法和请求/响应处理才算已封装；有 EventId、裸 JSON、WithField 或 Client.Request，不等于该领域已经具备类型化 SDK。已封装也不等于真实 QQ 联调已通过，更不保证每个客户端显示一致。表中的状态是实现覆盖，不是平台权限判断。

## 结论

单聊/群聊的消息主链路、富媒体上传、Markdown 图片素材、按钮和互动回调已具备主要封装。明显缺口是机器人资料与分享、全局菜单/指令面板、群管理，以及大部分频道管理与内容操作。

当前导航的 `autogen/api` 下有 47 个独立接口页面，其中 13 个已有对应封装，34 个缺少领域封装。这个统计**仅针对这 47 页**：不包含另外独立成文的访问凭证、素材管理、频道消息/表情表态等，因此不是整个 SDK 的覆盖率。

| 当前自动生成接口中的缺口 | 数量 |
| --- | ---: |
| 机器人详情、机器人频道列表、分享链接 | 3 |
| 全局菜单与指令面板 | 8 |
| 群资料、成员、禁言、黑名单、入群审批与策略 | 17 |
| 频道/子频道资料与子频道增删改查 | 6 |
| 合计 | 34 |

## 已有主要能力及边界

| 领域 | 当前入口 | 覆盖边界 |
| --- | --- | --- |
| 访问凭证与网关 | Client.AccessToken、Gateway，WebSocket/Webhook | 已有认证、接入、心跳和重连；会话进度在内存中 |
| 单聊、群聊收发 | OnMessage(type, handler)、Context.Message/Reply、Client.SendReply/SendMessage | 已有普通消息、Markdown、图文卡片、引用、主动发送、单聊输入状态等；平台资格仍由服务器检查 |
| 撤回 | Client.RecallMessage、Context.Recall | 单聊/群聊；频道/频道私信另有旧入口 |
| 富媒体 | UploadMedia、PrepareUpload、UploadPart、FinishUploadPart、UploadFile | URL/分片流程已封装；file_data 仅原样透传，官方未明确编码，未声称完成其真实验收 |
| 流式消息 | Client/Context.SendStreamChunk | 已有单聊协议；排序、游标持久化和恢复由业务处理，不自动重试 |
| Markdown 图片素材 | Prepare/Put/Complete/List/DeleteMarkdownImage、UploadMarkdownImage(File) | 主流程与资源管理已封装；不是普通富媒体 file_info；手机显示仍需实机验证 |
| 按钮/文本互动 | button、text、Outgoing | 自定义/模板键盘、动作、权限、提示、确认、文本交互等已封装 |
| 互动回调 | OnButton、OnMenu、OnInteraction、Interaction、AckInteraction | 已覆盖文档列出的互动类型；不等同于具备菜单或面板管理 API |
| 频道消息与表态 | api/channel 的 SendMessage/SendDM/SendArk/SendEmbed/Recall、AddReaction/RemoveReaction/GetReactionUsers | 已有旧封装，仍有返回模型、日志、转义和统一调用体验需要修正 |

## 当前明确的功能缺口

1. **机器人自身接口**：获取机器人详情、所在频道列表、生成带 callback_data 的邀请/分享链接。不能把能取 AccessToken 当作已覆盖这些机器人业务接口。
2. **全局菜单与指令面板**：菜单查询与更新；面板创建、列表、详情、修改、删除、关联对象更新。现在 OnMenu 只是消费回调，仍需外部方式管理菜单。
3. **群管理**：群基本资料、机器人群内状态、成员列表/详情、成员禁言、黑名单、批量移除、入群申请查询/审批；入群自动审批策略的创建/列表/修改/删除/执行/白名单更新。
4. **事件类型化与订阅**：群成员增减、入群申请、频道/子频道生命周期、好友/主动推送开关、订阅状态等，缺少统一 DTO 与 Context 解码入口。部分已有常量或旧结构，不能说完全无法接收；OnEvent 可读取 Payload.D，但业务仍需要手工解析。
5. **频道管理与内容能力**：成员、身份组、子频道权限、禁言、在线成员数、API 权限与授权链接、公告、精华消息、日程、帖子、音频控制和上下麦等尚无对应领域封装。
6. **频道私信会话创建**：已有发送/撤回频道私信，但未封装 `POST /users/@me/dms` 创建私信会话。

## 应先修的已有能力缺口

这些不应与“新加一个 API”混为一谈。

- **缺少 GROUP_MEMBER_EVENT Intent**：官方 GROUP_JOIN_REQUEST、GROUP_MEMBER_ADD、GROUP_MEMBER_REMOVE 使用 `GROUP_MEMBER_EVENT (1<<24)`，但 `pkg/config/config.go` 的 IntentMask 没有这个名称，会在配置校验时拒绝它。Webhook 可在平台侧另行订阅，不能用此结论否认所有传输模式下的原始事件接收能力。[入群申请文档](https://bot.q.qq.com/wiki/develop/api-v2/autogen/event/group_join_request.html)
- **生命周期事件的便捷回复未覆盖**：官方支持对 FRIEND_ADD、C2C_MSG_RECEIVE、GROUP_ADD_ROBOT、GROUP_MSG_RECEIVE 等部分事件按 envelope event_id 回复；当前 Context.replyTarget 仅识别普通消息和 INTERACTION_CREATE。低层 Client.SendReply 可以手动传 ReplyTarget，不代表 Context.Reply 已覆盖这些事件。
- **旧频道发送响应时间戳不匹配**：旧 api/channel 发送函数解码为 `message.SendMessageResponse`，其 Timestamp 是 int64；官方频道发送示例 timestamp 为 RFC3339 字符串。存在 HTTP 已发送成功却因响应解码报错的风险，需与新 ReplyResult 的兼容策略统一。[频道发送文档](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html)
- **旧表态分页参数没有 URL 编码**：GetReactionUsers 直接拼接 cookie，特殊字符可能改变查询语义；ID 路径拼接和 nil/context 校验也未统一到新 API 入口。
- **日志与异常处理已修正（2026-11-04）**：已删除旧频道消息体日志和日志包自动创建文件行为；Bot 支持独立 slog 注入及错误回调，handler panic 不再静默丢失。用法和停止返回值的迁移说明见 [日志与异常处理](diagnostics.md)。
- **频道接收视图未完整接入 Context**：统一 Received 不直接暴露频道 Member、Embeds、Seq 等完整视图；部分字段已有旧 channel.Message 结构，可手工解码，其他字段保留在 Raw，仍需完善类型化入口。
- **已知消息类型的场景校验仍不完整**：Embed 已有旧发送入口，但统一 builder 没有完整的 Embed/ARK 校验分支；不能仅因 New/WithField 能拼出 JSON 就声称具备完整高层封装。

## 不计作官方 API 缺失的事项

- URL 大文件拉取失败是联调中的平台拉取/转存问题，不能通过虚构错误码或调大本机 HTTP 超时解决。SDK 已提供显式本地分片路径。
- 当前手机端 qqbot-media 图片显示差异尚未实机复测，已补自动尺寸对照；这是兼容性验证，不应重新计作“素材管理 API 未实现”。
- 数据库持久化、缓存、跨重启去重、自动删除素材、多实例流式状态等属于业务/运行时策略，不是官方新增 HTTP 接口。当前 SDK 不隐式实现这些行为。
- 频道小程序文档中的 qq.getGuildInfo 等客户端 JS 能力不属于 Go 服务端 SDK；若业务需要对应开放数据解密，可另行评估，不混入服务端 HTTP 缺口统计。

## 建议补齐顺序

1. 修正 Intent、生命周期事件回复、旧响应模型和编码/日志问题，补上可复现的回归测试。
2. 机器人基础资料、分享链接、全局菜单与指令面板，配合已有按钮/回调能力组成完整交互链路。
3. 群信息与成员查询、群成员/审批事件 DTO，再按权限补禁言、黑名单、移除和审批策略。
4. 根据实际频道业务需要补频道管理、权限、公告、日程、论坛、音频等模块。

以下明细保留每项官方链接，便于后续实现时逐项验收。

## 自动生成接口逐项核对

| 官方接口 | HTTP | SDK 状态 |
| --- | --- | --- |
| [删除子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.delete.html) | `DELETE /channels/{channel_id}` | 缺少领域封装 |
| [获取子频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.get.html) | `GET /channels/{channel_id}` | 缺少领域封装 |
| [修改子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/channels_channel_id.patch.html) | `PATCH /channels/{channel_id}` | 缺少领域封装 |
| [获取通用 WSS 接入点](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/gateway.get.html) | `GET /gateway` | 已封装 |
| [获取频道详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id.get.html) | `GET /guilds/{guild_id}` | 缺少领域封装 |
| [获取子频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.get.html) | `GET /guilds/{guild_id}/channels` | 缺少领域封装 |
| [创建子频道](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/guilds_guild_id_channels.post.html) | `POST /guilds/{guild_id}/channels` | 缺少领域封装 |
| [互动事件响应](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/interactions_interaction_id.put.html) | `PUT /interactions/{interaction_id}` | 已封装 |
| [获取机器人详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me.get.html) | `GET /users/@me` | 缺少领域封装 |
| [获取机器人频道列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/users_me_guilds.get.html) | `GET /users/@me/guilds` | 缺少领域封装 |
| [生成分享链接](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_generate_url_link.post.html) | `POST /v2/generate_url_link` | 缺少领域封装 |
| [群聊分片上传完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_part_finish.post.html) | `POST /v2/groups/{group_id}/upload_part_finish` | 已封装 |
| [群聊富媒体预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_id_upload_prepare.post.html) | `POST /v2/groups/{group_id}/upload_prepare` | 已封装 |
| [入群申请审批](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_approval_join_request_member_openid.post.html) | `POST /v2/groups/{group_openid}/approval_join_request/{member_openid}` | 缺少领域封装 |
| [群成员批量移除](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_batch_remove_members.post.html) | `POST /v2/groups/{group_openid}/batch_remove_members` | 缺少领域封装 |
| [获取机器人群内状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_bot_state.get.html) | `GET /v2/groups/{group_openid}/bot_state` | 缺少领域封装 |
| [群聊富媒体上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_files.post.html) | `POST /v2/groups/{group_openid}/files` | 已封装 |
| [获取群基本信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_info.get.html) | `GET /v2/groups/{group_openid}/info` | 缺少领域封装 |
| [入群申请列表拉取](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_join_request_list.get.html) | `GET /v2/groups/{group_openid}/join_request_list` | 缺少领域封装 |
| [群黑名单查询](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.get.html) | `GET /v2/groups/{group_openid}/member_blacklist` | 缺少领域封装 |
| [群黑名单操作](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_member_blacklist.post.html) | `POST /v2/groups/{group_openid}/member_blacklist` | 缺少领域封装 |
| [获取群成员列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members.get.html) | `GET /v2/groups/{group_openid}/members` | 缺少领域封装 |
| [获取群成员信息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_members_member_openid.get.html) | `GET /v2/groups/{group_openid}/members/{member_openid}` | 缺少领域封装 |
| [发送群聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html) | `POST /v2/groups/{group_openid}/messages` | 已封装 |
| [撤回群聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages_message_id.delete.html) | `DELETE /v2/groups/{group_openid}/messages/{message_id}` | 已封装 |
| [查询群禁言状态](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.get.html) | `GET /v2/groups/{group_openid}/restrict_chat_setting` | 缺少领域封装 |
| [设置群成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_restrict_chat_setting.post.html) | `POST /v2/groups/{group_openid}/restrict_chat_setting` | 缺少领域封装 |
| [查询入群自动审批策略列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.get.html) | `GET /v2/groups/join_approval_strategy` | 缺少领域封装 |
| [创建入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy.post.html) | `POST /v2/groups/join_approval_strategy` | 缺少领域封装 |
| [删除入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.delete.html) | `DELETE /v2/groups/join_approval_strategy/{strategy_id}` | 缺少领域封装 |
| [修改入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id.patch.html) | `PATCH /v2/groups/join_approval_strategy/{strategy_id}` | 缺少领域封装 |
| [执行入群自动审批策略](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_execute.post.html) | `POST /v2/groups/join_approval_strategy/{strategy_id}/execute` | 缺少领域封装 |
| [修改入群自动审批策略的白名单号码](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_join_approval_strategy_strategy_id_whitelist_users.post.html) | `POST /v2/groups/join_approval_strategy/{strategy_id}/whitelist_users` | 缺少领域封装 |
| [查询全局自定义菜单](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_menu.get.html) | `GET /v2/menu` | 缺少领域封装 |
| [修改全局自定义菜单](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_menu.put.html) | `PUT /v2/menu` | 缺少领域封装 |
| [查询指令面板列表](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels.get.html) | `GET /v2/panels` | 缺少领域封装 |
| [创建指令面板](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels.post.html) | `POST /v2/panels` | 缺少领域封装 |
| [删除指令面板](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id.delete.html) | `DELETE /v2/panels/{panel_id}` | 缺少领域封装 |
| [查询指令面板详情](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id.get.html) | `GET /v2/panels/{panel_id}` | 缺少领域封装 |
| [修改指令面板](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id.put.html) | `PUT /v2/panels/{panel_id}` | 缺少领域封装 |
| [修改指令面板关联对象](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_panels_panel_id_target.put.html) | `PUT /v2/panels/{panel_id}/target` | 缺少领域封装 |
| [单聊分片上传完成](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_part_finish.post.html) | `POST /v2/users/{user_id}/upload_part_finish` | 已封装 |
| [单聊富媒体预上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_id_upload_prepare.post.html) | `POST /v2/users/{user_id}/upload_prepare` | 已封装 |
| [单聊富媒体上传](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_files.post.html) | `POST /v2/users/{user_openid}/files` | 已封装 |
| [发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages.post.html) | `POST /v2/users/{user_openid}/messages` | 已封装 |
| [撤回单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_messages_message_id.delete.html) | `DELETE /v2/users/{user_openid}/messages/{message_id}` | 已封装 |
| [流式发送单聊消息](https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_users_user_openid_stream_messages.post.html) | `POST /v2/users/{user_openid}/stream_messages` | 已封装 |

## 频道旧版分栏接口核对

此表按文档页面列出，不与上面 47 页混算；同一个路径可能承担单人/批量等不同请求模式。

| 官方接口 | HTTP（页面中的路径模板） | SDK 状态 |
| --- | --- | --- |
| [获取机器人在频道可用权限列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/api_permissions/get_guild_api_permission.html) | `GET /guilds/{guild_id}/api_permission` | 缺少领域封装 |
| [发送机器人在频道接口权限的授权链接](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/api_permissions/post_api_permission_demand.html) | `POST /guilds/{guild_id}/api_permission/demand` | 缺少领域封装 |
| [删除子频道公告（2022年3月15日后废弃）](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/delete_channel_announces.html) | `DELETE /channels/{channel_id}/announces/{message_id}` | 缺少领域封装 |
| [删除频道公告](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/delete_guild_announces.html) | `DELETE /guilds/{guild_id}/announces/{message_id}` | 缺少领域封装 |
| [创建子频道公告（2022年3月15日后废弃）](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/post_channel_announces.html) | `POST /channels/{channel_id}/announces` | 缺少领域封装 |
| [创建频道公告](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/announces/post_guild_announces.html) | `POST /guilds/{guild_id}/announces` | 缺少领域封装 |
| [音频控制](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/audio_control.html) | `POST /channels/{channel_id}/audio` | 缺少领域封装 |
| [机器人下麦](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/delete_mic.html) | `DELETE /channels/{channel_id}/mic` | 缺少领域封装 |
| [机器人上麦](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/audio/put_mic.html) | `PUT /channels/{channel_id}/mic` | 缺少领域封装 |
| [删除帖子](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/delete_thread.html) | `DELETE /channels/{channel_id}/threads/{thread_id}` | 缺少领域封装 |
| [获取帖子详情](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/get_thread.html) | `GET /channels/{channel_id}/threads/{thread_id}` | 缺少领域封装 |
| [获取帖子列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/get_threads_list.html) | `GET /channels/{channel_id}/threads` | 缺少领域封装 |
| [发表帖子](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/forum/put_thread.html) | `PUT /channels/{channel_id}/threads` | 缺少领域封装 |
| [删除精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/delete_pins_message.html) | `DELETE /channels/{channel_id}/pins/{message_id}` | 缺少领域封装 |
| [获取精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/get_pins_message.html) | `GET /channels/{channel_id}/pins` | 缺少领域封装 |
| [添加精华消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/pins/put_pins_message.html) | `PUT /channels/{channel_id}/pins/{message_id}` | 缺少领域封装 |
| [删除日程](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/delete_schedule.html) | `DELETE /channels/{channel_id}/schedules/{schedule_id}` | 缺少领域封装 |
| [获取日程详情](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/get_schedule.html) | `GET /channels/{channel_id}/schedules/{schedule_id}` | 缺少领域封装 |
| [获取频道日程列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/get_schedules.html) | `GET /channels/{channel_id}/schedules` | 缺少领域封装 |
| [修改日程](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/patch_schedule.html) | `PATCH /channels/{channel_id}/schedules/{schedule_id}` | 缺少领域封装 |
| [创建日程](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/content/schedule/post_schedule.html) | `POST /channels/{channel_id}/schedules` | 缺少领域封装 |
| [频道私信](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/dms.html) | `DELETE /dms/{guild_id}/messages/{message_id}?hidetip=false`<br>`POST /dms/{guild_id}/messages`<br>`POST /users/@me/dms` | 发送/撤回已有；创建会话未封装 |
| [撤回子频道消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/recall.html) | `DELETE /channels/{channel_id}/messages/{message_id}?hidetip=false` | 已有旧封装，需完善统一入口和返回模型 |
| [发送子频道消息](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/message/send.html) | `POST /channels/{channel_id}/messages` | 已有旧封装，需完善统一入口和返回模型 |
| [获取子频道用户权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/get_channel_permissions.html) | `GET /channels/{channel_id}/members/{user_id}/permissions` | 缺少领域封装 |
| [获取子频道身份组权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/get_channel_roles_permissions.html) | `GET /channels/{channel_id}/roles/{role_id}/permissions` | 缺少领域封装 |
| [修改子频道用户权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/put_channel_permissions.html) | `PUT /channels/{channel_id}/members/{user_id}/permissions` | 缺少领域封装 |
| [修改子频道身份组权限](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/channel_permissions/put_channel_roles_permissions.html) | `PUT /channels/{channel_id}/roles/{role_id}/permissions` | 缺少领域封装 |
| [删除频道身份组成员](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/delete_guild_member_role.html) | `DELETE /guilds/{guild_id}/members/{user_id}/roles/{role_id}` | 缺少领域封装 |
| [删除频道身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/delete_guild_role.html) | `DELETE /guilds/{guild_id}/roles/{role_id}` | 缺少领域封装 |
| [获取频道身份组列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/get_guild_roles.html) | `GET /guilds/{guild_id}/roles` | 缺少领域封装 |
| [修改频道身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/patch_guild_role.html) | `PATCH /guilds/{guild_id}/roles/{role_id}` | 缺少领域封装 |
| [创建频道身份组](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/post_guild_role.html) | `POST /guilds/{guild_id}/roles` | 缺少领域封装 |
| [创建频道身份组成员](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role-group/put_guild_member_role.html) | `PUT /guilds/{guild_id}/members/{user_id}/roles/{role_id}` | 缺少领域封装 |
| [获取子频道在线成员数](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/get_online_nums.html) | `GET /channels/{channel_id}/online_nums` | 缺少领域封装 |
| [删除频道成员](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/delete_member.html) | `DELETE /guilds/{guild_id}/members/{user_id}` | 缺少领域封装 |
| [获取频道成员详情](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_member.html) | `GET /guilds/{guild_id}/members/{user_id}` | 缺少领域封装 |
| [获取频道成员列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_members.html) | `GET /guilds/{guild_id}/members` | 缺少领域封装 |
| [获取频道身份组成员列表](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/role/member/get_role_members.html) | `GET /guilds/{guild_id}/roles/{role_id}/members` | 缺少领域封装 |
| [频道指定成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_member_mute.html) | `PATCH /guilds/{guild_id}/members/{user_id}/mute` | 缺少领域封装 |
| [频道全员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_mute.html) | `PATCH /guilds/{guild_id}/mute` | 缺少领域封装 |
| [频道批量成员禁言](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/patch_guild_mute_multi_member.html) | `PATCH /guilds/{guild_id}/mute` | 缺少领域封装 |
| [获取频道消息频率的设置详情](https://bot.q.qq.com/wiki/develop/api-v2/server-inter/channel/speak/setting/message_setting.html) | `GET /guilds/{guild_id}/message/setting` | 缺少领域封装 |
