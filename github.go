package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	HTMLURL string `json:"html_url"`
}

type GitHubClient struct {
	token     string
	owner     string
	repo      string
	baseURL   string
	client    *http.Client
}

func NewGitHubClient(token, owner, repo string) *GitHubClient {
	return &GitHubClient{
		token:   token,
		owner:   owner,
		repo:    repo,
		baseURL: "https://api.github.com",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GitHubClient) GetOpenIssues(maxIssues int) ([]Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=open&per_page=%d", 
		g.baseURL, g.owner, g.repo, maxIssues)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	// Filter out pull requests (they appear in issues endpoint too)
	var filteredIssues []Issue
	for _, issue := range issues {
		// Pull requests have a "pull_request" field, but we're not including it in our struct
		// so we can't filter them out easily. For now, we'll just return all issues.
		filteredIssues = append(filteredIssues, issue)
	}

	return filteredIssues, nil
}

func (g *GitHubClient) GetIssue(number int) (*Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", 
		g.baseURL, g.owner, g.repo, number)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

type CreatePRRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type PullRequest struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
}

func (g *GitHubClient) CreatePullRequest(title, body, head, base string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", 
		g.baseURL, g.owner, g.repo)
	
	prReq := CreatePRRequest{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
	}

	jsonData, err := json.Marshal(prReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error creating PR: %s - %s", resp.Status, string(body))
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", err
	}

	return pr.HTMLURL, nil
}

func (g *GitHubClient) AddIssueComment(issueNumber int, comment string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", 
		g.baseURL, g.owner, g.repo, issueNumber)
	
	reqBody := map[string]string{
		"body": comment,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error adding comment: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (g *GitHubClient) CloseIssue(issueNumber int) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", 
		g.baseURL, g.owner, g.repo, issueNumber)
	
	reqBody := map[string]string{
		"state": "closed",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error closing issue: %s - %s", resp.Status, string(body))
	}

	return nil
}
