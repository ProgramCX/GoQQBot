package group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// Info 是 QQ 返回的群资料。
type Info struct {
	GroupOpenID string   `json:"group_openid"`
	Name        string   `json:"group_name"`
	Description string   `json:"group_finger_memo"`
	Category    string   `json:"group_class_text"`
	Tags        []string `json:"group_tags"`
	MemberCount int      `json:"group_member_num"`
}

// BotState 是当前机器人在指定群中的身份和消息设置。
type BotState struct {
	MemberOpenID      string `json:"member_openid"`
	JoinedAt          string `json:"joined_at"`
	AllowProactiveMsg bool   `json:"allow_proactive_msg"`
	// ReceiveMessageSetting 可为 all、only_mention 或 mention_and_context。
	ReceiveMessageSetting string `json:"recv_msg_setting"`
	Role                  string `json:"member_role"`
}

// GetInfo 查询指定群的资料。QQ 会检查应用的白名单权限。
func (c *Client) GetInfo(ctx context.Context, groupOpenID string) (*Info, error) {
	path, err := apiwire.Path(groupPath+"/info", groupOpenID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Info](ctx, c.requester(), http.MethodGet, path, nil)
}

// GetBotState 查询当前机器人在群中的状态。QQ 会检查应用的白名单权限。
func (c *Client) GetBotState(ctx context.Context, groupOpenID string) (*BotState, error) {
	path, err := apiwire.Path(groupPath+"/bot_state", groupOpenID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[BotState](ctx, c.requester(), http.MethodGet, path, nil)
}

// Member 是 QQ 群成员信息。JoinedAt 保留平台返回的 RFC3339 时间。
type Member struct {
	MemberOpenID string `json:"member_openid"`
	Username     string `json:"username"`
	Role         string `json:"member_role"`
	Bot          bool   `json:"bot"`
	JoinedAt     string `json:"joined_at"`
	UnionOpenID  string `json:"union_openid"`
}

// MemberPage 是一页群成员。NextCursor 为空时已到末页。
type MemberPage struct {
	Members    []Member `json:"members"`
	NextCursor string   `json:"next_cursor"`
}

// ListMembers 查询一页群成员，每页最多 30 条。空 cursor 查询首页。
func (c *Client) ListMembers(ctx context.Context, groupOpenID, cursor string) (*MemberPage, error) {
	path, err := apiwire.Path(groupPath+"/members", groupOpenID)
	if err != nil {
		return nil, err
	}
	return getPage[MemberPage](ctx, c, path, PageOptions{Cursor: cursor}, 0)
}

// GetMember 查询指定群中的成员。memberOpenID 是群内标识。
func (c *Client) GetMember(ctx context.Context, groupOpenID, memberOpenID string) (*Member, error) {
	path, err := apiwire.Path(groupPath+"/members/%s", groupOpenID, memberOpenID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Member](ctx, c.requester(), http.MethodGet, path, nil)
}

// RemoveMembersRequest 指定 1 到 20 名待移除的群成员。
type RemoveMembersRequest struct {
	MemberOpenIDs        []string `json:"member_openids"`
	AddToMemberBlacklist bool     `json:"add_to_member_blacklist,omitempty"`
}

// RemoveMembersResult 报告移除结果和未能加入黑名单的成员。
type RemoveMembersResult struct {
	// Result 为 success 时表示移除成功。
	Result           string   `json:"remove_members_result"`
	BlacklistFailIDs []string `json:"add_to_member_blacklist_fail_openids"`
}

// RemoveMembers 批量移除成员。请求成功后仍应检查拉黑失败列表。
func (c *Client) RemoveMembers(ctx context.Context, groupOpenID string, request RemoveMembersRequest) (*RemoveMembersResult, error) {
	path, err := apiwire.Path(groupPath+"/batch_remove_members", groupOpenID)
	if err != nil {
		return nil, err
	}
	if err := validateBatch(request.MemberOpenIDs, 20, "member_openids"); err != nil {
		return nil, err
	}
	return apiwire.Do[RemoveMembersResult](ctx, c.requester(), http.MethodPost, path, request)
}

// BlacklistUser 是群黑名单中的一条用户记录。
type BlacklistUser struct {
	UnionOpenID  string `json:"union_openid"`
	MemberOpenID string `json:"member_openid"`
	Username     string `json:"username"`
	BannedAt     string `json:"banned_at"`
	Bot          bool   `json:"bot"`
}

// BlacklistPage 是一页黑名单记录。NextCursor 为空时已到末页。
type BlacklistPage struct {
	Users      []BlacklistUser `json:"users"`
	NextCursor string          `json:"next_cursor"`
}

// ListBlacklist 查询一页群黑名单。Limit 为 0 时由平台默认取 20 条，最多 100 条。
func (c *Client) ListBlacklist(ctx context.Context, groupOpenID string, page PageOptions) (*BlacklistPage, error) {
	path, err := apiwire.Path(groupPath+"/member_blacklist", groupOpenID)
	if err != nil {
		return nil, err
	}
	return getPage[BlacklistPage](ctx, c, path, page, 100)
}

// UpdateBlacklistRequest 指定黑名单操作及 1 到 20 个成员 OpenID。
type UpdateBlacklistRequest struct {
	Action        ListAction `json:"op"`
	MemberOpenIDs []string   `json:"member_openids"`
}

// UpdateBlacklistResult 列出 QQ 未能修改黑名单状态的成员。
type UpdateBlacklistResult struct {
	FailedOpenIDs []string `json:"fail_openids"`
}

// UpdateBlacklist 增删群黑名单条目。仍在群中的成员无法加入黑名单。
func (c *Client) UpdateBlacklist(ctx context.Context, groupOpenID string, request UpdateBlacklistRequest) (*UpdateBlacklistResult, error) {
	path, err := apiwire.Path(groupPath+"/member_blacklist", groupOpenID)
	if err != nil {
		return nil, err
	}
	if request.Action != ListAdd && request.Action != ListDelete {
		return nil, fmt.Errorf("blacklist action must be add or del")
	}
	if err := validateBatch(request.MemberOpenIDs, 20, "member_openids"); err != nil {
		return nil, err
	}
	return apiwire.Do[UpdateBlacklistResult](ctx, c.requester(), http.MethodPost, path, request)
}
