// Package robot 管理 QQ 机器人资料、分享链接、菜单和指令面板。
package robot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
	"github.com/ProgramCX/GoQQBot/pkg/api"
)

// Client 复用 API 客户端的凭证和 HTTP 设置。
type Client struct{ base *api.Client }

// NewClient 创建机器人管理客户端。
func NewClient(base *api.Client) *Client { return &Client{base: base} }
func (c *Client) requester() *api.Client {
	if c == nil {
		return nil
	}
	return c.base
}

// Profile 表示机器人账号资料。
type Profile struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Avatar           string `json:"avatar"`
	Bot              bool   `json:"bot"`
	UnionOpenID      string `json:"union_openid"`
	UnionUserAccount string `json:"union_user_account"`
	ShareURL         string `json:"share_url"`
	WelcomeMessage   string `json:"welcome_msg"`
}

// ShareLink 表示QQ 返回的好友邀请链接。
type ShareLink struct {
	URL string `json:"url"`
}

// Menu 表示单聊全局菜单。
type Menu struct {
	Items []MenuItem `json:"items"`
}

// MenuItem 表示顶层菜单项，额外字段随类型选择。
type MenuItem struct {
	Name         string        `json:"name"`
	Type         MenuItemType  `json:"type"`
	SubMenuItems []SubMenuItem `json:"sub_menu_items,omitempty"`
	SendMessage  string        `json:"send_message,omitempty"`
	Link         string        `json:"link,omitempty"`
	Switch       *MenuSwitch   `json:"switch,omitempty"`
}

// SubMenuItem 表示折叠菜单下的操作项。
type SubMenuItem struct {
	Name        string       `json:"name"`
	Type        MenuItemType `json:"type"`
	SendMessage string       `json:"send_message,omitempty"`
	Link        string       `json:"link,omitempty"`
}

// MenuSwitch 表示菜单开关标识和默认状态。
type MenuSwitch struct {
	ID      string `json:"switch_id"`
	Default bool   `json:"default"`
}

// MenuState 表示当前菜单和版本，未配置时 Menu 为空。
type MenuState struct {
	Menu    *Menu `json:"menu"`
	Version int64 `json:"version"`
}

// VersionResult 表示写入成功后 QQ 返回的配置版本。
type VersionResult struct {
	Version int64 `json:"version"`
}

// Panel 表示面板元素及备注。
type Panel struct {
	Items   []PanelItem `json:"items"`
	Remark  string      `json:"remark,omitempty"`
	Version int64       `json:"version,omitempty"`
}

// PanelItem 表示面板中的指令或链接。
type PanelItem struct {
	Name        string        `json:"name"`
	Description string        `json:"desc,omitempty"`
	Type        PanelItemType `json:"type"`
	OnlyAdmin   bool          `json:"only_admin"`
	Link        string        `json:"link,omitempty"`
}

// PanelRecord 表示面板配置、作用范围及关联对象。
type PanelRecord struct {
	ID           string          `json:"panel_id"`
	Scope        PanelScope      `json:"scope"`
	TargetType   PanelTargetType `json:"target_type"`
	Panel        Panel           `json:"panel"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
	Version      int64           `json:"version"`
	UserOpenIDs  []string        `json:"user_openids"`
	GroupOpenIDs []string        `json:"group_openids"`
}

// PanelPage 表示一页面板记录及下一页游标。
type PanelPage struct {
	Records    []PanelRecord `json:"records"`
	NextCursor string        `json:"next_cursor"`
	IsEnd      bool          `json:"is_end"`
}

// CreatePanelRequest 表示创建面板的场景、关联对象与配置。
type CreatePanelRequest struct {
	Scope        PanelScope      `json:"scope"`
	TargetType   PanelTargetType `json:"target_type,omitempty"`
	UserOpenIDs  []string        `json:"user_openids,omitempty"`
	GroupOpenIDs []string        `json:"group_openids,omitempty"`
	Panel        Panel           `json:"panel"`
}

// CreatePanelResult 表示新建面板的 ID。
type CreatePanelResult struct {
	ID string `json:"panel_id"`
}

// PanelTargetChange 表示指定面板关联对象的变更。
type PanelTargetChange struct {
	Action       PanelTargetAction `json:"op"`
	UserOpenIDs  []string          `json:"user_openids,omitempty"`
	GroupOpenIDs []string          `json:"group_openids,omitempty"`
}

// MenuItemType 表示菜单操作类型。
type MenuItemType string

// 菜单操作类型。
const (
	MenuSendMessage MenuItemType = "send_message"
	MenuLink        MenuItemType = "link"
	MenuFolder      MenuItemType = "menu"
	MenuToggle      MenuItemType = "switch"
)

// PanelItemType 表示面板元素类型。
type PanelItemType string

// 面板元素类型。
const (
	PanelCommand PanelItemType = "command"
	PanelLink    PanelItemType = "link"
)

// PanelScope 表示指令面板生效场景。
type PanelScope string

// 指令面板生效场景。
const (
	ScopeC2C     PanelScope = "c2c"
	ScopeGroup   PanelScope = "group"
	ScopeChannel PanelScope = "channel"
	ScopeDM      PanelScope = "dm"
)

// PanelTargetType 表示面板作用范围。
type PanelTargetType string

// 面板作用范围。
const (
	TargetAll      PanelTargetType = "all"
	TargetSpecific PanelTargetType = "specific"
)

// PanelTargetAction 表示关联对象的操作。
type PanelTargetAction string

// 关联对象的操作。
const (
	TargetAdd    PanelTargetAction = "add"
	TargetDelete PanelTargetAction = "del"
)

// GetProfile 获取当前机器人的资料。
func (c *Client) GetProfile(ctx context.Context) (*Profile, error) {
	return apiwire.Do[Profile](ctx, c.requester(), http.MethodGet, "/users/@me", nil)
}

// GenerateShareLink 生成好友邀请链接，callbackData 最多 32 个字符。
func (c *Client) GenerateShareLink(ctx context.Context, callbackData string) (*ShareLink, error) {
	if utf8.RuneCountInString(callbackData) > 32 {
		return nil, fmt.Errorf("callback_data exceeds 32 characters")
	}
	body := struct {
		CallbackData string `json:"callback_data,omitempty"`
	}{callbackData}
	result, err := apiwire.Do[struct {
		Data ShareLink `json:"data"`
	}](ctx, c.requester(), http.MethodPost, "/v2/generate_url_link", body)
	if err != nil {
		return nil, err
	}
	if result.Data.URL == "" {
		return nil, fmt.Errorf("share response has no URL")
	}
	return &result.Data, nil
}

// GetMenu 获取当前全局单聊菜单和版本。
func (c *Client) GetMenu(ctx context.Context) (*MenuState, error) {
	return apiwire.Do[MenuState](ctx, c.requester(), http.MethodGet, "/v2/menu", nil)
}

// UpdateMenu 覆盖整个全局单聊菜单，对该应用的所有单聊用户生效。
func (c *Client) UpdateMenu(ctx context.Context, menu Menu) (*VersionResult, error) {
	if len(menu.Items) > 10 {
		return nil, fmt.Errorf("menu exceeds 10 items")
	}
	for _, item := range menu.Items {
		if len(item.SubMenuItems) > 5 {
			return nil, fmt.Errorf("submenu exceeds 5 items")
		}
	}
	if menu.Items == nil {
		menu.Items = []MenuItem{}
	}
	return apiwire.Do[VersionResult](ctx, c.requester(), http.MethodPut, "/v2/menu", struct {
		Menu Menu `json:"menu"`
	}{menu})
}

// ListPanels 查询指定场景的一页面板，limit 为零时使用平台默认值。
func (c *Client) ListPanels(ctx context.Context, scope PanelScope, cursor string, limit int) (*PanelPage, error) {
	if scope == "" || limit < 0 || limit > 50 {
		return nil, fmt.Errorf("panel scope is required and limit must be within 0..50")
	}
	q := url.Values{"scope": {string(scope)}}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	return apiwire.Do[PanelPage](ctx, c.requester(), http.MethodGet, "/v2/panels?"+q.Encode(), nil)
}

// CreatePanel 创建指令面板，按指定对象生效仅支持单聊及群聊。
func (c *Client) CreatePanel(ctx context.Context, body CreatePanelRequest) (*CreatePanelResult, error) {
	if body.Scope == "" || len(body.Panel.Items) > 20 {
		return nil, fmt.Errorf("panel scope is required and at most 20 items are allowed")
	}
	return apiwire.Do[CreatePanelResult](ctx, c.requester(), http.MethodPost, "/v2/panels", body)
}

// GetPanel 读取面板详情和关联对象。
func (c *Client) GetPanel(ctx context.Context, id string) (*PanelRecord, error) {
	path, err := apiwire.Path("/v2/panels/%s", id)
	if err != nil {
		return nil, err
	}
	return apiwire.Do[PanelRecord](ctx, c.requester(), http.MethodGet, path, nil)
}

// UpdatePanel 替换面板内容，保留关联对象。
func (c *Client) UpdatePanel(ctx context.Context, id string, body Panel) (*VersionResult, error) {
	path, err := apiwire.Path("/v2/panels/%s", id)
	if err != nil {
		return nil, err
	}
	if len(body.Items) > 20 {
		return nil, fmt.Errorf("panel exceeds 20 items")
	}
	return apiwire.Do[VersionResult](ctx, c.requester(), http.MethodPut, path, struct {
		Panel Panel `json:"panel"`
	}{body})
}

// DeletePanel 删除指定面板。
func (c *Client) DeletePanel(ctx context.Context, id string) error {
	path, err := apiwire.Path("/v2/panels/%s", id)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, nil)
}

// UpdatePanelTarget 增删指定面板的关联对象。
func (c *Client) UpdatePanelTarget(ctx context.Context, id string, body PanelTargetChange) error {
	path, err := apiwire.Path("/v2/panels/%s/target", id)
	if err != nil {
		return err
	}
	if body.Action != TargetAdd && body.Action != TargetDelete {
		return fmt.Errorf("panel target action must be add or del")
	}
	if (len(body.UserOpenIDs) == 0) == (len(body.GroupOpenIDs) == 0) || len(body.UserOpenIDs)+len(body.GroupOpenIDs) > 20 {
		return fmt.Errorf("supply 1..20 user or group OpenIDs")
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPut, path, body)
}
