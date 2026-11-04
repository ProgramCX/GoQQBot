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

// AudioStatus 表示对语音子频道执行的播放操作。
type AudioStatus uint32

// 各音频播放操作对应的协议值。
const (
	AudioStart  AudioStatus = 0
	AudioPause  AudioStatus = 1
	AudioResume AudioStatus = 2
	AudioStop   AudioStatus = 3
)

// AudioControl 设置播放操作。开始播放时需提供 AudioURL，Text 为可选的状态文案。
type AudioControl struct {
	Status   AudioStatus `json:"status"`
	AudioURL string      `json:"audio_url,omitempty"`
	Text     string      `json:"text,omitempty"`
}

// ControlAudio 控制语音子频道的音频播放。机器人需获得平台音频接口权限。
func (c *Client) ControlAudio(ctx context.Context, channelID string, input AudioControl) error {
	path, err := apiwire.Path("/channels/%s/audio", channelID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPost, path, input)
}

// JoinMic 让机器人在指定语音子频道上麦。
func (c *Client) JoinMic(ctx context.Context, channelID string) error {
	path, err := apiwire.Path("/channels/%s/mic", channelID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPut, path, struct{}{})
}

// LeaveMic 让机器人在指定语音子频道下麦。
func (c *Client) LeaveMic(ctx context.Context, channelID string) error {
	path, err := apiwire.Path("/channels/%s/mic", channelID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, struct{}{})
}
