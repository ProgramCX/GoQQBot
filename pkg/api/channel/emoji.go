// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package channel

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
	"github.com/ProgramCX/GoQQBot/shared/types"
)

func AddReaction(c *types.Context, channelId string, messageId string, emojiType interaction.EmojiType, emojiId string) error {
	path := fmt.Sprintf(api.CHANNEL_EMOJI, channelId, messageId, strconv.Itoa(int(emojiType)), emojiId)
	_, err := c.Client.Request(c.RequestContext(), "PUT", path, nil)
	if err != nil {
		return fmt.Errorf("add reaction failed: %w", err)
	}
	return nil
}

func RemoveReaction(c *types.Context, channelId string, messageId string, emojiType interaction.EmojiType, emojiId string) error {
	path := fmt.Sprintf(api.CHANNEL_EMOJI, channelId, messageId, strconv.Itoa(int(emojiType)), emojiId)
	_, err := c.Client.Request(c.RequestContext(), "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("remove reaction failed: %w", err)
	}
	return nil
}

func GetReactionUsers(c *types.Context, channelId string, messageId string, emojiType interaction.EmojiType, emojiId string, cookie string, limit int) (*interaction.GetReactionUsersResponse, error) {
	path := fmt.Sprintf(api.CHANNEL_EMOJI, channelId, messageId, strconv.Itoa(int(emojiType)), emojiId)
	sep := "?"
	if cookie != "" {
		path += sep + "cookie=" + cookie
		sep = "&"
	}
	if limit > 0 {
		path += sep + "limit=" + strconv.Itoa(limit)
	}
	data, err := c.Client.Get(c.RequestContext(), path)
	if err != nil {
		return nil, fmt.Errorf("get reaction users failed: %w", err)
	}
	var resp interaction.GetReactionUsersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal reaction users response failed: %w", err)
	}
	return &resp, nil
}
