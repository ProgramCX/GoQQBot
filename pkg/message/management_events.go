package message

// 群成员管理事件使用 GROUP_MEMBER_EVENT 订阅。
const (
	GROUP_JOIN_REQUEST  EventId = "GROUP_JOIN_REQUEST"
	GROUP_MEMBER_ADD    EventId = "GROUP_MEMBER_ADD"
	GROUP_MEMBER_REMOVE EventId = "GROUP_MEMBER_REMOVE"
)

// 频道审核和表态事件分别使用 MESSAGE_AUDIT 和 GUILD_MESSAGE_REACTIONS 订阅。
const (
	MESSAGE_AUDIT_PASS      EventId = "MESSAGE_AUDIT_PASS"
	MESSAGE_AUDIT_REJECT    EventId = "MESSAGE_AUDIT_REJECT"
	MESSAGE_REACTION_ADD    EventId = "MESSAGE_REACTION_ADD"
	MESSAGE_REACTION_REMOVE EventId = "MESSAGE_REACTION_REMOVE"
)

// GroupMemberEvent 表示加入或退出群聊的成员。
type GroupMemberEvent struct {
	GroupOpenID  string `json:"group_openid"`
	MemberOpenID string `json:"member_openid"`
	// UserOpenID 仅在部分退出事件中提供。
	UserOpenID string `json:"user_openid"`
	// Timestamp 是 Unix 秒。
	Timestamp int64 `json:"timestamp"`
}
