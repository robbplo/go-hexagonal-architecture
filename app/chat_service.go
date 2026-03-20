package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type ChatService struct {
	companies        ports.CompanyRepository
	chatbots         ports.ChatbotRepository
	grants           ports.GrantRepository
	conversations    ports.ConversationRepository
	usage            ports.CompanyUsageRepository
	ai               ports.AIChatClient
	tokenCounter     ports.TokenCounter
	clock            ports.Clock
	ids              ports.IDGenerator
	logger           ports.Logger
	modelName        string
	historyMaxTokens int
}

type SendMessageResult struct {
	Conversation model.Conversation
	Messages     []model.Message
}

func NewChatService(
	companies ports.CompanyRepository,
	chatbots ports.ChatbotRepository,
	grants ports.GrantRepository,
	conversations ports.ConversationRepository,
	usage ports.CompanyUsageRepository,
	ai ports.AIChatClient,
	tokenCounter ports.TokenCounter,
	clock ports.Clock,
	ids ports.IDGenerator,
	logger ports.Logger,
	modelName string,
	historyMaxTokens int,
) *ChatService {
	if companies == nil {
		panic("ChatService: companies is required")
	}
	if chatbots == nil {
		panic("ChatService: chatbots is required")
	}
	if grants == nil {
		panic("ChatService: grants is required")
	}
	if conversations == nil {
		panic("ChatService: conversations is required")
	}
	if usage == nil {
		panic("ChatService: usage is required")
	}
	if ai == nil {
		panic("ChatService: ai is required")
	}
	if tokenCounter == nil {
		panic("ChatService: tokenCounter is required")
	}
	if clock == nil {
		panic("ChatService: clock is required")
	}
	if ids == nil {
		panic("ChatService: ids is required")
	}
	if logger == nil {
		panic("ChatService: logger is required")
	}
	return &ChatService{
		companies:        companies,
		chatbots:         chatbots,
		grants:           grants,
		conversations:    conversations,
		usage:            usage,
		ai:               ai,
		tokenCounter:     tokenCounter,
		clock:            clock,
		ids:              ids,
		logger:           logger,
		modelName:        modelName,
		historyMaxTokens: historyMaxTokens,
	}
}

func (s *ChatService) GetOrCreateActiveConversation(ctx context.Context, user model.UserProfile, chatbotID string) (model.Conversation, []model.Message, error) {
	chatbotID = strings.TrimSpace(chatbotID)
	if err := user.Validate(); err != nil {
		return model.Conversation{}, nil, fmt.Errorf("get conversation: %w", err)
	}
	if err := s.ensureUserAccess(ctx, user, chatbotID); err != nil {
		return model.Conversation{}, nil, fmt.Errorf("get conversation: %w", err)
	}
	conversation, messages, err := s.conversations.GetActiveByUserAndChatbot(ctx, user.ID, chatbotID)
	if err == nil {
		return conversation, messages, nil
	}
	if !errors.Is(err, domainerrors.ErrNotFound) {
		return model.Conversation{}, nil, fmt.Errorf("get conversation: %w", err)
	}
	conversation, err = model.NewConversation(s.ids.NewID("conv"), user.ID, chatbotID, s.clock.Now())
	if err != nil {
		return model.Conversation{}, nil, fmt.Errorf("get conversation: %w", err)
	}
	if err := s.conversations.Create(ctx, conversation); err != nil {
		return model.Conversation{}, nil, fmt.Errorf("get conversation: %w", err)
	}
	return conversation, nil, nil
}

func (s *ChatService) StartFreshConversation(ctx context.Context, user model.UserProfile, chatbotID string) (model.Conversation, []model.Message, error) {
	chatbotID = strings.TrimSpace(chatbotID)
	if err := s.ensureUserAccess(ctx, user, chatbotID); err != nil {
		return model.Conversation{}, nil, fmt.Errorf("start fresh conversation: %w", err)
	}
	conversation, err := model.NewConversation(s.ids.NewID("conv"), user.ID, chatbotID, s.clock.Now())
	if err != nil {
		return model.Conversation{}, nil, fmt.Errorf("start fresh conversation: %w", err)
	}
	if err := s.conversations.ArchiveActiveAndCreate(ctx, user.ID, chatbotID, s.clock.Now(), conversation); err != nil {
		return model.Conversation{}, nil, fmt.Errorf("start fresh conversation: %w", err)
	}
	return conversation, nil, nil
}

func (s *ChatService) SendMessage(ctx context.Context, user model.UserProfile, chatbotID, content string) (SendMessageResult, error) {
	chatbotID = strings.TrimSpace(chatbotID)
	content = strings.TrimSpace(content)
	if content == "" {
		return SendMessageResult{}, fmt.Errorf("send message: %w", &domainerrors.ValidationError{Field: "message", Message: "must not be empty"})
	}
	if err := s.ensureUserAccess(ctx, user, chatbotID); err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}

	company, err := s.companies.GetByID(ctx, *user.CompanyID)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}
	usage, err := s.usage.GetOrCreate(ctx, company.ID, model.NormalizeMonth(s.clock.Now()), company.MonthlyTokenBudget)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}
	if usage.ExceedsBudget() {
		return SendMessageResult{}, fmt.Errorf("send message: %w", domainerrors.ErrTokenBudgetExceeded)
	}

	chatbot, err := s.chatbots.GetByID(ctx, chatbotID)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}
	files, err := s.chatbots.ListFiles(ctx, chatbotID)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}
	conversation, messages, err := s.GetOrCreateActiveConversation(ctx, user, chatbotID)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}

	systemPrompt := composeSystemPrompt(chatbot.SystemPrompt, files)
	conversationMessages := trimConversationHistory(s.tokenCounter, messages, content, s.historyMaxTokens)
	conversationMessages = append(conversationMessages, ports.AIChatMessage{
		Role:    model.ChatMessageRoleUser,
		Content: content,
	})

	response, err := s.ai.Complete(ctx, ports.AIChatRequest{
		Model:        s.modelName,
		SystemPrompt: systemPrompt,
		Conversation: conversationMessages,
	})
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}

	now := s.clock.Now()
	userMessage, err := model.NewMessage(s.ids.NewID("msg"), conversation.ID, model.ChatMessageRoleUser, content, nextSequence(messages), now)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}
	assistantMessage, err := model.NewMessage(s.ids.NewID("msg"), conversation.ID, model.ChatMessageRoleAssistant, response.Content, userMessage.Sequence+1, now)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}
	assistantMessage.InputTokens = response.InputTokens
	assistantMessage.OutputTokens = response.OutputTokens

	if err := s.conversations.AppendMessages(ctx, userMessage, assistantMessage); err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}

	_, err = s.usage.AddEvent(ctx, model.TokenUsageEvent{
		ID:             s.ids.NewID("usage"),
		CompanyID:      company.ID,
		ConversationID: conversation.ID,
		UserID:         user.ID,
		ChatbotID:      chatbot.ID,
		AssistantMsgID: assistantMessage.ID,
		MonthStart:     model.NormalizeMonth(now),
		InputTokens:    response.InputTokens,
		OutputTokens:   response.OutputTokens,
		CreatedAt:      now.UTC(),
	}, company.MonthlyTokenBudget)
	if err != nil {
		return SendMessageResult{}, fmt.Errorf("send message: %w", err)
	}

	updatedMessages := append(append([]model.Message{}, messages...), userMessage, assistantMessage)
	s.logger.Info(ctx, "chat turn persisted", "conversation_id", conversation.ID, "chatbot_id", chatbot.ID, "user_id", user.ID)
	return SendMessageResult{
		Conversation: conversation,
		Messages:     updatedMessages,
	}, nil
}

func (s *ChatService) ensureUserAccess(ctx context.Context, user model.UserProfile, chatbotID string) error {
	if user.CompanyID == nil || *user.CompanyID == "" {
		return domainerrors.ErrForbidden
	}
	hasAccess, err := s.grants.CompanyHasAccess(ctx, *user.CompanyID, chatbotID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return domainerrors.ErrForbidden
	}
	return nil
}

func composeSystemPrompt(systemPrompt string, files []model.KnowledgeFile) string {
	if len(files) == 0 {
		return systemPrompt
	}
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(systemPrompt))
	builder.WriteString("\n\nKnowledge files:\n")
	for _, file := range files {
		builder.WriteString("\n--- ")
		builder.WriteString(file.Name)
		builder.WriteString(" ---\n")
		builder.WriteString(file.ExtractedText)
		builder.WriteString("\n")
	}
	return builder.String()
}

func trimConversationHistory(counter ports.TokenCounter, history []model.Message, currentUserMessage string, maxTokens int) []ports.AIChatMessage {
	if maxTokens <= 0 {
		return nil
	}
	remaining := maxTokens - counter.CountText(currentUserMessage)
	if remaining <= 0 {
		return nil
	}

	selected := make([]model.Message, 0, len(history))
	for index := len(history) - 1; index >= 0; index-- {
		message := history[index]
		tokens := counter.CountText(message.Content)
		if tokens > remaining {
			break
		}
		remaining -= tokens
		selected = append(selected, message)
	}

	out := make([]ports.AIChatMessage, 0, len(selected))
	for index := len(selected) - 1; index >= 0; index-- {
		out = append(out, ports.AIChatMessage{
			Role:    selected[index].Role,
			Content: selected[index].Content,
		})
	}
	return out
}

func nextSequence(messages []model.Message) int {
	if len(messages) == 0 {
		return 1
	}
	return messages[len(messages)-1].Sequence + 1
}
