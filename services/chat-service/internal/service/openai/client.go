package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"chat-service/configs"
	zlog "packages/logger"
)

// Client represents an OpenAI API client
type Client interface {
	ChatCompletion(ctx context.Context, messages []Message, model string, temperature float64, maxTokens int) (*ChatCompletionResponse, error)
}

// client implements the OpenAI API client
type client struct {
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	logger       *zlog.Logger
	defaultModel string
}

// Message represents a chat message for OpenAI
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request to OpenAI
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// ChatCompletionResponse represents the response from OpenAI
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewClient creates a new OpenAI client
func NewClient(cfg *configs.Config, logger *zlog.Logger) Client {
	return &client{
		apiKey:       cfg.OpenAIAPIKey,
		baseURL:      "https://api.openai.com/v1",
		defaultModel: cfg.OpenAIModel,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.OpenAITimeout) * time.Second,
		},
		logger: logger,
	}
}

// ChatCompletion sends a chat completion request to OpenAI
func (c *client) ChatCompletion(ctx context.Context, messages []Message, model string, temperature float64, maxTokens int) (*ChatCompletionResponse, error) {
	if model == "" {
		model = c.defaultModel
	}

	requestBody := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	c.logger.Debug(ctx, "Sending request to OpenAI", map[string]interface{}{
		"model":       model,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"messages":    len(messages),
	})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error(ctx, fmt.Errorf("OpenAI API error: %s", string(body)), "OpenAI API returned non-200 status", resp.StatusCode)
		return nil, fmt.Errorf("OpenAI API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debug(ctx, "Received response from OpenAI", map[string]interface{}{
		"model":        response.Model,
		"total_tokens": response.Usage.TotalTokens,
		"choices":      len(response.Choices),
	})

	return &response, nil
}

// GetFirstChoiceContent returns the content of the first choice
func (r *ChatCompletionResponse) GetFirstChoiceContent() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}

// GetTotalTokens returns the total tokens used
func (r *ChatCompletionResponse) GetTotalTokens() int {
	return r.Usage.TotalTokens
}
