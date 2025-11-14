package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIClient interface {
	AnalyzeAndFix(issue Issue, context *RepoContext) (*Fix, error)
}

type AIService interface {
	GetAvailableModels() ([]string, error)
}

type Fix struct {
	FileChanges    []FileChange
	Explanation    string
	Confidence     string // "high", "medium", "low"
	NeedsMoreInfo  bool
	Questions      []string
}

// OpenAI/ChatGPT Client
type OpenAIClient struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	analytics *SessionAnalytics
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (o *OpenAIClient) SetAnalytics(analytics *SessionAnalytics) {
	o.analytics = analytics
}

// xAI Client (Grok models)
type XAIClient struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	analytics *SessionAnalytics
}

func NewXAIClient(apiKey, model string) *XAIClient {
	if model == "" {
		model = "grok-beta"
	}
	return &XAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.x.ai/v1",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (x *XAIClient) SetAnalytics(analytics *SessionAnalytics) {
	x.analytics = analytics
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

func (o *OpenAIClient) AnalyzeAndFix(issue Issue, context *RepoContext) (*Fix, error) {
	// Track API call
	if o.analytics != nil {
		o.analytics.RecordAPICall("chatgpt")
	}

	prompt := o.buildPrompt(issue, context)

	reqBody := OpenAIRequest{
		Model: o.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are an expert software developer. Analyze issues and provide fixes in a structured JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2,
		MaxTokens:   8000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, err
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return o.parseFix(openaiResp.Choices[0].Message.Content)
}

func (o *OpenAIClient) buildPrompt(issue Issue, context *RepoContext) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("# Issue to Fix\n\n"))
	prompt.WriteString(fmt.Sprintf("**Title:** %s\n\n", issue.Title))
	prompt.WriteString(fmt.Sprintf("**Description:**\n%s\n\n", issue.Body))

	prompt.WriteString("# Repository Context\n\n")
	prompt.WriteString("## Directory Structure\n```\n")
	prompt.WriteString(context.Structure)
	prompt.WriteString("\n```\n\n")

	if len(context.Files) > 0 {
		prompt.WriteString("## Key Files\n\n")
		for path, content := range context.Files {
			// Limit content size
			if len(content) > 5000 {
				content = content[:5000] + "\n... (truncated)"
			}
			prompt.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content))
		}
	}

	prompt.WriteString(`# Task

Analyze the issue and provide a fix. Your response MUST be in the following JSON format:

{
  "confidence": "high|medium|low",
  "needs_more_info": false,
  "questions": [],
  "explanation": "Brief explanation of what the fix does",
  "files": [
    {
      "path": "relative/path/to/file.ext",
      "content": "complete file content with the fix applied"
    }
  ]
}

Instructions:
- If you're CONFIDENT you understand the issue and can fix it, set confidence to "high" and provide the fix
- If you need more information, set "needs_more_info" to true and list specific "questions" to ask in the issue
- Provide COMPLETE file content, not diffs or patches
- Only include files that need to be modified or created
- Keep explanations concise but clear
- Ensure the fix actually addresses the issue
- If you need to create a new file, include its full content
- Return valid JSON only, no markdown code blocks

Now provide the fix:`)

	return prompt.String()
}

func (o *OpenAIClient) parseFix(response string) (*Fix, error) {
	// Clean up markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result struct {
		Confidence    string   `json:"confidence"`
		NeedsMoreInfo bool     `json:"needs_more_info"`
		Questions     []string `json:"questions"`
		Explanation   string   `json:"explanation"`
		Files         []struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		} `json:"files"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\nResponse: %s", err, response)
	}

	fix := &Fix{
		Confidence:    result.Confidence,
		NeedsMoreInfo: result.NeedsMoreInfo,
		Questions:     result.Questions,
		Explanation:   result.Explanation,
		FileChanges:   make([]FileChange, len(result.Files)),
	}

	for i, file := range result.Files {
		fix.FileChanges[i] = FileChange{
			FilePath: file.Path,
			Content:  file.Content,
		}
	}

	return fix, nil
}

func (o *OpenAIClient) GetAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", o.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Return default models if API call fails
		return []string{
			"gpt-4o",
			"gpt-4o-mini",
			"gpt-4-turbo",
			"gpt-3.5-turbo",
		}, nil
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []string{"gpt-4o"}, nil
	}

	models := make([]string, 0)
	for _, m := range result.Data {
		// Only include GPT models
		if strings.HasPrefix(m.ID, "gpt-") {
			models = append(models, m.ID)
		}
	}

	if len(models) == 0 {
		return []string{"gpt-4o"}, nil
	}

	return models, nil
}

// Ollama Client (Free local AI: https://ollama.com)
type OllamaClient struct {
	baseURL   string
	model     string
	client    *http.Client
	analytics *SessionAnalytics
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 300 * time.Second}, // Longer timeout for local models
	}
}

func (o *OllamaClient) SetAnalytics(analytics *SessionAnalytics) {
	o.analytics = analytics
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (o *OllamaClient) AnalyzeAndFix(issue Issue, context *RepoContext) (*Fix, error) {
	// Track API call
	if o.analytics != nil {
		o.analytics.RecordAPICall("ollama")
	}

	prompt := o.buildPrompt(issue, context)

	reqBody := OllamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s - %s", resp.Status, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	return o.parseFix(ollamaResp.Response)
}

func (o *OllamaClient) buildPrompt(issue Issue, context *RepoContext) string {
	// Same prompt building logic as Groq
	g := &OpenAIClient{}
	return g.buildPrompt(issue, context)
}

func (o *OllamaClient) parseFix(response string) (*Fix, error) {
	// Same parsing logic as Groq
	g := &OpenAIClient{}
	return g.parseFix(response)
}

// xAI Client methods
func (x *XAIClient) AnalyzeAndFix(issue Issue, context *RepoContext) (*Fix, error) {
	// Track API call
	if x.analytics != nil {
		x.analytics.RecordAPICall("grok")
	}

	prompt := x.buildPrompt(issue, context)

	reqBody := OpenAIRequest{ // Uses same structure as Groq (OpenAI-compatible)
		Model: x.model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are an expert software developer. Analyze issues and provide fixes in a structured JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2,
		MaxTokens:   8000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", x.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+x.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("xAI API error: %s - %s", resp.Status, string(body))
	}

	var xaiResp OpenAIResponse // Uses same response structure
	if err := json.NewDecoder(resp.Body).Decode(&xaiResp); err != nil {
		return nil, err
	}

	if len(xaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return x.parseFix(xaiResp.Choices[0].Message.Content)
}

func (x *XAIClient) buildPrompt(issue Issue, context *RepoContext) string {
	// Same prompt building logic as Groq
	g := &OpenAIClient{}
	return g.buildPrompt(issue, context)
}

func (x *XAIClient) parseFix(response string) (*Fix, error) {
	// Same parsing logic as Groq
	g := &OpenAIClient{}
	return g.parseFix(response)
}

func (x *XAIClient) GetAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", x.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+x.apiKey)

	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{"grok-beta", "grok-vision-beta"}, nil
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []string{"grok-beta"}, nil
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}

	if len(models) == 0 {
		return []string{"grok-beta"}, nil
	}

	return models, nil
}

func (o *OllamaClient) GetAvailableModels() ([]string, error) {
	req, err := http.NewRequest("GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{"llama2", "codellama", "mistral"}, nil
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []string{"llama2"}, nil
	}

	models := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		models = append(models, m.Name)
	}

	if len(models) == 0 {
		return []string{"llama2"}, nil
	}

	return models, nil
}

