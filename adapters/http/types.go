package httpapi

import "mime/multipart"

type AuthParam struct {
	Authorization string `header:"Authorization" doc:"Bearer access token" example:"Bearer eyJhbGciOi..." required:"true"`
}

type CompanyResponse struct {
	ID                 string `json:"id" doc:"Company identifier" example:"cmp_123"`
	Name               string `json:"name" doc:"Company name" example:"Acme BV"`
	MonthlyTokenBudget int64  `json:"monthly_token_budget" doc:"Monthly token budget" example:"100000"`
}

type CompanyUsageResponse struct {
	CompanyID              string `json:"company_id" doc:"Company identifier" example:"cmp_123"`
	MonthStart             string `json:"month_start" doc:"Month bucket start" format:"date" example:"2026-03-01"`
	BudgetTokens           int64  `json:"budget_tokens" doc:"Configured monthly token budget" example:"100000"`
	InputTokens            int64  `json:"input_tokens" doc:"Consumed input tokens" example:"1234"`
	OutputTokens           int64  `json:"output_tokens" doc:"Consumed output tokens" example:"567"`
	ManualAdjustmentTokens int64  `json:"manual_adjustment_tokens" doc:"Manual adjustment applied to effective usage" example:"-100"`
	EffectiveUsage         int64  `json:"effective_usage" doc:"Effective usage after manual adjustment" example:"1701"`
	RemainingTokens        int64  `json:"remaining_tokens" doc:"Remaining monthly tokens" example:"98299"`
}

type UserResponse struct {
	ID        string  `json:"id" doc:"User identifier" example:"usr_123"`
	Email     string  `json:"email" doc:"Email address" format:"email" example:"user@example.com"`
	Role      string  `json:"role" doc:"Assigned platform role" enum:"admin,user" example:"user"`
	Status    string  `json:"status" doc:"Account lifecycle status" enum:"invited,active,disabled" example:"active"`
	CompanyID *string `json:"company_id,omitempty" doc:"Associated company identifier" example:"cmp_123"`
}

type KnowledgeFileResponse struct {
	ID              string `json:"id" doc:"Knowledge file identifier" example:"file_123"`
	Name            string `json:"name" doc:"Uploaded file name" example:"playbook.pdf"`
	ContentType     string `json:"content_type" doc:"Uploaded MIME type" example:"application/pdf"`
	SizeBytes       int64  `json:"size_bytes" doc:"File size in bytes" example:"2048"`
	ExtractedTokens int    `json:"extracted_tokens" doc:"Estimated extracted token count" example:"512"`
}

type ChatbotResponse struct {
	ID                   string                  `json:"id" doc:"Chatbot identifier" example:"bot_123"`
	Name                 string                  `json:"name" doc:"Chatbot name" example:"Support Bot"`
	Description          string                  `json:"description" doc:"Chatbot description" example:"Customer support assistant"`
	SystemPrompt         string                  `json:"system_prompt,omitempty" doc:"System prompt text" example:"You are a concise support assistant."`
	TotalKnowledgeTokens int                     `json:"total_knowledge_tokens" doc:"Total extracted knowledge token estimate" example:"900"`
	Files                []KnowledgeFileResponse `json:"files,omitempty" doc:"Attached knowledge files"`
	CompanyAssignments   []CompanyResponse       `json:"company_assignments,omitempty" doc:"Companies with access"`
}

type MessageResponse struct {
	ID           string `json:"id" doc:"Message identifier" example:"msg_123"`
	Role         string `json:"role" doc:"Message role" enum:"user,assistant" example:"assistant"`
	Content      string `json:"content" doc:"Rendered message text" example:"How can I help?"`
	Sequence     int    `json:"sequence" doc:"Monotonic message sequence" example:"2"`
	InputTokens  int    `json:"input_tokens,omitempty" doc:"Prompt tokens charged for assistant turns" example:"123"`
	OutputTokens int    `json:"output_tokens,omitempty" doc:"Completion tokens charged for assistant turns" example:"45"`
}

type ConversationResponse struct {
	ID        string            `json:"id" doc:"Conversation identifier" example:"conv_123"`
	ChatbotID string            `json:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
	Status    string            `json:"status" doc:"Conversation status" enum:"active,archived" example:"active"`
	Messages  []MessageResponse `json:"messages" doc:"Conversation messages"`
}

type CreateCompanyBody struct {
	Name               string `json:"name" doc:"Company name" minLength:"1" maxLength:"120" example:"Acme BV"`
	MonthlyTokenBudget int64  `json:"monthly_token_budget" doc:"Monthly token budget" minimum:"0" example:"100000"`
}

type UpdateCompanyBody struct {
	Name               string `json:"name" doc:"Company name" minLength:"1" maxLength:"120" example:"Acme BV"`
	MonthlyTokenBudget int64  `json:"monthly_token_budget" doc:"Monthly token budget" minimum:"0" example:"150000"`
}

type AdjustUsageBody struct {
	Delta int64 `json:"delta" doc:"Positive or negative token adjustment" example:"-100"`
}

type InviteUserBody struct {
	Email     string `json:"email" doc:"Invited user email" format:"email" example:"user@example.com"`
	CompanyID string `json:"company_id" doc:"Company identifier for the invited user" minLength:"1" example:"cmp_123"`
}

type CreateChatbotBody struct {
	Name         string `json:"name" doc:"Chatbot name" minLength:"1" maxLength:"120" example:"Support Bot"`
	Description  string `json:"description" doc:"Chatbot description" maxLength:"500" example:"Customer support assistant"`
	SystemPrompt string `json:"system_prompt" doc:"System prompt prepended to each completion" minLength:"1" example:"You are a concise support assistant."`
}

type UpdateChatbotBody struct {
	Name         string `json:"name" doc:"Chatbot name" minLength:"1" maxLength:"120" example:"Support Bot"`
	Description  string `json:"description" doc:"Chatbot description" maxLength:"500" example:"Customer support assistant"`
	SystemPrompt string `json:"system_prompt" doc:"System prompt prepended to each completion" minLength:"1" example:"You are a concise support assistant."`
}

type SendMessageBody struct {
	Message string `json:"message" doc:"User chat message" minLength:"1" maxLength:"12000" example:"What is our refund policy?"`
}

type UploadKnowledgeFileForm struct {
	RawBody multipart.Form
}
