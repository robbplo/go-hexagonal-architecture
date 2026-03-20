package httpapi

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/linkai/go-chatbot-api/app"
	"github.com/linkai/go-chatbot-api/domain/model"
)

type companyService interface {
	Create(context.Context, app.CreateCompanyCommand) (model.Company, error)
	List(context.Context) ([]model.Company, error)
	Update(context.Context, app.UpdateCompanyCommand) (model.Company, error)
	Delete(context.Context, string) error
	GetCurrentUsage(context.Context, string) (model.CompanyMonthUsage, error)
	AdjustCurrentUsage(context.Context, string, int64) (model.CompanyMonthUsage, error)
	ResetCurrentUsage(context.Context, string) (model.CompanyMonthUsage, error)
}

type userService interface {
	Invite(context.Context, app.InviteUserCommand) (model.UserProfile, error)
	List(context.Context, model.UserFilter) ([]model.UserProfile, error)
	Disable(context.Context, string) (model.UserProfile, error)
	Delete(context.Context, string) error
}

type chatbotService interface {
	Create(context.Context, app.CreateChatbotCommand) (model.Chatbot, error)
	List(context.Context) ([]app.AdminChatbotView, error)
	Update(context.Context, app.UpdateChatbotCommand) (model.Chatbot, error)
	Delete(context.Context, string) error
	UploadFile(context.Context, app.UploadKnowledgeFileCommand) (model.KnowledgeFile, error)
	DeleteFile(context.Context, string, string) error
	GrantAccess(context.Context, string, string) error
	RevokeAccess(context.Context, string, string) error
}

type chatService interface {
	GetOrCreateActiveConversation(context.Context, model.UserProfile, string) (model.Conversation, []model.Message, error)
	StartFreshConversation(context.Context, model.UserProfile, string) (model.Conversation, []model.Message, error)
	SendMessage(context.Context, model.UserProfile, string, string) (app.SendMessageResult, error)
}

type catalogService interface {
	ListForUser(context.Context, model.UserProfile) ([]model.Chatbot, error)
}

type Handler struct {
	companies      companyService
	users          userService
	chatbots       chatbotService
	chat           chatService
	catalog        catalogService
	inviteRedirect string
}

func NewHandler(companies companyService, users userService, chatbots chatbotService, chat chatService, catalog catalogService, inviteRedirect string) *Handler {
	return &Handler{
		companies:      companies,
		users:          users,
		chatbots:       chatbots,
		chat:           chat,
		catalog:        catalog,
		inviteRedirect: strings.TrimSpace(inviteRedirect),
	}
}

func (h *Handler) Register(api huma.API, authenticator func(huma.Context, func(huma.Context)), adminOnly func(huma.Context, func(huma.Context))) {
	admin := huma.NewGroup(api, "/admin")
	admin.UseMiddleware(authenticator)
	admin.UseMiddleware(adminOnly)
	registerAdminCompanyRoutes(admin, h)
	registerAdminUserRoutes(admin, h)
	registerAdminChatbotRoutes(admin, h)

	me := huma.NewGroup(api, "/me")
	me.UseMiddleware(authenticator)
	registerSelfRoutes(me, h)
}

func registerAdminCompanyRoutes(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "list-companies",
		Method:      http.MethodGet,
		Path:        "/companies",
		Summary:     "List companies",
		Description: "Returns all companies configured in the platform.",
		Tags:        []string{"Admin Companies"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 500},
	}, func(ctx context.Context, input *struct{ AuthParam }) (*struct {
		Body struct {
			Items []CompanyResponse `json:"items" doc:"Companies configured in the platform"`
		}
	}, error) {
		companies, err := h.companies.List(ctx)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct {
			Body struct {
				Items []CompanyResponse `json:"items" doc:"Companies configured in the platform"`
			}
		}{}
		resp.Body.Items = make([]CompanyResponse, 0, len(companies))
		for _, company := range companies {
			resp.Body.Items = append(resp.Body.Items, toCompanyResponse(company))
		}
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-company",
		Method:        http.MethodPost,
		Path:          "/companies",
		Summary:       "Create a company",
		Description:   "Creates a company with a monthly token budget.",
		Tags:          []string{"Admin Companies"},
		DefaultStatus: http.StatusCreated,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 409, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		Body CreateCompanyBody
	}) (*struct {
		Body CompanyResponse
	}, error) {
		company, err := h.companies.Create(ctx, app.CreateCompanyCommand{
			Name:               input.Body.Name,
			MonthlyTokenBudget: input.Body.MonthlyTokenBudget,
		})
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body CompanyResponse }{}
		resp.Body = toCompanyResponse(company)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-company",
		Method:      http.MethodPatch,
		Path:        "/companies/{company_id}",
		Summary:     "Update a company",
		Description: "Updates company name and monthly token budget.",
		Tags:        []string{"Admin Companies"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 409, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
		Body      UpdateCompanyBody
	}) (*struct {
		Body CompanyResponse
	}, error) {
		company, err := h.companies.Update(ctx, app.UpdateCompanyCommand{
			CompanyID:          input.CompanyID,
			Name:               input.Body.Name,
			MonthlyTokenBudget: input.Body.MonthlyTokenBudget,
		})
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body CompanyResponse }{}
		resp.Body = toCompanyResponse(company)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-company",
		Method:        http.MethodDelete,
		Path:          "/companies/{company_id}",
		Summary:       "Delete a company",
		Description:   "Deletes a company once active users are removed.",
		Tags:          []string{"Admin Companies"},
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 409, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
	}) (*struct{}, error) {
		if err := h.companies.Delete(ctx, input.CompanyID); err != nil {
			return nil, mapDomainError(err)
		}
		return &struct{}{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-company-usage",
		Method:      http.MethodGet,
		Path:        "/companies/{company_id}/usage/current",
		Summary:     "Get current monthly token usage",
		Description: "Returns aggregated token usage for the current calendar month.",
		Tags:        []string{"Admin Companies"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
	}) (*struct {
		Body CompanyUsageResponse
	}, error) {
		usage, err := h.companies.GetCurrentUsage(ctx, input.CompanyID)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body CompanyUsageResponse }{}
		resp.Body = toUsageResponse(usage)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "adjust-company-usage",
		Method:      http.MethodPost,
		Path:        "/companies/{company_id}/usage/adjust",
		Summary:     "Adjust current monthly usage",
		Description: "Applies a positive or negative manual token adjustment for the current month.",
		Tags:        []string{"Admin Companies"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
		Body      AdjustUsageBody
	}) (*struct {
		Body CompanyUsageResponse
	}, error) {
		usage, err := h.companies.AdjustCurrentUsage(ctx, input.CompanyID, input.Body.Delta)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body CompanyUsageResponse }{}
		resp.Body = toUsageResponse(usage)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "reset-company-usage",
		Method:      http.MethodPost,
		Path:        "/companies/{company_id}/usage/reset",
		Summary:     "Reset current monthly usage",
		Description: "Resets effective usage for the current month to zero using a manual adjustment.",
		Tags:        []string{"Admin Companies"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
	}) (*struct {
		Body CompanyUsageResponse
	}, error) {
		usage, err := h.companies.ResetCurrentUsage(ctx, input.CompanyID)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body CompanyUsageResponse }{}
		resp.Body = toUsageResponse(usage)
		return resp, nil
	})
}

func registerAdminUserRoutes(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List users",
		Description: "Returns invited, active, and disabled users.",
		Tags:        []string{"Admin Users"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string   `query:"company_id" doc:"Optional company filter" example:"cmp_123"`
		Status    []string `query:"status" doc:"Optional repeated status filter" enum:"invited,active,disabled"`
	}) (*struct {
		Body struct {
			Items []UserResponse `json:"items" doc:"Users visible to the administrator"`
		}
	}, error) {
		filter := model.UserFilter{}
		if strings.TrimSpace(input.CompanyID) != "" {
			filter.CompanyID = &input.CompanyID
		}
		for _, status := range input.Status {
			filter.Statuses = append(filter.Statuses, model.UserStatus(status))
		}
		users, err := h.users.List(ctx, filter)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct {
			Body struct {
				Items []UserResponse `json:"items" doc:"Users visible to the administrator"`
			}
		}{}
		resp.Body.Items = make([]UserResponse, 0, len(users))
		for _, user := range users {
			resp.Body.Items = append(resp.Body.Items, toUserResponse(user))
		}
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "invite-user",
		Method:        http.MethodPost,
		Path:          "/users/invitations",
		Summary:       "Invite a user",
		Description:   "Creates an invited account bound to a company and triggers a Supabase invite email.",
		Tags:          []string{"Admin Users"},
		DefaultStatus: http.StatusCreated,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 409, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		Body InviteUserBody
	}) (*struct {
		Body UserResponse
	}, error) {
		user, err := h.users.Invite(ctx, app.InviteUserCommand{
			Email:       input.Body.Email,
			CompanyID:   input.Body.CompanyID,
			RedirectURL: h.inviteRedirect,
		})
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body UserResponse }{}
		resp.Body = toUserResponse(user)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "disable-user",
		Method:      http.MethodPost,
		Path:        "/users/{user_id}/disable",
		Summary:     "Disable a user",
		Description: "Disables a user in both Supabase Auth and the local profile store.",
		Tags:        []string{"Admin Users"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 409, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		UserID string `path:"user_id" doc:"User identifier" example:"usr_123"`
	}) (*struct {
		Body UserResponse
	}, error) {
		user, err := h.users.Disable(ctx, input.UserID)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body UserResponse }{}
		resp.Body = toUserResponse(user)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-user",
		Method:        http.MethodDelete,
		Path:          "/users/{user_id}",
		Summary:       "Delete a user",
		Description:   "Deletes a user profile and its Supabase auth account.",
		Tags:          []string{"Admin Users"},
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 409, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		UserID string `path:"user_id" doc:"User identifier" example:"usr_123"`
	}) (*struct{}, error) {
		if err := h.users.Delete(ctx, input.UserID); err != nil {
			return nil, mapDomainError(err)
		}
		return &struct{}{}, nil
	})
}

func registerAdminChatbotRoutes(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "list-chatbots",
		Method:      http.MethodGet,
		Path:        "/chatbots",
		Summary:     "List chatbots",
		Description: "Returns all chatbots, attached knowledge files, and company assignments.",
		Tags:        []string{"Admin Chatbots"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 500},
	}, func(ctx context.Context, input *struct{ AuthParam }) (*struct {
		Body struct {
			Items []ChatbotResponse `json:"items" doc:"Configured chatbots"`
		}
	}, error) {
		views, err := h.chatbots.List(ctx)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct {
			Body struct {
				Items []ChatbotResponse `json:"items" doc:"Configured chatbots"`
			}
		}{}
		resp.Body.Items = make([]ChatbotResponse, 0, len(views))
		for _, view := range views {
			resp.Body.Items = append(resp.Body.Items, toAdminChatbotResponse(view))
		}
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-chatbot",
		Method:        http.MethodPost,
		Path:          "/chatbots",
		Summary:       "Create a chatbot",
		Description:   "Creates a chatbot with its name, description, and system prompt.",
		Tags:          []string{"Admin Chatbots"},
		DefaultStatus: http.StatusCreated,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 409, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		Body CreateChatbotBody
	}) (*struct {
		Body ChatbotResponse
	}, error) {
		chatbot, err := h.chatbots.Create(ctx, app.CreateChatbotCommand{
			Name:         input.Body.Name,
			Description:  input.Body.Description,
			SystemPrompt: input.Body.SystemPrompt,
		})
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body ChatbotResponse }{}
		resp.Body = toChatbotResponse(chatbot)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-chatbot",
		Method:      http.MethodPatch,
		Path:        "/chatbots/{chatbot_id}",
		Summary:     "Update a chatbot",
		Description: "Updates chatbot metadata and system prompt.",
		Tags:        []string{"Admin Chatbots"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 409, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
		Body      UpdateChatbotBody
	}) (*struct {
		Body ChatbotResponse
	}, error) {
		chatbot, err := h.chatbots.Update(ctx, app.UpdateChatbotCommand{
			ChatbotID:    input.ChatbotID,
			Name:         input.Body.Name,
			Description:  input.Body.Description,
			SystemPrompt: input.Body.SystemPrompt,
		})
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body ChatbotResponse }{}
		resp.Body = toChatbotResponse(chatbot)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-chatbot",
		Method:        http.MethodDelete,
		Path:          "/chatbots/{chatbot_id}",
		Summary:       "Delete a chatbot",
		Description:   "Deletes a chatbot and cascades its files, assignments, and conversations.",
		Tags:          []string{"Admin Chatbots"},
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
	}) (*struct{}, error) {
		if err := h.chatbots.Delete(ctx, input.ChatbotID); err != nil {
			return nil, mapDomainError(err)
		}
		return &struct{}{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "upload-chatbot-file",
		Method:        http.MethodPost,
		Path:          "/chatbots/{chatbot_id}/files",
		Summary:       "Upload a chatbot knowledge file",
		Description:   "Accepts multipart form uploads and attaches a supported knowledge file to the chatbot.",
		Tags:          []string{"Admin Chatbots"},
		DefaultStatus: http.StatusCreated,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 409, 422, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
		RawBody   multipart.Form
	}) (*struct {
		Body KnowledgeFileResponse
	}, error) {
		file, err := singleUpload(input.RawBody, "file")
		if err != nil {
			return nil, mapDomainError(err)
		}
		knowledgeFile, err := h.chatbots.UploadFile(ctx, app.UploadKnowledgeFileCommand{
			ChatbotID:   input.ChatbotID,
			FileName:    file.name,
			ContentType: file.contentType,
			SizeBytes:   int64(len(file.data)),
			Data:        file.data,
		})
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body KnowledgeFileResponse }{}
		resp.Body = toKnowledgeFileResponse(knowledgeFile)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-chatbot-file",
		Method:        http.MethodDelete,
		Path:          "/chatbots/{chatbot_id}/files/{file_id}",
		Summary:       "Delete a chatbot knowledge file",
		Description:   "Removes a chatbot knowledge file from storage and the database.",
		Tags:          []string{"Admin Chatbots"},
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
		FileID    string `path:"file_id" doc:"Knowledge file identifier" example:"file_123"`
	}) (*struct{}, error) {
		if err := h.chatbots.DeleteFile(ctx, input.ChatbotID, input.FileID); err != nil {
			return nil, mapDomainError(err)
		}
		return &struct{}{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "grant-chatbot-access",
		Method:        http.MethodPut,
		Path:          "/companies/{company_id}/chatbots/{chatbot_id}",
		Summary:       "Grant company access to chatbot",
		Description:   "Allows all users in a company to access a chatbot.",
		Tags:          []string{"Admin Chatbots"},
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
	}) (*struct{}, error) {
		if err := h.chatbots.GrantAccess(ctx, input.CompanyID, input.ChatbotID); err != nil {
			return nil, mapDomainError(err)
		}
		return &struct{}{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "revoke-chatbot-access",
		Method:        http.MethodDelete,
		Path:          "/companies/{company_id}/chatbots/{chatbot_id}",
		Summary:       "Revoke company access to chatbot",
		Description:   "Removes company access to a chatbot.",
		Tags:          []string{"Admin Chatbots"},
		DefaultStatus: http.StatusNoContent,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		CompanyID string `path:"company_id" doc:"Company identifier" example:"cmp_123"`
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
	}) (*struct{}, error) {
		if err := h.chatbots.RevokeAccess(ctx, input.CompanyID, input.ChatbotID); err != nil {
			return nil, mapDomainError(err)
		}
		return &struct{}{}, nil
	})
}

func registerSelfRoutes(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "",
		Summary:     "Get current user profile",
		Description: "Returns the authenticated user profile loaded by middleware.",
		Tags:        []string{"Me"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 500},
	}, func(ctx context.Context, input *struct{ AuthParam }) (*struct {
		Body UserResponse
	}, error) {
		profile, ok := UserProfileFromContext(ctx)
		if !ok {
			return nil, huma.Error500InternalServerError("missing user profile")
		}
		resp := &struct{ Body UserResponse }{}
		resp.Body = toUserResponse(profile)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-my-chatbots",
		Method:      http.MethodGet,
		Path:        "/chatbots",
		Summary:     "List available chatbots",
		Description: "Returns the chatbots assigned to the authenticated user's company.",
		Tags:        []string{"Me"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 500},
	}, func(ctx context.Context, input *struct{ AuthParam }) (*struct {
		Body struct {
			Items []ChatbotResponse `json:"items" doc:"Chatbots available to the authenticated user"`
		}
	}, error) {
		profile, ok := UserProfileFromContext(ctx)
		if !ok {
			return nil, huma.Error500InternalServerError("missing user profile")
		}
		chatbots, err := h.catalog.ListForUser(ctx, profile)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct {
			Body struct {
				Items []ChatbotResponse `json:"items" doc:"Chatbots available to the authenticated user"`
			}
		}{}
		resp.Body.Items = make([]ChatbotResponse, 0, len(chatbots))
		for _, chatbot := range chatbots {
			resp.Body.Items = append(resp.Body.Items, toChatbotResponse(chatbot))
		}
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-active-conversation",
		Method:      http.MethodGet,
		Path:        "/chatbots/{chatbot_id}/conversation",
		Summary:     "Get active conversation",
		Description: "Returns the active conversation for the authenticated user and chatbot, creating one if missing.",
		Tags:        []string{"Me"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
	}) (*struct {
		Body ConversationResponse
	}, error) {
		profile, ok := UserProfileFromContext(ctx)
		if !ok {
			return nil, huma.Error500InternalServerError("missing user profile")
		}
		conversation, messages, err := h.chat.GetOrCreateActiveConversation(ctx, profile, input.ChatbotID)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body ConversationResponse }{}
		resp.Body = toConversationResponse(conversation, messages)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "start-fresh-conversation",
		Method:        http.MethodPost,
		Path:          "/chatbots/{chatbot_id}/conversations",
		Summary:       "Start a fresh conversation",
		Description:   "Archives the current active conversation and creates a new one.",
		Tags:          []string{"Me"},
		DefaultStatus: http.StatusCreated,
		Security:      []map[string][]string{{"bearer": {}}},
		Errors:        []int{401, 403, 404, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
	}) (*struct {
		Body ConversationResponse
	}, error) {
		profile, ok := UserProfileFromContext(ctx)
		if !ok {
			return nil, huma.Error500InternalServerError("missing user profile")
		}
		conversation, messages, err := h.chat.StartFreshConversation(ctx, profile, input.ChatbotID)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body ConversationResponse }{}
		resp.Body = toConversationResponse(conversation, messages)
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "send-message",
		Method:      http.MethodPost,
		Path:        "/chatbots/{chatbot_id}/messages",
		Summary:     "Send a chat message",
		Description: "Sends a message to the chatbot, persists the assistant response, and records token usage.",
		Tags:        []string{"Me"},
		Security:    []map[string][]string{{"bearer": {}}},
		Errors:      []int{401, 403, 404, 422, 429, 500},
	}, func(ctx context.Context, input *struct {
		AuthParam
		ChatbotID string `path:"chatbot_id" doc:"Chatbot identifier" example:"bot_123"`
		Body      SendMessageBody
	}) (*struct {
		Body ConversationResponse
	}, error) {
		profile, ok := UserProfileFromContext(ctx)
		if !ok {
			return nil, huma.Error500InternalServerError("missing user profile")
		}
		result, err := h.chat.SendMessage(ctx, profile, input.ChatbotID, input.Body.Message)
		if err != nil {
			return nil, mapDomainError(err)
		}
		resp := &struct{ Body ConversationResponse }{}
		resp.Body = toConversationResponse(result.Conversation, result.Messages)
		return resp, nil
	})
}

type uploadedFile struct {
	name        string
	contentType string
	data        []byte
}

func singleUpload(form multipart.Form, field string) (uploadedFile, error) {
	files := form.File[field]
	if len(files) != 1 {
		return uploadedFile{}, fmt.Errorf("upload file: expected exactly one file")
	}
	handle, err := files[0].Open()
	if err != nil {
		return uploadedFile{}, err
	}
	defer handle.Close()
	data, err := io.ReadAll(handle)
	if err != nil {
		return uploadedFile{}, err
	}
	contentType := files[0].Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return uploadedFile{
		name:        files[0].Filename,
		contentType: contentType,
		data:        data,
	}, nil
}

func toCompanyResponse(company model.Company) CompanyResponse {
	return CompanyResponse{
		ID:                 company.ID,
		Name:               company.Name,
		MonthlyTokenBudget: company.MonthlyTokenBudget,
	}
}

func toUsageResponse(usage model.CompanyMonthUsage) CompanyUsageResponse {
	return CompanyUsageResponse{
		CompanyID:              usage.CompanyID,
		MonthStart:             usage.MonthStart.Format("2006-01-02"),
		BudgetTokens:           usage.BudgetTokens,
		InputTokens:            usage.InputTokens,
		OutputTokens:           usage.OutputTokens,
		ManualAdjustmentTokens: usage.ManualAdjustmentTokens,
		EffectiveUsage:         usage.EffectiveUsage(),
		RemainingTokens:        usage.RemainingTokens(),
	}
}

func toUserResponse(user model.UserProfile) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CompanyID: user.CompanyID,
	}
}

func toKnowledgeFileResponse(file model.KnowledgeFile) KnowledgeFileResponse {
	return KnowledgeFileResponse{
		ID:              file.ID,
		Name:            file.Name,
		ContentType:     file.ContentType,
		SizeBytes:       file.SizeBytes,
		ExtractedTokens: file.ExtractedTokens,
	}
}

func toChatbotResponse(chatbot model.Chatbot) ChatbotResponse {
	return ChatbotResponse{
		ID:                   chatbot.ID,
		Name:                 chatbot.Name,
		Description:          chatbot.Description,
		SystemPrompt:         chatbot.SystemPrompt,
		TotalKnowledgeTokens: chatbot.TotalKnowledgeTokens,
	}
}

func toAdminChatbotResponse(view app.AdminChatbotView) ChatbotResponse {
	response := toChatbotResponse(view.Chatbot)
	response.Files = make([]KnowledgeFileResponse, 0, len(view.Files))
	for _, file := range view.Files {
		response.Files = append(response.Files, toKnowledgeFileResponse(file))
	}
	response.CompanyAssignments = make([]CompanyResponse, 0, len(view.Companies))
	for _, company := range view.Companies {
		response.CompanyAssignments = append(response.CompanyAssignments, toCompanyResponse(company))
	}
	return response
}

func toConversationResponse(conversation model.Conversation, messages []model.Message) ConversationResponse {
	response := ConversationResponse{
		ID:        conversation.ID,
		ChatbotID: conversation.ChatbotID,
		Status:    string(conversation.Status),
		Messages:  make([]MessageResponse, 0, len(messages)),
	}
	for _, message := range messages {
		response.Messages = append(response.Messages, MessageResponse{
			ID:           message.ID,
			Role:         string(message.Role),
			Content:      message.Content,
			Sequence:     message.Sequence,
			InputTokens:  message.InputTokens,
			OutputTokens: message.OutputTokens,
		})
	}
	return response
}
