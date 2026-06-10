package qwen

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

const baseURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

type Provider struct{}

func (p *Provider) Name() string { return "qwen" }

type input struct {
	Messages []llm.ChatMessage `json:"messages"`
}

type parameters struct {
	MaxTokens   uint    `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type requestBody struct {
	Model      string     `json:"model"`
	Input      input      `json:"input"`
	Parameters parameters `json:"parameters"`
}

type responseBody struct {
	Output struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"output"`
}

func (p *Provider) Chat(messages []llm.ChatMessage) (string, error) {
	cfg := config.LoadConfig()

	body := requestBody{
		Model: cfg.Model,
		Input: input{Messages: messages},
		Parameters: parameters{
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("qwen request: %s", string(data)))

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(data))
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
		return "", errors.New(fmt.Sprintf("qwen api error, status: %d, body: %s", resp.StatusCode, string(respData)))
	}

	logger.Info(fmt.Sprintf("qwen response: %s", string(respData)))

	var result responseBody
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return "", err
	}

	if result.Output.Text != "" {
		logger.Info(fmt.Sprintf("qwen reply: %s", result.Output.Text))
		return result.Output.Text, nil
	}

	return "", errors.New("qwen: empty response")
}
