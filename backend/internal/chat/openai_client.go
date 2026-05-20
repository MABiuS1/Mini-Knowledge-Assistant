package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIClient(baseURL string, apiKey string, model string, timeout time.Duration) *OpenAIClient {
	return &OpenAIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *OpenAIClient) Complete(ctx context.Context, messages []Message) (Completion, error) {
	requestBody := openAIChatRequest{
		Model:       c.model,
		Temperature: 0.2,
		Messages:    make([]openAIMessage, 0, len(messages)),
	}

	for _, message := range messages {
		requestBody.Messages = append(requestBody.Messages, openAIMessage{
			Role:    string(message.Role),
			Content: message.Content,
		})
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return Completion{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return Completion{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var errorResponse openAIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil && errorResponse.Error.Message != "" {
			return Completion{}, fmt.Errorf("openai error: %s", errorResponse.Error.Message)
		}
		return Completion{}, fmt.Errorf("openai request failed with status %d", resp.StatusCode)
	}

	var response openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Completion{}, err
	}

	if len(response.Choices) == 0 {
		return Completion{}, ErrAssistantEmptyMessage
	}

	return Completion{
		Content: response.Choices[0].Message.Content,
		Model:   response.Model,
		Usage: Usage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}, nil
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
