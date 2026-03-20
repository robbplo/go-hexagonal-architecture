package ports

import (
	"context"
	"time"

	"github.com/linkai/go-chatbot-api/domain/model"
)

type CompanyRepository interface {
	Create(context.Context, model.Company) error
	GetByID(context.Context, string) (model.Company, error)
	List(context.Context) ([]model.Company, error)
	Update(context.Context, model.Company) error
	Delete(context.Context, string) error
}

type CompanyUsageRepository interface {
	GetOrCreate(context.Context, string, time.Time, int64) (model.CompanyMonthUsage, error)
	AddEvent(context.Context, model.TokenUsageEvent, int64) (model.CompanyMonthUsage, error)
	AdjustCurrentMonth(context.Context, string, time.Time, int64, int64) (model.CompanyMonthUsage, error)
	ResetCurrentMonth(context.Context, string, time.Time, int64) (model.CompanyMonthUsage, error)
}

type UserRepository interface {
	Create(context.Context, model.UserProfile) error
	GetByID(context.Context, string) (model.UserProfile, error)
	GetByEmail(context.Context, string) (model.UserProfile, error)
	List(context.Context, model.UserFilter) ([]model.UserProfile, error)
	Update(context.Context, model.UserProfile) error
	Delete(context.Context, string) error
	CountActiveByCompany(context.Context, string) (int, error)
	ListByCompanyAndStatuses(context.Context, string, ...model.UserStatus) ([]model.UserProfile, error)
}

type ChatbotRepository interface {
	Create(context.Context, model.Chatbot) error
	GetByID(context.Context, string) (model.Chatbot, error)
	List(context.Context) ([]model.Chatbot, error)
	ListByCompany(context.Context, string) ([]model.Chatbot, error)
	Update(context.Context, model.Chatbot) error
	Delete(context.Context, string) error
	ListFiles(context.Context, string) ([]model.KnowledgeFile, error)
	CountFiles(context.Context, string) (int, error)
	AddFile(context.Context, model.KnowledgeFile) error
	GetFile(context.Context, string, string) (model.KnowledgeFile, error)
	DeleteFile(context.Context, string, string) (model.KnowledgeFile, error)
	SetTotalKnowledgeTokens(context.Context, string, int) error
}

type GrantRepository interface {
	Grant(context.Context, model.CompanyChatbotGrant) error
	Revoke(context.Context, string, string) error
	CompanyHasAccess(context.Context, string, string) (bool, error)
	ListCompaniesByChatbot(context.Context, string) ([]model.Company, error)
}

type ConversationRepository interface {
	GetActiveByUserAndChatbot(context.Context, string, string) (model.Conversation, []model.Message, error)
	Create(context.Context, model.Conversation) error
	ArchiveActiveAndCreate(context.Context, string, string, time.Time, model.Conversation) error
	AppendMessages(context.Context, ...model.Message) error
	ListMessages(context.Context, string) ([]model.Message, error)
}
