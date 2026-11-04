// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

// ReactionPage 指定表态用户分页参数。零值读取第一页。
type ReactionPage struct {
	Cookie string
	Limit  int
}

// AddReaction 为指定频道消息添加机器人表态。
func (c *Client) AddReaction(ctx context.Context, channelID, messageID string, emojiType interaction.EmojiType, emojiID string) error {
	path, err := reactionPath(channelID, messageID, emojiType, emojiID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPut, path, nil)
}

// RemoveReaction 删除机器人在指定消息上的表态。
func (c *Client) RemoveReaction(ctx context.Context, channelID, messageID string, emojiType interaction.EmojiType, emojiID string) error {
	path, err := reactionPath(channelID, messageID, emojiType, emojiID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}

// ListReactionUsers 查询一页表态用户。QQ 返回的 Cookie 可原样传给下一次调用。
func (c *Client) ListReactionUsers(ctx context.Context, channelID, messageID string, emojiType interaction.EmojiType, emojiID string, page ReactionPage) (*interaction.GetReactionUsersResponse, error) {
	path, err := reactionPath(channelID, messageID, emojiType, emojiID)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	if page.Cookie != "" {
		q.Set("cookie", page.Cookie)
	}
	if page.Limit > 0 {
		q.Set("limit", strconv.Itoa(page.Limit))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return apiwire.Do[interaction.GetReactionUsersResponse](ctx, c.requester(), http.MethodGet, path, nil)
}

func reactionPath(channelID, messageID string, emojiType interaction.EmojiType, emojiID string) (string, error) {
	return apiwire.Path(api.CHANNEL_EMOJI, channelID, messageID, strconv.Itoa(int(emojiType)), emojiID)
}

// AddReaction 使用当前事件的客户端添加表态。
func AddReaction(c *types.Context, channelID, messageID string, emojiType interaction.EmojiType, emojiID string) error {
	if c == nil {
		return fmt.Errorf("event context is required")
	}
	return NewClient(c.Client).AddReaction(c.RequestContext(), channelID, messageID, emojiType, emojiID)
}

// RemoveReaction 使用当前事件的客户端删除表态。
func RemoveReaction(c *types.Context, channelID, messageID string, emojiType interaction.EmojiType, emojiID string) error {
	if c == nil {
		return fmt.Errorf("event context is required")
	}
	return NewClient(c.Client).RemoveReaction(c.RequestContext(), channelID, messageID, emojiType, emojiID)
}

// GetReactionUsers 使用当前事件的客户端查询一页表态用户。
func GetReactionUsers(c *types.Context, channelID, messageID string, emojiType interaction.EmojiType, emojiID, cookie string, limit int) (*interaction.GetReactionUsersResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("event context is required")
	}
	return NewClient(c.Client).ListReactionUsers(c.RequestContext(), channelID, messageID, emojiType, emojiID, ReactionPage{Cookie: cookie, Limit: limit})
}
