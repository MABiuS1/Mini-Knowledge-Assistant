package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GeminiClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiClient(baseURL string, apiKey string, model string, timeout time.Duration) *GeminiClient {
	return &GeminiClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *GeminiClient) Complete(ctx context.Context, messages []Message) (Completion, error) {
	requestBody := geminiGenerateRequest{
		Contents: make([]geminiContent, 0, len(messages)),
		GenerationConfig: geminiGenerationConfig{
			Temperature: 0.2,
		},
	}

	var systemParts []geminiPart
	for _, message := range messages {
		if message.Role == RoleSystem {
			systemParts = append(systemParts, geminiPart{Text: message.Content})
			continue
		}

		role := "user"
		if message.Role == RoleAssistant {
			role = "model"
		}

		requestBody.Contents = append(requestBody.Contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: message.Content}},
		})
	}

	if len(systemParts) > 0 {
		requestBody.SystemInstruction = &geminiContent{Parts: systemParts}
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return Completion{}, err
	}

	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", c.baseURL, url.PathEscape(c.model), url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Completion{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Completion{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var errorResponse geminiErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil && errorResponse.Error.Message != "" {
			return Completion{}, fmt.Errorf("gemini error: %s", errorResponse.Error.Message)
		}
		return Completion{}, fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
	}

	var response geminiGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Completion{}, err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return Completion{}, ErrAssistantEmptyMessage
	}

	var builder strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		builder.WriteString(part.Text)
	}

	model := response.ModelVersion
	if model == "" {
		model = c.model
	}

	return Completion{
		Content: builder.String(),
		Model:   model,
		Usage: Usage{
			PromptTokens:     response.UsageMetadata.PromptTokenCount,
			CompletionTokens: response.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      response.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

type geminiGenerateRequest struct {
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiGenerationConfig struct {
	Temperature float64 `json:"temperature"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	ModelVersion string `json:"modelVersion"`
}

type geminiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
