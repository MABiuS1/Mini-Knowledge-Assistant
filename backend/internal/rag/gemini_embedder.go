package rag

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

type GeminiEmbedder struct {
	baseURL    string
	apiKey     string
	model      string
	dimensions int
	httpClient *http.Client
}

func NewGeminiEmbedder(baseURL string, apiKey string, model string, dimensions int, timeout time.Duration) *GeminiEmbedder {
	return &GeminiEmbedder{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		dimensions: dimensions,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (e *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	requestBody := geminiEmbedRequest{
		Content: geminiEmbedContent{
			Parts: []geminiEmbedPart{{Text: text}},
		},
		OutputDimensionality: e.dimensions,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/v1beta/models/%s:embedContent?key=%s", e.baseURL, url.PathEscape(e.model), url.QueryEscape(e.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var errorResponse geminiEmbedErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil && errorResponse.Error.Message != "" {
			return nil, fmt.Errorf("gemini embedding error: %s", errorResponse.Error.Message)
		}
		return nil, fmt.Errorf("gemini embedding request failed with status %d", resp.StatusCode)
	}

	var response geminiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Embedding.Values, nil
}

type geminiEmbedRequest struct {
	Content              geminiEmbedContent `json:"content"`
	OutputDimensionality int                `json:"outputDimensionality,omitempty"`
}

type geminiEmbedContent struct {
	Parts []geminiEmbedPart `json:"parts"`
}

type geminiEmbedPart struct {
	Text string `json:"text"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

type geminiEmbedErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
