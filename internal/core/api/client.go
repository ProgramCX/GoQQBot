// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ProgramCX/GoQQBot/internal/foundation/logger"
	"github.com/ProgramCX/GoQQBot/pkg/config"
)

const (
	TokenURL = "https://bots.qq.com/app/getAppAccessToken"
	BaseURL  = "https://api.sgroup.qq.com"
)

type Client struct {
	appID string

	appSecret   string
	accessToken string
	expiresAt   time.Time
	mu          sync.RWMutex
	httpClient  *http.Client
}

func NewClient(configPath string) *Client {
	loadConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return nil
	}
	return newClient(loadConfig.Bot.AppId, loadConfig.Bot.AppSecret)
}

func newClient(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) EnsureToken() error {
	logger.DEBUG("INFO: %+v", c)
	c.mu.RLock()
	if c.accessToken != "" && time.Until(c.expiresAt) > 5*time.Minute {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Until(c.expiresAt) > 5*time.Minute {
		return nil
	}

	payload := map[string]string{
		"appId":        c.appID,
		"clientSecret": c.appSecret,
	}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(TokenURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		logger.ERROR("Get access token failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.ERROR("Get access token failed[%d]: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	expiresSec, _ := strconv.Atoi(result.ExpiresIn)
	c.accessToken = result.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(expiresSec) * time.Second)

	logger.INFO("Get access token success, expires in %s seconds", result.ExpiresIn)
	return nil
}

func (c *Client) Request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	if err := c.EnsureToken(); err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)

		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	url := BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	req.Header.Set("Authorization", "QQBot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed[%d]: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}

//func (c *Client) SendGroupMessage(ctx context.Context, groupID string, content string, msgId string, eventId string) error {
//	path := fmt.Sprintf("/v2/groups/%s/messages", groupID)
//	body := map[string]interface{}{
//		"content":  content,
//		"msg_type": 0,
//		"msg_id":   msgId,
//		"event_id": eventId,
//	}
//	_, err := c.Post(ctx, path, body)
//	return err
//}
//
//func (c *Client) SendGroupMarkdownMessage(ctx context.Context, groupID string, markdown string, msgId string, eventId string) error {
//	path := fmt.Sprintf("/v2/groups/%s/messages", groupID)
//	body := map[string]interface{}{
//		"markdown": markdown,
//		"msg_type": 2,
//		"msg_id":   msgId,
//		"event_id": eventId,
//	}
//	_, err := c.Post(ctx, path, body)
//	return err
//}
//
//func (c *Client) SendGroupButtonMessage(ctx context.Context, groupID string, keyboardId string,
//	markdownId string, params []MarkdownParam, msgId string, eventId string) error {
//	path := fmt.Sprintf("/v2/groups/%s/messages", groupID)
//
//	body := map[string]interface{}{
//		//"keyboard": map[string]interface{}{
//		//	"id": keyboardId,
//		//},
//		"markdown": map[string]interface{}{
//			"custom_template_id": markdownId,
//			"params":             params,
//		},
//		"msg_type": 2,
//		"msg_id":   msgId,
//		"event_id": eventId,
//	}
//
//	debugData, _ := json.MarshalIndent(body, "", "  ")
//	logger.DEBUG("SendGroupButtonMessage 请求体:\n%s", string(debugData))
//
//	_, err := c.Post(ctx, path, body)
//	return err
//}
//
//type MarkdownParam struct {
//	Key    string   `json:"key"`
//	Values []string `json:"values"`
//}
