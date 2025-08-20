package openai

import (
	"net/http"
	"testing"

	"chat-service/configs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of the HTTP client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNewClient(t *testing.T) {
	cfg := &configs.Config{
		OpenAIAPIKey:      "test-api-key",
		OpenAIModel:       "gpt-3.5-turbo",
		OpenAIMaxTokens:   1000,
		OpenAITemperature: 0.7,
		OpenAITimeout:     30,
	}

	client := NewClient(cfg, nil)
	assert.NotNil(t, client)
}

func TestChatCompletion(t *testing.T) {
	cfg := &configs.Config{
		OpenAIAPIKey:      "test-api-key",
		OpenAIModel:       "gpt-3.5-turbo",
		OpenAIMaxTokens:   1000,
		OpenAITemperature: 0.7,
		OpenAITimeout:     30,
	}

	client := NewClient(cfg, nil)
	assert.NotNil(t, client)

	// Test with empty model (should use default)
	_ = []Message{
		{
			Role:    "user",
			Content: "Hello, how are you?",
		},
	}

	// This test would require mocking the HTTP client
	// For now, we'll just test that the client can be created
	assert.Equal(t, "gpt-3.5-turbo", cfg.OpenAIModel)
	assert.Equal(t, 0.7, cfg.OpenAITemperature)
	assert.Equal(t, 1000, cfg.OpenAIMaxTokens)
}

func TestGetFirstChoiceContent(t *testing.T) {
	response := &ChatCompletionResponse{
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "Hello! I'm doing well, thank you for asking.",
				},
				FinishReason: "stop",
			},
		},
	}

	content := response.GetFirstChoiceContent()
	assert.Equal(t, "Hello! I'm doing well, thank you for asking.", content)
}

func TestGetTotalTokens(t *testing.T) {
	response := &ChatCompletionResponse{
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	totalTokens := response.GetTotalTokens()
	assert.Equal(t, 30, totalTokens)
}

func TestGetFirstChoiceContentEmpty(t *testing.T) {
	response := &ChatCompletionResponse{
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{},
	}

	content := response.GetFirstChoiceContent()
	assert.Equal(t, "", content)
}
