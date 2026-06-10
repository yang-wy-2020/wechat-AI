package openai

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

func (p *Provider) Name() string { return "openai" }

type requestBody struct {
	Model       string            `json:"model"`
	Messages    []llm.ChatMessage `json:"messages"`
	MaxTokens   uint              `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
}

type responseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []struct {
		Message      llm.ChatMessage `json:"message"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
	Usage map[string]interface{} `json:"usage"`
}

func (p *Provider) Chat(messages []llm.ChatMessage) (string, error) {
	cfg := config.LoadConfig()

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/"
	}

	body := requestBody{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("openai request: %s", string(data)))

	req, err := http.NewRequest("POST", baseURL+"chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.ApiKey)

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
		return "", errors.New(fmt.Sprintf("openai api error, status: %d, body: %s", resp.StatusCode, string(respData)))
	}

	logger.Info(fmt.Sprintf("openai response: %s", string(respData)))

	var result responseBody
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		reply := result.Choices[0].Message.Content
		logger.Info(fmt.Sprintf("openai reply: %s", reply))
		return reply, nil
	}

	return "", errors.New("openai: empty response")
}
