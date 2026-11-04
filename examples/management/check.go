package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/api/group"
	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
	"github.com/ProgramCX/GoQQBot/pkg/config"
)

type checkInputs struct {
	GroupOpenID       string `json:"group_openid"`
	MemberOpenID      string `json:"member_openid"`
	GuildID           string `json:"guild_id"`
	ChannelID         string `json:"channel_id"`
	AudioChannelID    string `json:"audio_channel_id"`
	ForumChannelID    string `json:"forum_channel_id"`
	ScheduleChannelID string `json:"schedule_channel_id"`
	UserID            string `json:"user_id"`
	RoleID            string `json:"role_id"`
	MessageID         string `json:"message_id"`
	ScheduleID        string `json:"schedule_id"`
	ThreadID          string `json:"thread_id"`
	PanelID           string `json:"panel_id"`
}

type checkResult struct {
	Operation string `json:"operation"`
	Status    string `json:"status"`
	HTTP      int    `json:"http,omitempty"`
	Code      int    `json:"code,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func runChecks(configPath, inputPath string) error {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(strings.TrimPrefix(string(raw), "\ufeff")))
	decoder.DisallowUnknownFields()
	var input checkInputs
	if err := decoder.Decode(&input); err != nil {
		return fmt.Errorf("读取验收参数: %w", err)
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	client := api.NewClient(cfg.Bot.AppId, cfg.Bot.AppSecret)
	ops := make(map[string]operation)
	registerGroup(ops, group.NewClient(client))
	registerChannel(ops, channel.NewClient(client))
	registerRobot(ops, robot.NewClient(client))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	names := make([]string, 0, len(ops))
	for name := range ops {
		names = append(names, name)
	}
	sort.Strings(names)
	report := make([]checkResult, 0, len(ops))
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		report = append(report, checkOperation(ctx, name, ops[name], input))
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func checkOperation(ctx context.Context, name string, op operation, input checkInputs) checkResult {
	result := checkResult{Operation: name, Status: "manual"}
	if op.write {
		result.Reason = "需要在测试资源上显式执行写入"
		return result
	}
	ids, missing := input.arguments(name, op.ids)
	if missing != "" {
		result.Status = "missing_fixture"
		result.Reason = "缺少 " + missing
		return result
	}
	callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := op.run(callCtx, ids, nil)
	if err == nil {
		result.Status = "pass"
		return result
	}
	result.Status = "fail"
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		if errors.Is(err, api.ErrAPIAccessDenied) || errors.Is(err, api.ErrOperationDenied) {
			result.Status = "permission_blocked"
		}
		result.HTTP = apiErr.StatusCode
		result.Code = apiErr.Code
		result.TraceID = apiErr.TraceID
		result.Reason = apiErr.DiagnosticMessage()
	} else {
		result.Reason = err.Error()
	}
	return result
}

func (input checkInputs) arguments(name, spec string) ([]string, string) {
	channelID := input.ChannelID
	channelField := "channel_id"
	switch name {
	case "channel-online":
		channelID, channelField = input.AudioChannelID, "audio_channel_id"
	case "channel-schedules", "channel-schedule":
		channelID, channelField = input.ScheduleChannelID, "schedule_channel_id"
	case "channel-threads", "channel-thread":
		channelID, channelField = input.ForumChannelID, "forum_channel_id"
	}
	values := map[string]string{
		"group-openid": input.GroupOpenID, "member-openid": input.MemberOpenID,
		"guild-id": input.GuildID, "channel-id": channelID, "user-id": input.UserID,
		"role-id": input.RoleID, "message-id": input.MessageID, "emoji-id": "203",
		"schedule-id": input.ScheduleID, "thread-id": input.ThreadID,
		"scope": "c2c", "panel-id": input.PanelID,
	}
	var ids []string
	for _, key := range strings.Fields(spec) {
		if strings.TrimSpace(values[key]) == "" {
			if key == "channel-id" {
				return nil, channelField
			}
			return nil, strings.ReplaceAll(key, "-", "_")
		}
		ids = append(ids, values[key])
	}
	return ids, ""
}
