package ai

import "time"

// AIReport adalah representasi laporan AI dari database.
type AIReport struct {
	ID          string    `json:"id"`
	EmployeeID  string    `json:"employee_id"`
	GeneratedBy string    `json:"generated_by"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Prompt      string    `json:"prompt,omitempty"`
	Response    string    `json:"response"`
	ModelUsed   string    `json:"model_used,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AnalyzeRequest adalah body request untuk generate laporan AI.
type AnalyzeRequest struct {
	PeriodStart string `json:"period_start"` // Format: YYYY-MM-DD
	PeriodEnd   string `json:"period_end"`   // Format: YYYY-MM-DD
	CustomPrompt string `json:"custom_prompt,omitempty"`
}

// ChatMessage adalah satu pesan dalam conversation LLM.
type ChatMessage struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// ChatRequest adalah request body ke OpenAI-compatible API.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse adalah response dari OpenAI-compatible API.
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}
