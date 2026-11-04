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

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// ThreadFormat 表示发帖内容的编码格式。
type ThreadFormat uint32

// 发帖支持的格式及其协议值。
const (
	ThreadText     ThreadFormat = 1
	ThreadHTML     ThreadFormat = 2
	ThreadMarkdown ThreadFormat = 3
	ThreadJSON     ThreadFormat = 4
)

// ThreadInput 是发帖参数。使用 ThreadJSON 时，Content 是编码后的富文本 JSON 字符串。
type ThreadInput struct {
	Title   string       `json:"title"`
	Content string       `json:"content"`
	Format  ThreadFormat `json:"format"`
}

// ThreadTask 是平台接收发帖任务后返回的结果。
type ThreadTask struct {
	TaskID string `json:"task_id"`
	// CreateTime 是平台返回的 Unix 秒时间戳字符串。
	CreateTime string `json:"create_time"`
}

// Thread 表示一个论坛主题及其所属子频道和作者。
type Thread struct {
	GuildID   string     `json:"guild_id"`
	ChannelID string     `json:"channel_id"`
	AuthorID  string     `json:"author_id"`
	Info      ThreadInfo `json:"thread_info"`
}

// ThreadInfo 保存主题的正文和发表时间。
type ThreadInfo struct {
	ID    string `json:"thread_id"`
	Title string `json:"title"`
	// Content 保留平台返回的富文本 JSON 字符串。
	Content string `json:"content"`
	// DateTime 是平台返回的 ISO 8601 时间。
	DateTime string `json:"date_time"`
}

// ThreadPage 是平台返回的一批帖子。
type ThreadPage struct {
	Threads []Thread `json:"threads"`
	// IsFinish 为 1 时表示拉取完毕，为 0 时表示未结束。
	IsFinish uint32 `json:"is_finish"`
}

// ListThreads 查询帖子列表。平台目前将该接口开放给私域机器人。
func (c *Client) ListThreads(ctx context.Context, channelID string) (*ThreadPage, error) {
	path, err := apiwire.Path("/channels/%s/threads", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[ThreadPage](ctx, c.requester(), http.MethodGet, path, nil)
}

// GetThread 查询论坛主题及其富文本正文。
func (c *Client) GetThread(ctx context.Context, channelID, threadID string) (*Thread, error) {
	path, err := apiwire.Path("/channels/%s/threads/%s", channelID, threadID)
	if err != nil {
		return nil, err
	}
	result, err := apiwire.Do[struct {
		Thread Thread `json:"thread"`
	}](ctx, c.requester(), http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return &result.Thread, nil
}

// CreateThread 提交发帖任务。返回的是任务 ID，供业务记录提交结果。
func (c *Client) CreateThread(ctx context.Context, channelID string, input ThreadInput) (*ThreadTask, error) {
	path, err := apiwire.Path("/channels/%s/threads", channelID)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[ThreadTask](ctx, c.requester(), http.MethodPut, path, input)
}

// DeleteThread 删除指定子频道中的主题。
func (c *Client) DeleteThread(ctx context.Context, channelID, threadID string) error {
	path, err := apiwire.Path("/channels/%s/threads/%s", channelID, threadID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}
