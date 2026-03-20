package postgres

import "time"

type Company struct {
	ID                 string
	Name               string
	MonthlyTokenBudget int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type UserProfile struct {
	ID          string
	Email       string
	Role        string
	Status      string
	CompanyID   *string
	InvitedAt   *time.Time
	ActivatedAt *time.Time
	DisabledAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Chatbot struct {
	ID                   string
	Name                 string
	Description          string
	SystemPrompt         string
	TotalKnowledgeTokens int32
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ChatbotFile struct {
	ID              string
	ChatbotID       string
	Name            string
	ContentType     string
	SizeBytes       int64
	Kind            string
	StoragePath     string
	ExtractedText   string
	ExtractedTokens int32
	CreatedAt       time.Time
}

type Conversation struct {
	ID         string
	UserID     string
	ChatbotID  string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	Sequence       int32
	InputTokens    int32
	OutputTokens   int32
	CreatedAt      time.Time
}

type CompanyMonthUsage struct {
	CompanyID              string
	MonthStart             time.Time
	BudgetTokens           int64
	InputTokens            int64
	OutputTokens           int64
	ManualAdjustmentTokens int64
	UpdatedAt              time.Time
}
