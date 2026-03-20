package httpapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/linkai/go-chatbot-api/app"
	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMyChatbots_Success(t *testing.T) {
	_, api := humatest.New(t, huma.DefaultConfig("Test", "1.0.0"))

	authenticator := stubAuthenticator{
		user: ports.AuthenticatedUser{
			ID:    "usr_1",
			Email: "user@example.com",
		},
	}
	sessions := stubSessionResolver{
		profile: model.UserProfile{
			ID:        "usr_1",
			Email:     "user@example.com",
			Role:      model.RoleUser,
			Status:    model.UserStatusActive,
			CompanyID: ptr("cmp_1"),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	handler := NewHandler(
		stubCompanyService{},
		stubUserService{},
		stubChatbotService{},
		stubChatService{},
		stubCatalogService{
			chatbots: []model.Chatbot{
				{
					ID:           "bot_1",
					Name:         "Support Bot",
					Description:  "Support helper",
					SystemPrompt: "You are helpful.",
				},
			},
		},
		"https://app.example.com/invite",
	)

	handler.Register(api, AuthMiddleware(api, authenticator, sessions), AdminOnlyMiddleware(api))

	resp := api.Get("/me/chatbots", "Authorization: Bearer token")

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Support Bot")
}

func TestCreateCompany_ForbidsNonAdmin(t *testing.T) {
	_, api := humatest.New(t, huma.DefaultConfig("Test", "1.0.0"))

	authenticator := stubAuthenticator{
		user: ports.AuthenticatedUser{
			ID:    "usr_1",
			Email: "user@example.com",
		},
	}
	sessions := stubSessionResolver{
		profile: model.UserProfile{
			ID:        "usr_1",
			Email:     "user@example.com",
			Role:      model.RoleUser,
			Status:    model.UserStatusActive,
			CompanyID: ptr("cmp_1"),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	handler := NewHandler(
		stubCompanyService{},
		stubUserService{},
		stubChatbotService{},
		stubChatService{},
		stubCatalogService{},
		"https://app.example.com/invite",
	)

	handler.Register(api, AuthMiddleware(api, authenticator, sessions), AdminOnlyMiddleware(api))

	resp := api.Post("/admin/companies", "Authorization: Bearer token", map[string]any{
		"name":                 "Acme",
		"monthly_token_budget": 100,
	})

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestMapDomainError_NotFound(t *testing.T) {
	err := mapDomainError(domainerrors.ErrNotFound)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

type stubAuthenticator struct {
	user ports.AuthenticatedUser
	err  error
}

func (s stubAuthenticator) Authenticate(_ context.Context, _ string) (ports.AuthenticatedUser, error) {
	return s.user, s.err
}

type stubSessionResolver struct {
	profile model.UserProfile
	err     error
}

func (s stubSessionResolver) ResolveAuthenticatedProfile(_ context.Context, _ ports.AuthenticatedUser) (model.UserProfile, error) {
	return s.profile, s.err
}

type stubCompanyService struct{}

func (stubCompanyService) Create(context.Context, app.CreateCompanyCommand) (model.Company, error) {
	return model.Company{}, errors.New("not implemented")
}

func (stubCompanyService) List(context.Context) ([]model.Company, error) {
	return nil, nil
}

func (stubCompanyService) Update(context.Context, app.UpdateCompanyCommand) (model.Company, error) {
	return model.Company{}, errors.New("not implemented")
}

func (stubCompanyService) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func (stubCompanyService) GetCurrentUsage(context.Context, string) (model.CompanyMonthUsage, error) {
	return model.CompanyMonthUsage{}, errors.New("not implemented")
}

func (stubCompanyService) AdjustCurrentUsage(context.Context, string, int64) (model.CompanyMonthUsage, error) {
	return model.CompanyMonthUsage{}, errors.New("not implemented")
}

func (stubCompanyService) ResetCurrentUsage(context.Context, string) (model.CompanyMonthUsage, error) {
	return model.CompanyMonthUsage{}, errors.New("not implemented")
}

type stubUserService struct{}

func (stubUserService) Invite(context.Context, app.InviteUserCommand) (model.UserProfile, error) {
	return model.UserProfile{}, errors.New("not implemented")
}

func (stubUserService) List(context.Context, model.UserFilter) ([]model.UserProfile, error) {
	return nil, nil
}

func (stubUserService) Disable(context.Context, string) (model.UserProfile, error) {
	return model.UserProfile{}, errors.New("not implemented")
}

func (stubUserService) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

type stubChatbotService struct{}

func (stubChatbotService) Create(context.Context, app.CreateChatbotCommand) (model.Chatbot, error) {
	return model.Chatbot{}, errors.New("not implemented")
}

func (stubChatbotService) List(context.Context) ([]app.AdminChatbotView, error) {
	return nil, nil
}

func (stubChatbotService) Update(context.Context, app.UpdateChatbotCommand) (model.Chatbot, error) {
	return model.Chatbot{}, errors.New("not implemented")
}

func (stubChatbotService) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func (stubChatbotService) UploadFile(context.Context, app.UploadKnowledgeFileCommand) (model.KnowledgeFile, error) {
	return model.KnowledgeFile{}, errors.New("not implemented")
}

func (stubChatbotService) DeleteFile(context.Context, string, string) error {
	return errors.New("not implemented")
}

func (stubChatbotService) GrantAccess(context.Context, string, string) error {
	return errors.New("not implemented")
}

func (stubChatbotService) RevokeAccess(context.Context, string, string) error {
	return errors.New("not implemented")
}

type stubChatService struct{}

func (stubChatService) GetOrCreateActiveConversation(context.Context, model.UserProfile, string) (model.Conversation, []model.Message, error) {
	return model.Conversation{}, nil, errors.New("not implemented")
}

func (stubChatService) StartFreshConversation(context.Context, model.UserProfile, string) (model.Conversation, []model.Message, error) {
	return model.Conversation{}, nil, errors.New("not implemented")
}

func (stubChatService) SendMessage(context.Context, model.UserProfile, string, string) (app.SendMessageResult, error) {
	return app.SendMessageResult{}, errors.New("not implemented")
}

type stubCatalogService struct {
	chatbots []model.Chatbot
}

func (s stubCatalogService) ListForUser(context.Context, model.UserProfile) ([]model.Chatbot, error) {
	return s.chatbots, nil
}

func ptr(value string) *string {
	return &value
}
