package group

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ProgramCX/GoQQBot/internal/apiwire"
)

// StrategyStatus 是自动审批策略的启用状态。
type StrategyStatus string

const (
	// StrategyOn 启用策略。
	StrategyOn StrategyStatus = "on"
	// StrategyOff 停用策略。
	StrategyOff StrategyStatus = "off"
)

// GroupNumbers 保存响应中的 QQ 群号，也接受平台返回的脱敏字符串。
type GroupNumbers []string

// UnmarshalJSON 将数字形式的 QQ 群号保留为十进制文本，避免精度丢失。
func (g *GroupNumbers) UnmarshalJSON(data []byte) error {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	var numbers GroupNumbers
	if items != nil {
		numbers = make(GroupNumbers, len(items))
	}
	for i, item := range items {
		if item[0] == '"' {
			if err := json.Unmarshal(item, &numbers[i]); err != nil {
				return err
			}
			continue
		}
		if _, err := strconv.ParseUint(string(item), 10, 64); err != nil {
			return fmt.Errorf("invalid group_ids entry at index %d", i)
		}
		numbers[i] = string(item)
	}
	*g = numbers
	return nil
}

// Strategy 是 QQ 返回的自动审批策略。
type Strategy struct {
	ID                 string         `json:"strategy_id"`
	GroupOpenIDs       []string       `json:"group_openids"`
	GroupIDs           GroupNumbers   `json:"group_ids"`
	WhitelistUserCount int            `json:"whitelist_user_count"`
	Status             StrategyStatus `json:"is_enable"`
	ExpiresAt          string         `json:"expire_at"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	Remark             string         `json:"remark"`
}

// StrategyPage 是一页自动审批策略。NextCursor 为空时已到末页。
type StrategyPage struct {
	Strategies []Strategy `json:"strategies"`
	NextCursor string     `json:"next_cursor"`
}

// ListStrategies 查询一页生效中的策略。Limit 为 0 时由平台默认取 20 条，最多 50 条。
func (c *Client) ListStrategies(ctx context.Context, page PageOptions) (*StrategyPage, error) {
	if err := page.validate(50); err != nil {
		return nil, err
	}
	// 该接口的官方 GET 示例明确使用 JSON 正文。
	return apiwire.Do[StrategyPage](ctx, c.requester(), http.MethodGet, strategyPath, page)
}

// StrategyGroups 指定关联群。GroupOpenIDs 和 GroupIDs 二选一。
type StrategyGroups struct {
	GroupOpenIDs []string `json:"group_openids,omitempty"`
	GroupIDs     []uint64 `json:"group_ids,omitempty"`
}

func (g StrategyGroups) validate() error {
	if (len(g.GroupOpenIDs) == 0) == (len(g.GroupIDs) == 0) {
		return fmt.Errorf("provide exactly one of group_openids or group_ids")
	}
	return nil
}

// CreateStrategyRequest 配置自动审批策略和关联群。关联群最多 100 个。
type CreateStrategyRequest struct {
	StrategyGroups
	// Status 为空时由 QQ 默认启用。
	Status StrategyStatus `json:"is_enable,omitempty"`
	// ExpiresAt 使用 RFC3339 格式。省略时由 QQ 设置一年有效期。
	ExpiresAt string `json:"expire_at,omitempty"`
	Remark    string `json:"remark,omitempty"`
}

// CreateStrategyResult 包含新策略的 ID、启用状态及过期时间。
type CreateStrategyResult struct {
	ID        string         `json:"strategy_id"`
	Status    StrategyStatus `json:"is_enable"`
	ExpiresAt string         `json:"expire_at"`
}

// CreateStrategy 创建自动审批策略。策略只在机器人拥有管理员身份的群生效。
func (c *Client) CreateStrategy(ctx context.Context, request CreateStrategyRequest) (*CreateStrategyResult, error) {
	if err := request.StrategyGroups.validate(); err != nil {
		return nil, err
	}
	if len(request.GroupOpenIDs)+len(request.GroupIDs) > 100 {
		return nil, fmt.Errorf("strategy may be associated with at most 100 groups")
	}
	return apiwire.Do[CreateStrategyResult](ctx, c.requester(), http.MethodPost, strategyPath, request)
}

// StrategyGroupChange 增删策略关联群。群标识类型应与创建时一致。
type StrategyGroupChange struct {
	Action ListAction `json:"op"`
	StrategyGroups
}

// UpdateStrategyRequest 修改指定字段。Remark 指向空串时发送空备注。
type UpdateStrategyRequest struct {
	Status      StrategyStatus       `json:"is_enable,omitempty"`
	ExpiresAt   string               `json:"expire_at,omitempty"`
	GroupChange *StrategyGroupChange `json:"group_action,omitempty"`
	Remark      *string              `json:"remark,omitempty"`
}

// UpdateStrategyResult 是修改后的启用状态与过期时间。
type UpdateStrategyResult struct {
	Status    StrategyStatus `json:"is_enable"`
	ExpiresAt string         `json:"expire_at"`
}

// UpdateStrategy 修改策略状态、有效期、关联群或备注。
func (c *Client) UpdateStrategy(ctx context.Context, strategyID string, request UpdateStrategyRequest) (*UpdateStrategyResult, error) {
	path, err := apiwire.Path(strategyPath+"/%s", strategyID)
	if err != nil {
		return nil, err
	}
	if change := request.GroupChange; change != nil {
		if change.Action != ListAdd && change.Action != ListDelete {
			return nil, fmt.Errorf("strategy group action must be add or del")
		}
		if err := change.StrategyGroups.validate(); err != nil {
			return nil, err
		}
	}
	return apiwire.Do[UpdateStrategyResult](ctx, c.requester(), http.MethodPatch, path, request)
}

// DeleteStrategy 删除指定的自动审批策略。
func (c *Client) DeleteStrategy(ctx context.Context, strategyID string) error {
	path, err := apiwire.Path(strategyPath+"/%s", strategyID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodDelete, path, struct{}{})
}

// ExecuteStrategy 请求 QQ 异步扫描关联群的待审批申请。返回成功表示请求已受理。
func (c *Client) ExecuteStrategy(ctx context.Context, strategyID string) error {
	path, err := apiwire.Path(strategyPath+"/%s/execute", strategyID)
	if err != nil {
		return err
	}
	return apiwire.Exec(ctx, c.requester(), http.MethodPost, path, struct{}{})
}

// UpdateStrategyWhitelistRequest 增删自动审批白名单中的 QQ 号。
type UpdateStrategyWhitelistRequest struct {
	Action ListAction `json:"op"`
	// Users 是字符串形式的 QQ 号。每次最多 10000 个。
	Users []string `json:"whitelist_users"`
}

// UpdateStrategyWhitelistResult 保存修改后的估算号码数和更新时间。
type UpdateStrategyWhitelistResult struct {
	StrategyID         string `json:"strategy_id"`
	WhitelistUserCount int    `json:"whitelist_user_count"`
	UpdatedAt          string `json:"updated_at"`
}

// UpdateStrategyWhitelist 修改指定策略的白名单。
func (c *Client) UpdateStrategyWhitelist(ctx context.Context, strategyID string, request UpdateStrategyWhitelistRequest) (*UpdateStrategyWhitelistResult, error) {
	path, err := apiwire.Path(strategyPath+"/%s/whitelist_users", strategyID)
	if err != nil {
		return nil, err
	}
	if request.Action != ListAdd && request.Action != ListDelete {
		return nil, fmt.Errorf("strategy whitelist action must be add or del")
	}
	if err := validateBatch(request.Users, 10000, "whitelist_users"); err != nil {
		return nil, err
	}
	return apiwire.Do[UpdateStrategyWhitelistResult](ctx, c.requester(), http.MethodPost, path, request)
}
