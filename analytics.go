package main

import (
	"fmt"
	"sync"
	"time"
)

// SessionAnalytics tracks API calls and costs during a session
type SessionAnalytics struct {
	StartTime      time.Time
	APICallCount   int
	EstimatedCost  float64
	IssuesHandled  int
	PRsCreated     int
	QuestionsAsked int
	mutex          sync.Mutex
}

// Cost estimates per provider (approximate, in SEK/kr)
var costPerCall = map[string]float64{
	"chatgpt": 0.02,   // ~0.02 kr per request (gpt-4)
	"openai":  0.02,
	"grok":    0.01,   // ~0.01 kr per request
	"xai":     0.01,
	"ollama":  0.0,    // Free (local)
}

func NewSessionAnalytics() *SessionAnalytics {
	return &SessionAnalytics{
		StartTime: time.Now(),
	}
}

func (s *SessionAnalytics) RecordAPICall(service string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.APICallCount++
	if cost, ok := costPerCall[service]; ok {
		s.EstimatedCost += cost
	}
}

func (s *SessionAnalytics) RecordIssueHandled() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.IssuesHandled++
}

func (s *SessionAnalytics) RecordPRCreated() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.PRsCreated++
}

func (s *SessionAnalytics) RecordQuestionAsked() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.QuestionsAsked++
}

func (s *SessionAnalytics) EstimateCostForIssues(count int, service string) float64 {
	cost, ok := costPerCall[service]
	if !ok {
		cost = 0.001 // Default estimate
	}
	// Each issue typically requires 1-2 API calls
	return float64(count) * cost * 1.5
}

func (s *SessionAnalytics) PrintSummary() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	duration := time.Since(s.StartTime)
	
	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    📊 Session Summary                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n⏱️  Duration: %s\n", duration.Round(time.Second))
	fmt.Printf("📞 API Calls: %d\n", s.APICallCount)
	fmt.Printf("🐛 Issues Handled: %d\n", s.IssuesHandled)
	fmt.Printf("🔧 Pull Requests Created: %d\n", s.PRsCreated)
	fmt.Printf("❓ Questions Asked: %d\n", s.QuestionsAsked)
	
	if s.EstimatedCost > 0 {
		fmt.Printf("💰 Estimated Cost: %.4f kr\n", s.EstimatedCost)
	} else {
		fmt.Printf("💰 Cost: Free (local model)\n")
	}
	fmt.Println()
}

func (s *SessionAnalytics) PrintCostEstimate(issueCount int, service string) {
	cost := s.EstimateCostForIssues(issueCount, service)
	
	if cost > 0 {
		fmt.Printf("\n💰 Estimated cost for %d issue(s): %.4f kr\n", issueCount, cost)
		if cost > 1.0 {
			fmt.Println("⚠️  This will cost more than 1 kr - proceed with caution")
		}
	}
}
