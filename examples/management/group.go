package main

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api/group"
)

func registerGroup(ops map[string]operation, c *group.Client) {
	add(ops, "group-info", "group-openid", false, func(ctx context.Context, ids []string, _ noInput) (*group.Info, error) {
		return c.GetInfo(ctx, ids[0])
	})
	add(ops, "group-bot-state", "group-openid", false, func(ctx context.Context, ids []string, _ noInput) (*group.BotState, error) {
		return c.GetBotState(ctx, ids[0])
	})
	add(ops, "group-members", "group-openid", false, func(ctx context.Context, ids []string, body string) (*group.MemberPage, error) {
		return c.ListMembers(ctx, ids[0], body)
	})
	add(ops, "group-member", "group-openid member-openid", false, func(ctx context.Context, ids []string, _ noInput) (*group.Member, error) {
		return c.GetMember(ctx, ids[0], ids[1])
	})
	add(ops, "group-mutes", "group-openid", false, func(ctx context.Context, ids []string, _ noInput) (*group.MuteSettings, error) {
		return c.GetMuteSettings(ctx, ids[0])
	})
	add(ops, "group-set-mutes", "group-openid", true, func(ctx context.Context, ids []string, body []group.MemberMuteChange) (any, error) {
		return ok(c.SetMemberMutes(ctx, ids[0], body))
	})
	add(ops, "group-remove-members", "group-openid", true, func(ctx context.Context, ids []string, body group.RemoveMembersRequest) (*group.RemoveMembersResult, error) {
		return c.RemoveMembers(ctx, ids[0], body)
	})
	add(ops, "group-blacklist", "group-openid", false, func(ctx context.Context, ids []string, body group.PageOptions) (*group.BlacklistPage, error) {
		return c.ListBlacklist(ctx, ids[0], body)
	})
	add(ops, "group-update-blacklist", "group-openid", true, func(ctx context.Context, ids []string, body group.UpdateBlacklistRequest) (*group.UpdateBlacklistResult, error) {
		return c.UpdateBlacklist(ctx, ids[0], body)
	})
	add(ops, "group-join-requests", "group-openid", false, func(ctx context.Context, ids []string, body group.PageOptions) (*group.JoinRequestPage, error) {
		return c.ListJoinRequests(ctx, ids[0], body)
	})
	add(ops, "group-decide-join", "group-openid member-openid", true, func(ctx context.Context, ids []string, body group.DecideJoinRequest) (any, error) {
		return ok(c.DecideJoin(ctx, ids[0], ids[1], body))
	})
	add(ops, "group-strategies", "", false, func(ctx context.Context, _ []string, body group.PageOptions) (*group.StrategyPage, error) {
		return c.ListStrategies(ctx, body)
	})
	add(ops, "group-create-strategy", "", true, func(ctx context.Context, _ []string, body group.CreateStrategyRequest) (*group.CreateStrategyResult, error) {
		return c.CreateStrategy(ctx, body)
	})
	add(ops, "group-update-strategy", "strategy-id", true, func(ctx context.Context, ids []string, body group.UpdateStrategyRequest) (*group.UpdateStrategyResult, error) {
		return c.UpdateStrategy(ctx, ids[0], body)
	})
	add(ops, "group-delete-strategy", "strategy-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.DeleteStrategy(ctx, ids[0]))
	})
	add(ops, "group-execute-strategy", "strategy-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) {
		return ok(c.ExecuteStrategy(ctx, ids[0]))
	})
	add(ops, "group-strategy-whitelist", "strategy-id", true, func(ctx context.Context, ids []string, body group.UpdateStrategyWhitelistRequest) (*group.UpdateStrategyWhitelistResult, error) {
		return c.UpdateStrategyWhitelist(ctx, ids[0], body)
	})
}
