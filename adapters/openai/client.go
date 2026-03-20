package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/linkai/go-chatbot-api/domain/ports"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type chatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []chatCompletionMessage `json:"messages"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Complete(ctx context.Context, request ports.AIChatRequest) (ports.AIChatResponse, error) {
	messages := make([]chatCompletionMessage, 0, len(request.Conversation)+1)
	messages = append(messages, chatCompletionMessage{
		Role:    "system",
		Content: request.SystemPrompt,
	})
	for _, message := range request.Conversation {
		messages = append(messages, chatCompletionMessage{
			Role:    string(message.Role),
			Content: message.Content,
		})
	}

	payload, err := json.Marshal(chatCompletionRequest{
		Model:    request.Model,
		Messages: messages,
	})
	if err != nil {
		return ports.AIChatResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return ports.AIChatResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ports.AIChatResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ports.AIChatResponse{}, err
	}
	if resp.StatusCode >= 400 {
		return ports.AIChatResponse{}, fmt.Errorf("openai request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ports.AIChatResponse{}, err
	}
	if len(decoded.Choices) == 0 {
		return ports.AIChatResponse{}, fmt.Errorf("openai request failed: missing choices")
	}
	return ports.AIChatResponse{
		Content:      strings.TrimSpace(decoded.Choices[0].Message.Content),
		InputTokens:  decoded.Usage.PromptTokens,
		OutputTokens: decoded.Usage.CompletionTokens,
		FinishReason: decoded.Choices[0].FinishReason,
	}, nil
}
