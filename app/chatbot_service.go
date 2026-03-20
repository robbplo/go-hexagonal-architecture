package app

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

var unsafeFileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type ChatbotService struct {
	chatbots           ports.ChatbotRepository
	companies          ports.CompanyRepository
	grants             ports.GrantRepository
	storage            ports.ObjectStorage
	extractor          ports.TextExtractor
	tokenCounter       ports.TokenCounter
	clock              ports.Clock
	ids                ports.IDGenerator
	logger             ports.Logger
	maxFileBytes       int64
	maxFilesPerBot     int
	knowledgeMaxTokens int
	allowedTypes       map[string]struct{}
}

type CreateChatbotCommand struct {
	Name         string
	Description  string
	SystemPrompt string
}

type UpdateChatbotCommand struct {
	ChatbotID    string
	Name         string
	Description  string
	SystemPrompt string
}

type UploadKnowledgeFileCommand struct {
	ChatbotID   string
	FileName    string
	ContentType string
	SizeBytes   int64
	Data        []byte
}

type AdminChatbotView struct {
	Chatbot   model.Chatbot
	Files     []model.KnowledgeFile
	Companies []model.Company
}

func NewChatbotService(
	chatbots ports.ChatbotRepository,
	companies ports.CompanyRepository,
	grants ports.GrantRepository,
	storage ports.ObjectStorage,
	extractor ports.TextExtractor,
	tokenCounter ports.TokenCounter,
	clock ports.Clock,
	ids ports.IDGenerator,
	logger ports.Logger,
	maxFileBytes int64,
	maxFilesPerBot int,
	knowledgeMaxTokens int,
	allowedTypes []string,
) *ChatbotService {
	if chatbots == nil {
		panic("ChatbotService: chatbots is required")
	}
	if companies == nil {
		panic("ChatbotService: companies is required")
	}
	if grants == nil {
		panic("ChatbotService: grants is required")
	}
	if storage == nil {
		panic("ChatbotService: storage is required")
	}
	if extractor == nil {
		panic("ChatbotService: extractor is required")
	}
	if tokenCounter == nil {
		panic("ChatbotService: tokenCounter is required")
	}
	if clock == nil {
		panic("ChatbotService: clock is required")
	}
	if ids == nil {
		panic("ChatbotService: ids is required")
	}
	if logger == nil {
		panic("ChatbotService: logger is required")
	}

	types := make(map[string]struct{}, len(allowedTypes))
	for _, value := range allowedTypes {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		types[value] = struct{}{}
	}

	return &ChatbotService{
		chatbots:           chatbots,
		companies:          companies,
		grants:             grants,
		storage:            storage,
		extractor:          extractor,
		tokenCounter:       tokenCounter,
		clock:              clock,
		ids:                ids,
		logger:             logger,
		maxFileBytes:       maxFileBytes,
		maxFilesPerBot:     maxFilesPerBot,
		knowledgeMaxTokens: knowledgeMaxTokens,
		allowedTypes:       types,
	}
}

func (s *ChatbotService) Create(ctx context.Context, cmd CreateChatbotCommand) (model.Chatbot, error) {
	chatbot, err := model.NewChatbot(s.ids.NewID("bot"), cmd.Name, cmd.Description, cmd.SystemPrompt, s.clock.Now())
	if err != nil {
		return model.Chatbot{}, fmt.Errorf("create chatbot: %w", err)
	}
	if err := s.chatbots.Create(ctx, chatbot); err != nil {
		return model.Chatbot{}, fmt.Errorf("create chatbot: %w", err)
	}
	s.logger.Info(ctx, "chatbot created", "chatbot_id", chatbot.ID)
	return chatbot, nil
}

func (s *ChatbotService) List(ctx context.Context) ([]AdminChatbotView, error) {
	chatbots, err := s.chatbots.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list chatbots: %w", err)
	}
	views := make([]AdminChatbotView, 0, len(chatbots))
	for _, chatbot := range chatbots {
		files, err := s.chatbots.ListFiles(ctx, chatbot.ID)
		if err != nil {
			return nil, fmt.Errorf("list chatbots: %w", err)
		}
		companies, err := s.grants.ListCompaniesByChatbot(ctx, chatbot.ID)
		if err != nil {
			return nil, fmt.Errorf("list chatbots: %w", err)
		}
		views = append(views, AdminChatbotView{
			Chatbot:   chatbot,
			Files:     files,
			Companies: companies,
		})
	}
	return views, nil
}

func (s *ChatbotService) Update(ctx context.Context, cmd UpdateChatbotCommand) (model.Chatbot, error) {
	chatbot, err := s.chatbots.GetByID(ctx, strings.TrimSpace(cmd.ChatbotID))
	if err != nil {
		return model.Chatbot{}, fmt.Errorf("update chatbot: %w", err)
	}
	chatbot, err = chatbot.WithUpdatedFields(cmd.Name, cmd.Description, cmd.SystemPrompt, s.clock.Now())
	if err != nil {
		return model.Chatbot{}, fmt.Errorf("update chatbot: %w", err)
	}
	if err := s.chatbots.Update(ctx, chatbot); err != nil {
		return model.Chatbot{}, fmt.Errorf("update chatbot: %w", err)
	}
	s.logger.Info(ctx, "chatbot updated", "chatbot_id", chatbot.ID)
	return chatbot, nil
}

func (s *ChatbotService) Delete(ctx context.Context, chatbotID string) error {
	if err := s.chatbots.Delete(ctx, strings.TrimSpace(chatbotID)); err != nil {
		return fmt.Errorf("delete chatbot: %w", err)
	}
	s.logger.Info(ctx, "chatbot deleted", "chatbot_id", chatbotID)
	return nil
}

func (s *ChatbotService) UploadFile(ctx context.Context, cmd UploadKnowledgeFileCommand) (model.KnowledgeFile, error) {
	chatbotID := strings.TrimSpace(cmd.ChatbotID)
	if _, err := s.chatbots.GetByID(ctx, chatbotID); err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	if cmd.SizeBytes <= 0 || cmd.SizeBytes > s.maxFileBytes {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", &domainerrors.ValidationError{Field: "file", Message: "size exceeds configured limit"})
	}
	if _, ok := s.allowedTypes[strings.TrimSpace(cmd.ContentType)]; !ok {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", domainerrors.ErrUnsupportedFileType)
	}

	fileCount, err := s.chatbots.CountFiles(ctx, chatbotID)
	if err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	if fileCount >= s.maxFilesPerBot {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", &domainerrors.ValidationError{Field: "files", Message: "maximum files per chatbot reached"})
	}

	extracted, err := s.extractor.Extract(ctx, cmd.FileName, cmd.ContentType, cmd.Data)
	if err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	tokens := s.tokenCounter.CountText(extracted.Text)
	if tokens <= 0 {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", &domainerrors.ValidationError{Field: "file", Message: "extracted text is empty"})
	}

	files, err := s.chatbots.ListFiles(ctx, chatbotID)
	if err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	totalTokens := tokens
	for _, file := range files {
		totalTokens += file.ExtractedTokens
	}
	if err := model.ValidateKnowledgeTokenBudget(totalTokens, s.knowledgeMaxTokens); err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}

	fileID := s.ids.NewID("file")
	storagePath := fmt.Sprintf("chatbot-files/%s/%s-%s", chatbotID, fileID, safeFileName(cmd.FileName))
	if err := s.storage.PutObject(ctx, storagePath, cmd.ContentType, bytes.NewReader(cmd.Data), cmd.SizeBytes); err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}

	file, err := model.NewKnowledgeFile(fileID, chatbotID, cmd.FileName, cmd.ContentType, storagePath, extracted.Text, cmd.SizeBytes, tokens, s.clock.Now())
	if err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	if err := s.chatbots.AddFile(ctx, file); err != nil {
		_ = s.storage.DeleteObject(ctx, storagePath)
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	if err := s.chatbots.SetTotalKnowledgeTokens(ctx, chatbotID, totalTokens); err != nil {
		_ = s.storage.DeleteObject(ctx, storagePath)
		_, _ = s.chatbots.DeleteFile(ctx, chatbotID, file.ID)
		return model.KnowledgeFile{}, fmt.Errorf("upload chatbot file: %w", err)
	}
	s.logger.Info(ctx, "chatbot file uploaded", "chatbot_id", chatbotID, "file_id", file.ID)
	return file, nil
}

func (s *ChatbotService) DeleteFile(ctx context.Context, chatbotID, fileID string) error {
	file, err := s.chatbots.GetFile(ctx, strings.TrimSpace(chatbotID), strings.TrimSpace(fileID))
	if err != nil {
		return fmt.Errorf("delete chatbot file: %w", err)
	}
	if err := s.storage.DeleteObject(ctx, file.StoragePath); err != nil {
		return fmt.Errorf("delete chatbot file: %w", err)
	}
	if _, err := s.chatbots.DeleteFile(ctx, file.ChatbotID, file.ID); err != nil {
		return fmt.Errorf("delete chatbot file: %w", err)
	}
	files, err := s.chatbots.ListFiles(ctx, file.ChatbotID)
	if err != nil {
		return fmt.Errorf("delete chatbot file: %w", err)
	}
	totalTokens := 0
	for _, item := range files {
		totalTokens += item.ExtractedTokens
	}
	if err := s.chatbots.SetTotalKnowledgeTokens(ctx, file.ChatbotID, totalTokens); err != nil {
		return fmt.Errorf("delete chatbot file: %w", err)
	}
	s.logger.Info(ctx, "chatbot file deleted", "chatbot_id", file.ChatbotID, "file_id", file.ID)
	return nil
}

func (s *ChatbotService) GrantAccess(ctx context.Context, companyID, chatbotID string) error {
	companyID = strings.TrimSpace(companyID)
	chatbotID = strings.TrimSpace(chatbotID)
	if _, err := s.companies.GetByID(ctx, companyID); err != nil {
		return fmt.Errorf("grant chatbot access: %w", err)
	}
	if _, err := s.chatbots.GetByID(ctx, chatbotID); err != nil {
		return fmt.Errorf("grant chatbot access: %w", err)
	}
	if err := s.grants.Grant(ctx, model.CompanyChatbotGrant{
		CompanyID: companyID,
		ChatbotID: chatbotID,
		CreatedAt: s.clock.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("grant chatbot access: %w", err)
	}
	s.logger.Info(ctx, "chatbot access granted", "company_id", companyID, "chatbot_id", chatbotID)
	return nil
}

func (s *ChatbotService) RevokeAccess(ctx context.Context, companyID, chatbotID string) error {
	if err := s.grants.Revoke(ctx, strings.TrimSpace(companyID), strings.TrimSpace(chatbotID)); err != nil {
		return fmt.Errorf("revoke chatbot access: %w", err)
	}
	s.logger.Info(ctx, "chatbot access revoked", "company_id", companyID, "chatbot_id", chatbotID)
	return nil
}

func safeFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = unsafeFileNamePattern.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		return "file"
	}
	return base
}
