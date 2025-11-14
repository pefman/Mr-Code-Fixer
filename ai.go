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

type Fix struct {
	FileChanges []FileChange
	Explanation string
}

// Groq Client (Free tier available: https://console.groq.com)
type GroqClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewGroqClient(apiKey, model string) *GroqClient {
	return &GroqClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.groq.com/openai/v1",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type GroqRequest struct {
	Model    string          `json:"model"`
	Messages []GroqMessage   `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message GroqMessage `json:"message"`
	} `json:"choices"`
}

func (g *GroqClient) AnalyzeAndFix(issue Issue, context *RepoContext) (*Fix, error) {
	prompt := g.buildPrompt(issue, context)

	reqBody := GroqRequest{
		Model: g.model,
		Messages: []GroqMessage{
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

	req, err := http.NewRequest("POST", g.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, err
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return g.parseFix(groqResp.Choices[0].Message.Content)
}

func (g *GroqClient) buildPrompt(issue Issue, context *RepoContext) string {
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
  "explanation": "Brief explanation of what the fix does",
  "files": [
    {
      "path": "relative/path/to/file.ext",
      "content": "complete file content with the fix applied"
    }
  ]
}

Rules:
- Provide COMPLETE file content, not diffs or patches
- Only include files that need to be modified or created
- Keep explanations concise but clear
- Ensure the fix actually addresses the issue
- If you need to create a new file, include its full content
- Return valid JSON only, no markdown code blocks

Now provide the fix:`)

	return prompt.String()
}

func (g *GroqClient) parseFix(response string) (*Fix, error) {
	// Clean up markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result struct {
		Explanation string `json:"explanation"`
		Files       []struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		} `json:"files"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\nResponse: %s", err, response)
	}

	fix := &Fix{
		Explanation: result.Explanation,
		FileChanges: make([]FileChange, len(result.Files)),
	}

	for i, file := range result.Files {
		fix.FileChanges[i] = FileChange{
			FilePath: file.Path,
			Content:  file.Content,
		}
	}

	return fix, nil
}

// Ollama Client (Free local AI: https://ollama.com)
type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 300 * time.Second}, // Longer timeout for local models
	}
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
	g := &GroqClient{}
	return g.buildPrompt(issue, context)
}

func (o *OllamaClient) parseFix(response string) (*Fix, error) {
	// Same parsing logic as Groq
	g := &GroqClient{}
	return g.parseFix(response)
}
