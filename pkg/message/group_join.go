package message

// JoinRequest 描述待审批的 QQ 群入群申请。
type JoinRequest struct {
	JoinRequestID string `json:"join_request_id"`
	RiskTips      string `json:"risk_tips"`
	UnionOpenID   string `json:"union_openid"`
	MemberOpenID  string `json:"member_openid"`
	Username      string `json:"username"`
	// ApplyAt 保留平台返回的 RFC3339 申请时间。
	ApplyAt string `json:"apply_at"`
	// ApplySource 为 self_apply 或 invited 时分别表示自主申请和受邀入群。
	ApplySource string            `json:"apply_source"`
	InvitedBy   string            `json:"invited_by"`
	Bot         bool              `json:"bot"`
	VerifyInfo  *JoinVerification `json:"verify_info"`
}

// JoinVerification 保存入群验证消息或管理员设置的问答。
type JoinVerification struct {
	Method        string     `json:"method"`
	VerifyMessage string     `json:"verify_message"`
	ReviewQAList  []ReviewQA `json:"review_qa_list"`
}

// ReviewQA 是一条入群验证问题及申请人的回答。
type ReviewQA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// GroupJoinRequestEvent 是 GROUP_JOIN_REQUEST 的事件数据。
type GroupJoinRequestEvent struct {
	GroupOpenID string `json:"group_openid"`
	JoinRequest
	// AutoApproved 仅在申请已通过自动审批时存在。
	AutoApproved *AutoApproval `json:"auto_approved"`
}

// AutoApproval 标识批准该申请的自动审批策略。
type AutoApproval struct {
	StrategyID string `json:"strategy_id"`
}
