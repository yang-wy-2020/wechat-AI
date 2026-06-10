package claude

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yang-wy-2020/wechat-AI/config"
	"github.com/yang-wy-2020/wechat-AI/pkg/llm"
	"github.com/yang-wy-2020/wechat-AI/pkg/logger"
)

type Provider struct{}

func (p *Provider) Name() string { return "claude" }

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestBody struct {
	Model       string    `json:"model"`
	MaxTokens   uint      `json:"max_tokens"`
	Messages    []message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type responseBody struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (p *Provider) Chat(messages []llm.ChatMessage) (string, error) {
	cfg := config.LoadConfig()

	systemText := ""
	claudeMessages := make([]message, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			systemText = m.Content
		} else {
			claudeMessages = append(claudeMessages, message{Role: m.Role, Content: m.Content})
		}
	}

	body := requestBody{
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Messages:    claudeMessages,
		System:      systemText,
		Temperature: cfg.Temperature,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("claude request: %s", string(data)))

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/"
	}

	req, err := http.NewRequest("POST", baseURL+"messages", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.ApiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("claude api error, status: %d, body: %s", resp.StatusCode, string(respData)))
	}

	logger.Info(fmt.Sprintf("claude response: %s", string(respData)))

	var result responseBody
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return "", err
	}

	if len(result.Content) > 0 && result.Content[0].Type == "text" {
		reply := result.Content[0].Text
		logger.Info(fmt.Sprintf("claude reply: %s", reply))
		return reply, nil
	}

	return "", errors.New("claude: empty response")
}
