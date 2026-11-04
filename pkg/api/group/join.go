package group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/message"
)

// JoinRequest 是入群申请列表和入群事件共用的申请数据。
type JoinRequest = message.JoinRequest

// JoinRequestPage 是一页入群申请。NextCursor 为空时已到末页。
type JoinRequestPage struct {
	Requests   []JoinRequest `json:"list"`
	NextCursor string        `json:"next_cursor"`
}

// ListJoinRequests 查询一页待审批申请。Limit 为 0 时由平台默认取 20 条，最多 50 条。
func (c *Client) ListJoinRequests(ctx context.Context, groupOpenID string, page PageOptions) (*JoinRequestPage, error) {
	path, err := apiwire.Path(groupPath+"/join_request_list", groupOpenID)
	if err != nil {
		return nil, err
	}
	return getPage[JoinRequestPage](ctx, c, path, page, 50)
}

// JoinDecision 表示同意或拒绝入群申请。
type JoinDecision string

const (
	// JoinApprove 同意入群申请。
	JoinApprove JoinDecision = "approve"
	// JoinDecline 拒绝入群申请。
	JoinDecline JoinDecision = "decline"
)

// DecideJoinRequest 指定入群申请及审批决定。
type DecideJoinRequest struct {
	Decision JoinDecision `json:"op"`
	// JoinRequestID 使用 QQ 在列表或事件中下发的申请 ID。
	JoinRequestID string `json:"join_request_id,omitempty"`
	// RejectReason 和 AddToMemberBlacklist 仅适用于拒绝决定。
	RejectReason         string `json:"reject_reason,omitempty"`
	AddToMemberBlacklist bool   `json:"add_to_member_blacklist,omitempty"`
}

// DecideJoin 提交一项明确的审批决定。QQ 会检查机器人的群管理员身份。
func (c *Client) DecideJoin(ctx context.Context, groupOpenID, memberOpenID string, request DecideJoinRequest) error {
	path, err := apiwire.Path(groupPath+"/approval_join_request/%s", groupOpenID, memberOpenID)
	if err != nil {
		return err
	}
	switch request.Decision {
	case JoinApprove:
		if request.RejectReason != "" || request.AddToMemberBlacklist {
			return fmt.Errorf("reject_reason and add_to_member_blacklist require a decline decision")
		}
	case JoinDecline:
	default:
		return fmt.Errorf("join decision must be approve or decline")
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPost, path, request)
}
