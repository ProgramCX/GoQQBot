// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ProgramCX/GoQQBot/internal/core/api"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

func AddReaction(channelId string, messageId string, emojiType interaction.EmojiType, emojiId string, configPath string) error {
	path := fmt.Sprintf(api.CHANNEL_EMOJI, channelId, messageId, strconv.Itoa(int(emojiType)), emojiId)
	_, err := api.NewClient(configPath).Put(context.Background(), path, nil)
	if err != nil {
		return fmt.Errorf("add reaction failed: %w", err)
	}
	return nil
}

func RemoveReaction(channelId string, messageId string, emojiType interaction.EmojiType, emojiId string, configPath string) error {
	path := fmt.Sprintf(api.CHANNEL_EMOJI, channelId, messageId, strconv.Itoa(int(emojiType)), emojiId)
	_, err := api.NewClient(configPath).Delete(context.Background(), path)
	if err != nil {
		return fmt.Errorf("remove reaction failed: %w", err)
	}
	return nil
}

func GetReactionUsers(channelId string, messageId string, emojiType interaction.EmojiType, emojiId string, cookie string, limit int, configPath string) (*interaction.GetReactionUsersResponse, error) {
	path := fmt.Sprintf(api.CHANNEL_EMOJI, channelId, messageId, strconv.Itoa(int(emojiType)), emojiId)
	sep := "?"
	if cookie != "" {
		path += sep + "cookie=" + cookie
		sep = "&"
	}
	if limit > 0 {
		path += sep + "limit=" + strconv.Itoa(limit)
	}
	data, err := api.NewClient(configPath).Get(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("get reaction users failed: %w", err)
	}
	var resp interaction.GetReactionUsersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal reaction users response failed: %w", err)
	}
	return &resp, nil
}
