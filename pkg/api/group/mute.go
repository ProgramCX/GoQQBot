package group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// MuteSettings 是群级禁言配置和当前仍在禁言中的成员。
type MuteSettings struct {
	GlobalRule GlobalMuteRule    `json:"global_rule"`
	Members    []MemberMuteState `json:"members"`
}

// GlobalMuteRule 描述全员禁言模式及定时规则。
type GlobalMuteRule struct {
	// Mode 可为 none、always 或 schedule。
	Mode           string          `json:"mode"`
	ScheduleRules  []MuteSchedule  `json:"schedule_rules"`
	RecurringRules []MuteRecurring `json:"recurring_rules"`
}

// MuteSchedule 是一次性禁言任务。开始和结束时间使用 RFC3339 格式。
type MuteSchedule struct {
	TaskID  string `json:"task_id"`
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
	Enabled bool   `json:"enabled"`
}

// MuteRecurring 是周期禁言任务。时间按北京时间 HH:mm 表示。
type MuteRecurring struct {
	TaskID string `json:"task_id"`
	// Weekdays 中 1 表示星期一，7 表示星期日。
	Weekdays  []int  `json:"weekdays"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Enabled   bool   `json:"enabled"`
}

// MemberMuteState 是一条尚未过期的成员禁言记录。
type MemberMuteState struct {
	MemberOpenID string `json:"member_openid"`
	ExpiresAt    string `json:"mute_expire_at"`
	Username     string `json:"username"`
	UnionOpenID  string `json:"union_openid"`
}

// GetMuteSettings 查询群级禁言配置和未过期的成员禁言。
func (c *Client) GetMuteSettings(ctx context.Context, groupOpenID string) (*MuteSettings, error) {
	path, err := apiwire.Path(groupPath+"/restrict_chat_setting", groupOpenID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[MuteSettings](ctx, c.requester(), http.MethodGet, path, nil)
}

// MuteAction 表示新增、更新或解除一名成员的禁言。
type MuteAction string

const (
	// MuteAdd 为成员设置禁言。
	MuteAdd MuteAction = "add"
	// MuteUpdate 更新成员禁言的到期时间。
	MuteUpdate MuteAction = "update"
	// MuteDelete 解除成员禁言。
	MuteDelete MuteAction = "del"
)

// MemberMuteChange 指定一名群成员的禁言操作。
type MemberMuteChange struct {
	Action       MuteAction `json:"op"`
	MemberOpenID string     `json:"member_openid"`
	// ExpiresAt 使用 RFC3339 格式。解除禁言时可省略。
	ExpiresAt string `json:"mute_expire_at,omitempty"`
}

// SetMemberMutes 批量设置成员禁言，每次最多 20 条。新增和更新只适用于普通成员。
func (c *Client) SetMemberMutes(ctx context.Context, groupOpenID string, changes []MemberMuteChange) error {
	path, err := apiwire.Path(groupPath+"/restrict_chat_setting", groupOpenID)
	if err != nil {
		return err
	}
	if len(changes) == 0 || len(changes) > 20 {
		return fmt.Errorf("mute changes must contain between 1 and 20 entries")
	}
	for _, change := range changes {
		if change.MemberOpenID == "" {
			return fmt.Errorf("mute change requires member_openid")
		}
		switch change.Action {
		case MuteAdd, MuteUpdate, MuteDelete:
		default:
			return fmt.Errorf("mute action must be add, update, or del")
		}
	}
	body := struct {
		Members []MemberMuteChange `json:"members"`
	}{Members: changes}
	return apiwire.Exec(ctx, c.requester(), http.MethodPost, path, body)
}
