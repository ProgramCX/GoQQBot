package main

import (
	"context"

	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
)

func registerRobot(ops map[string]operation, c *robot.Client) {
	add(ops, "robot-profile", "", false, func(ctx context.Context, _ []string, _ noInput) (*robot.Profile, error) { return c.GetProfile(ctx) })
	add(ops, "robot-share", "", true, func(ctx context.Context, _ []string, body string) (*robot.ShareLink, error) {
		return c.GenerateShareLink(ctx, body)
	})
	add(ops, "robot-menu", "", false, func(ctx context.Context, _ []string, _ noInput) (*robot.MenuState, error) { return c.GetMenu(ctx) })
	add(ops, "robot-menu-update", "", true, func(ctx context.Context, _ []string, body robot.Menu) (*robot.VersionResult, error) {
		return c.UpdateMenu(ctx, body)
	})
	add(ops, "robot-panels", "scope", false, func(ctx context.Context, ids []string, body panelPageInput) (*robot.PanelPage, error) {
		return c.ListPanels(ctx, robot.PanelScope(ids[0]), body.Cursor, body.Limit)
	})
	add(ops, "robot-panel-create", "", true, func(ctx context.Context, _ []string, body robot.CreatePanelRequest) (*robot.CreatePanelResult, error) {
		return c.CreatePanel(ctx, body)
	})
	add(ops, "robot-panel", "panel-id", false, func(ctx context.Context, ids []string, _ noInput) (*robot.PanelRecord, error) {
		return c.GetPanel(ctx, ids[0])
	})
	add(ops, "robot-panel-update", "panel-id", true, func(ctx context.Context, ids []string, body robot.Panel) (*robot.VersionResult, error) {
		return c.UpdatePanel(ctx, ids[0], body)
	})
	add(ops, "robot-panel-target", "panel-id", true, func(ctx context.Context, ids []string, body robot.PanelTargetChange) (any, error) {
		return ok(c.UpdatePanelTarget(ctx, ids[0], body))
	})
	add(ops, "robot-panel-delete", "panel-id", true, func(ctx context.Context, ids []string, _ noInput) (any, error) { return ok(c.DeletePanel(ctx, ids[0])) })
}

type panelPageInput struct {
	Cursor string
	Limit  int
}
