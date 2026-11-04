// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"context"
	"net/http"
	"strconv"

	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// ScheduleReminder 表示平台的日程提醒时间。
type ScheduleReminder string

// 各日程提醒时间对应的协议值。
const (
	RemindNever          ScheduleReminder = "0"
	RemindAtStart        ScheduleReminder = "1"
	RemindFiveMinutes    ScheduleReminder = "2"
	RemindFifteenMinutes ScheduleReminder = "3"
	RemindThirtyMinutes  ScheduleReminder = "4"
	RemindSixtyMinutes   ScheduleReminder = "5"
)

// ScheduleInput 包含创建或修改日程时可提交的字段。
type ScheduleInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// StartTimestamp 和 EndTimestamp 使用 Unix 毫秒，协议中以字符串发送。
	StartTimestamp uint64           `json:"start_timestamp,string"`
	EndTimestamp   uint64           `json:"end_timestamp,string"`
	JumpChannelID  string           `json:"jump_channel_id,omitempty"`
	Reminder       ScheduleReminder `json:"remind_type,omitempty"`
}

// Schedule 表示已保存的日程，包括平台分配的 ID 与创建者。
type Schedule struct {
	ScheduleInput
	ID      string             `json:"id"`
	Creator *channelmsg.Member `json:"creator"`
}

type scheduleBody struct {
	Schedule ScheduleInput `json:"schedule"`
}

// ListSchedules 按结束时间查询日程。sinceMillis 为零时使用平台默认的当天范围。
func (c *Client) ListSchedules(ctx context.Context, channelID string, sinceMillis uint64) ([]Schedule, error) {
	path, err := apiwire.Path("/channels/%s/schedules", channelID)
	if err != nil {
		return nil, err
	}
	if sinceMillis != 0 {
		path += "?since=" + strconv.FormatUint(sinceMillis, 10)
	}
	items, err := apiwire.Do[[]Schedule](ctx, c.requester(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return *items, nil
}

// GetSchedule 查询指定日程的详情。
func (c *Client) GetSchedule(ctx context.Context, channelID, scheduleID string) (*Schedule, error) {
	path, err := apiwire.Path("/channels/%s/schedules/%s", channelID, scheduleID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Schedule](ctx, c.requester(), http.MethodGet, path, nil)
}

// CreateSchedule 在日程子频道创建日程。平台校验机器人管理员权限。
func (c *Client) CreateSchedule(ctx context.Context, channelID string, input ScheduleInput) (*Schedule, error) {
	path, err := apiwire.Path("/channels/%s/schedules", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Schedule](ctx, c.requester(), http.MethodPost, path, scheduleBody{Schedule: input})
}

// UpdateSchedule 提交指定日程的新设置。平台校验机器人管理员权限。
func (c *Client) UpdateSchedule(ctx context.Context, channelID, scheduleID string, input ScheduleInput) (*Schedule, error) {
	path, err := apiwire.Path("/channels/%s/schedules/%s", channelID, scheduleID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[Schedule](ctx, c.requester(), http.MethodPatch, path, scheduleBody{Schedule: input})
}

// DeleteSchedule 删除指定日程。平台校验机器人管理员权限。
func (c *Client) DeleteSchedule(ctx context.Context, channelID, scheduleID string) error {
	path, err := apiwire.Path("/channels/%s/schedules/%s", channelID, scheduleID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}
