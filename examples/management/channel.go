package main

import (
	"context"
	"fmt"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message"
	channelmsg "github.com/ProgramCX/GoQQBot/pkg/message/channel"
	"github.com/ProgramCX/GoQQBot/pkg/message/interaction"
)

func registerChannel(ops map[string]operation, c *channel.Client) {
	add(ops, "guilds", "", false, func(ctx context.Context, _ []string, body channel.GuildListOptions) ([]channel.Guild, error) {
		return c.ListGuilds(ctx, body)
	})
	add(ops, "guild", "guild-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Guild, error) {
		return c.GetGuild(ctx, ids[0])
	})
	add(ops, "channels", "guild-id", false, func(ctx context.Context, ids []string, _ noInput) ([]channel.Channel, error) {
		return c.ListChannels(ctx, ids[0])
	})
	add(ops, "channel", "channel-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Channel, error) {
		return c.GetChannel(ctx, ids[0])
	})
	add(ops, "channel-create", "guild-id", true, func(ctx context.Context, ids []string, body channel.ChannelCreate) (*channel.Channel, error) {
		return c.CreateChannel(ctx, ids[0], body)
	})
	add(ops, "channel-update", "channel-id", true, func(ctx context.Context, ids []string, body channel.ChannelPatch) (*channel.Channel, error) {
		return c.UpdateChannel(ctx, ids[0], body)
	})
	add(ops, "channel-delete", "channel-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeleteChannel(ctx, ids[0]))
	})
	add(ops, "guild-members", "guild-id", false, func(ctx context.Context, ids []string, body channel.MemberListOptions) ([]channel.Member, error) {
		return c.ListMembers(ctx, ids[0], body)
	})
	add(ops, "guild-member", "guild-id user-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Member, error) {
		return c.GetMember(ctx, ids[0], ids[1])
	})
	add(ops, "guild-remove-member", "guild-id user-id", true, func(ctx context.Context, ids []string, body channel.RemoveMemberOptions) (any, error) {
		return ok(c.RemoveMember(ctx, ids[0], ids[1], body))
	})
	add(ops, "channel-online", "channel-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.OnlineCount, error) {
		return c.GetOnlineCount(ctx, ids[0])
	})
	add(ops, "guild-roles", "guild-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.RoleList, error) {
		return c.ListRoles(ctx, ids[0])
	})
	add(ops, "guild-create-role", "guild-id", true, func(ctx context.Context, ids []string, body channel.RoleOptions) (*channel.RoleResult, error) {
		return c.CreateRole(ctx, ids[0], body)
	})
	add(ops, "guild-update-role", "guild-id role-id", true, func(ctx context.Context, ids []string, body channel.RoleOptions) (*channel.RoleResult, error) {
		return c.UpdateRole(ctx, ids[0], ids[1], body)
	})
	add(ops, "guild-delete-role", "guild-id role-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeleteRole(ctx, ids[0], ids[1]))
	})
	add(ops, "guild-role-members", "guild-id role-id", false, func(ctx context.Context, ids []string, body channel.RoleMemberListOptions) (*channel.RoleMemberPage, error) {
		return c.ListRoleMembers(ctx, ids[0], ids[1], body)
	})
	add(ops, "guild-add-member-role", "guild-id user-id role-id", true, func(ctx context.Context, ids []string, body channel.RoleMembershipOptions) (any, error) {
		return ok(c.AddMemberRole(ctx, ids[0], ids[1], ids[2], body))
	})
	add(ops, "guild-remove-member-role", "guild-id user-id role-id", true, func(ctx context.Context, ids []string, body channel.RoleMembershipOptions) (any, error) {
		return ok(c.RemoveMemberRole(ctx, ids[0], ids[1], ids[2], body))
	})
	add(ops, "channel-member-permissions", "channel-id user-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Permissions, error) {
		return c.GetMemberPermissions(ctx, ids[0], ids[1])
	})
	add(ops, "channel-update-member-permissions", "channel-id user-id", true, func(ctx context.Context, ids []string, body channel.PermissionChanges) (any, error) {
		return ok(c.UpdateMemberPermissions(ctx, ids[0], ids[1], body))
	})
	add(ops, "channel-role-permissions", "channel-id role-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Permissions, error) {
		return c.GetRolePermissions(ctx, ids[0], ids[1])
	})
	add(ops, "channel-update-role-permissions", "channel-id role-id", true, func(ctx context.Context, ids []string, body channel.PermissionChanges) (any, error) {
		return ok(c.UpdateRolePermissions(ctx, ids[0], ids[1], body))
	})
	add(ops, "guild-api-permissions", "guild-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.APIPermissionList, error) {
		return c.ListAPIPermissions(ctx, ids[0])
	})
	add(ops, "guild-request-api-permission", "guild-id", true, func(ctx context.Context, ids []string, body channel.APIPermissionRequest) (*channel.APIPermissionDemand, error) {
		return c.RequestAPIPermission(ctx, ids[0], body)
	})
	add(ops, "guild-mute", "guild-id", true, func(ctx context.Context, ids []string, body channel.MuteOptions) (any, error) {
		return ok(c.SetGuildMute(ctx, ids[0], body))
	})
	add(ops, "guild-member-mute", "guild-id user-id", true, func(ctx context.Context, ids []string, body channel.MuteOptions) (any, error) {
		return ok(c.SetMemberMute(ctx, ids[0], ids[1], body))
	})
	add(ops, "guild-message-settings", "guild-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.MessageSettings, error) {
		return c.GetMessageSettings(ctx, ids[0])
	})
	add(ops, "guild-create-announcement", "guild-id", true, func(ctx context.Context, ids []string, body channel.AnnouncementInput) (*channel.Announcement, error) {
		return c.CreateAnnouncement(ctx, ids[0], body)
	})
	add(ops, "guild-delete-announcement", "guild-id message-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeleteAnnouncement(ctx, ids[0], ids[1]))
	})
	add(ops, "channel-add-pin", "channel-id message-id", true, func(ctx context.Context, ids []string, _ noInput) (*channel.PinnedMessages, error) {
		return c.AddPin(ctx, ids[0], ids[1])
	})
	add(ops, "channel-pins", "channel-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.PinnedMessages, error) {
		return c.ListPins(ctx, ids[0])
	})
	add(ops, "channel-delete-pin", "channel-id message-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeletePin(ctx, ids[0], ids[1]))
	})
	add(ops, "channel-schedules", "channel-id", false, func(ctx context.Context, ids []string, body uint64) ([]channel.Schedule, error) {
		return c.ListSchedules(ctx, ids[0], body)
	})
	add(ops, "channel-schedule", "channel-id schedule-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Schedule, error) {
		return c.GetSchedule(ctx, ids[0], ids[1])
	})
	add(ops, "channel-create-schedule", "channel-id", true, func(ctx context.Context, ids []string, body channel.ScheduleInput) (*channel.Schedule, error) {
		return c.CreateSchedule(ctx, ids[0], body)
	})
	add(ops, "channel-update-schedule", "channel-id schedule-id", true, func(ctx context.Context, ids []string, body channel.ScheduleInput) (*channel.Schedule, error) {
		return c.UpdateSchedule(ctx, ids[0], ids[1], body)
	})
	add(ops, "channel-delete-schedule", "channel-id schedule-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeleteSchedule(ctx, ids[0], ids[1]))
	})
	add(ops, "channel-threads", "channel-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.ThreadPage, error) {
		return c.ListThreads(ctx, ids[0])
	})
	add(ops, "channel-thread", "channel-id thread-id", false, func(ctx context.Context, ids []string, _ noInput) (*channel.Thread, error) {
		return c.GetThread(ctx, ids[0], ids[1])
	})
	add(ops, "channel-create-thread", "channel-id", true, func(ctx context.Context, ids []string, body channel.ThreadInput) (*channel.ThreadTask, error) {
		return c.CreateThread(ctx, ids[0], body)
	})
	add(ops, "channel-delete-thread", "channel-id thread-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeleteThread(ctx, ids[0], ids[1]))
	})
	add(ops, "channel-audio", "channel-id", true, func(ctx context.Context, ids []string, body channel.AudioControl) (any, error) {
		return ok(c.ControlAudio(ctx, ids[0], body))
	})
	add(ops, "channel-join-mic", "channel-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.JoinMic(ctx, ids[0]))
	})
	add(ops, "channel-leave-mic", "channel-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.LeaveMic(ctx, ids[0]))
	})
	add(ops, "dm-create", "guild-id user-id", true, func(ctx context.Context, ids []string, _ noInput) (*channel.DirectSession, error) {
		return c.CreateDirectSession(ctx, ids[0], ids[1])
	})
	add(ops, "channel-recall", "channel-id message-id", true, func(ctx context.Context, ids []string, body channel.RecallOptions) (any, error) {
		return ok(c.RecallChannelMessage(ctx, ids[0], ids[1], body))
	})
	add(ops, "dm-recall", "session-id message-id", true, func(ctx context.Context, ids []string, body channel.RecallOptions) (any, error) {
		return ok(c.RecallDirectMessage(ctx, ids[0], ids[1], body))
	})

	add(ops, "guild-members-mute", "guild-id", true, func(ctx context.Context, ids []string, body membersMuteInput) (*channel.MuteMembersResult, error) {
		return c.SetMembersMute(ctx, ids[0], body.UserIDs, body.Mute)
	})
	add(ops, "channel-reaction-add", "channel-id message-id emoji-id", true, func(ctx context.Context, ids []string, body reactionInput) (any, error) {
		return ok(c.AddReaction(ctx, ids[0], ids[1], body.emojiType(), ids[2]))
	})
	add(ops, "channel-reaction-remove", "channel-id message-id emoji-id", true, func(ctx context.Context, ids []string, body reactionInput) (any, error) {
		return ok(c.RemoveReaction(ctx, ids[0], ids[1], body.emojiType(), ids[2]))
	})
	add(ops, "channel-reaction-users", "channel-id message-id emoji-id", false, func(ctx context.Context, ids []string, body reactionInput) (*interaction.GetReactionUsersResponse, error) {
		return c.ListReactionUsers(ctx, ids[0], ids[1], body.emojiType(), ids[2], channel.ReactionPage{Cookie: body.Cookie, Limit: body.Limit})
	})
	add(ops, "channel-send", "channel-id", true, func(ctx context.Context, ids []string, body sendInput) (*channelmsg.Message, error) {
		image, err := readImage(body.ImageFile)
		if err != nil {
			return nil, err
		}
		return sentMessage(c.SendChannelMessage(ctx, ids[0], body.ChannelSendMessage, image))
	})
	add(ops, "dm-send", "session-id", true, func(ctx context.Context, ids []string, body sendInput) (*channelmsg.Message, error) {
		image, err := readImage(body.ImageFile)
		if err != nil {
			return nil, err
		}
		return sentMessage(c.SendDirectMessage(ctx, ids[0], body.ChannelSendMessage, image))
	})
}

func sentMessage(result *message.ReplyResult, err error) (*channelmsg.Message, error) {
	if err != nil {
		return nil, err
	}
	sent, err := channel.DecodeSentMessage(result)
	if err != nil {
		return nil, fmt.Errorf("消息可能已发送，解析完整响应失败: %w", err)
	}
	return sent, nil
}

type reactionInput struct {
	Type   interaction.EmojiType `json:"type"`
	Cookie string                `json:"cookie"`
	Limit  int                   `json:"limit"`
}

func (r reactionInput) emojiType() interaction.EmojiType {
	if r.Type == 0 {
		return interaction.EmojiTypeSystem
	}
	return r.Type
}

type membersMuteInput struct {
	UserIDs []string            `json:"user_ids"`
	Mute    channel.MuteOptions `json:"mute"`
}

type sendInput struct {
	channelmsg.ChannelSendMessage
	ImageFile string `json:"image_file"`
}
