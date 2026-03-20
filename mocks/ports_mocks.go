package mocks

import (
	"context"
	"io"
	"time"

	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
	"github.com/stretchr/testify/mock"
)

type testingT interface {
	mock.TestingT
	Cleanup(func())
}

type MockCompanyRepository struct{ mock.Mock }

func NewMockCompanyRepository(t testingT) *MockCompanyRepository {
	m := &MockCompanyRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockCompanyRepository) Create(ctx context.Context, company model.Company) error {
	return m.Called(ctx, company).Error(0)
}

func (m *MockCompanyRepository) GetByID(ctx context.Context, id string) (model.Company, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Company), args.Error(1)
}

func (m *MockCompanyRepository) List(ctx context.Context) ([]model.Company, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Company), args.Error(1)
}

func (m *MockCompanyRepository) Update(ctx context.Context, company model.Company) error {
	return m.Called(ctx, company).Error(0)
}

func (m *MockCompanyRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type MockCompanyUsageRepository struct{ mock.Mock }

func NewMockCompanyUsageRepository(t testingT) *MockCompanyUsageRepository {
	m := &MockCompanyUsageRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockCompanyUsageRepository) GetOrCreate(ctx context.Context, companyID string, month time.Time, budget int64) (model.CompanyMonthUsage, error) {
	args := m.Called(ctx, companyID, month, budget)
	return args.Get(0).(model.CompanyMonthUsage), args.Error(1)
}

func (m *MockCompanyUsageRepository) AddEvent(ctx context.Context, event model.TokenUsageEvent, budget int64) (model.CompanyMonthUsage, error) {
	args := m.Called(ctx, event, budget)
	return args.Get(0).(model.CompanyMonthUsage), args.Error(1)
}

func (m *MockCompanyUsageRepository) AdjustCurrentMonth(ctx context.Context, companyID string, month time.Time, budget int64, delta int64) (model.CompanyMonthUsage, error) {
	args := m.Called(ctx, companyID, month, budget, delta)
	return args.Get(0).(model.CompanyMonthUsage), args.Error(1)
}

func (m *MockCompanyUsageRepository) ResetCurrentMonth(ctx context.Context, companyID string, month time.Time, budget int64) (model.CompanyMonthUsage, error) {
	args := m.Called(ctx, companyID, month, budget)
	return args.Get(0).(model.CompanyMonthUsage), args.Error(1)
}

type MockUserRepository struct{ mock.Mock }

func NewMockUserRepository(t testingT) *MockUserRepository {
	m := &MockUserRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockUserRepository) Create(ctx context.Context, profile model.UserProfile) error {
	return m.Called(ctx, profile).Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (model.UserProfile, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.UserProfile), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (model.UserProfile, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(model.UserProfile), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, filter model.UserFilter) ([]model.UserProfile, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.UserProfile), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, profile model.UserProfile) error {
	return m.Called(ctx, profile).Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) CountActiveByCompany(ctx context.Context, companyID string) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) ListByCompanyAndStatuses(ctx context.Context, companyID string, statuses ...model.UserStatus) ([]model.UserProfile, error) {
	inputs := []any{ctx, companyID}
	for _, status := range statuses {
		inputs = append(inputs, status)
	}
	args := m.Called(inputs...)
	return args.Get(0).([]model.UserProfile), args.Error(1)
}

type MockChatbotRepository struct{ mock.Mock }

func NewMockChatbotRepository(t testingT) *MockChatbotRepository {
	m := &MockChatbotRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockChatbotRepository) Create(ctx context.Context, chatbot model.Chatbot) error {
	return m.Called(ctx, chatbot).Error(0)
}

func (m *MockChatbotRepository) GetByID(ctx context.Context, id string) (model.Chatbot, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Chatbot), args.Error(1)
}

func (m *MockChatbotRepository) List(ctx context.Context) ([]model.Chatbot, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Chatbot), args.Error(1)
}

func (m *MockChatbotRepository) ListByCompany(ctx context.Context, companyID string) ([]model.Chatbot, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]model.Chatbot), args.Error(1)
}

func (m *MockChatbotRepository) Update(ctx context.Context, chatbot model.Chatbot) error {
	return m.Called(ctx, chatbot).Error(0)
}

func (m *MockChatbotRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockChatbotRepository) ListFiles(ctx context.Context, chatbotID string) ([]model.KnowledgeFile, error) {
	args := m.Called(ctx, chatbotID)
	return args.Get(0).([]model.KnowledgeFile), args.Error(1)
}

func (m *MockChatbotRepository) CountFiles(ctx context.Context, chatbotID string) (int, error) {
	args := m.Called(ctx, chatbotID)
	return args.Int(0), args.Error(1)
}

func (m *MockChatbotRepository) AddFile(ctx context.Context, file model.KnowledgeFile) error {
	return m.Called(ctx, file).Error(0)
}

func (m *MockChatbotRepository) GetFile(ctx context.Context, chatbotID, fileID string) (model.KnowledgeFile, error) {
	args := m.Called(ctx, chatbotID, fileID)
	return args.Get(0).(model.KnowledgeFile), args.Error(1)
}

func (m *MockChatbotRepository) DeleteFile(ctx context.Context, chatbotID, fileID string) (model.KnowledgeFile, error) {
	args := m.Called(ctx, chatbotID, fileID)
	return args.Get(0).(model.KnowledgeFile), args.Error(1)
}

func (m *MockChatbotRepository) SetTotalKnowledgeTokens(ctx context.Context, chatbotID string, total int) error {
	return m.Called(ctx, chatbotID, total).Error(0)
}

type MockGrantRepository struct{ mock.Mock }

func NewMockGrantRepository(t testingT) *MockGrantRepository {
	m := &MockGrantRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockGrantRepository) Grant(ctx context.Context, grant model.CompanyChatbotGrant) error {
	return m.Called(ctx, grant).Error(0)
}

func (m *MockGrantRepository) Revoke(ctx context.Context, companyID, chatbotID string) error {
	return m.Called(ctx, companyID, chatbotID).Error(0)
}

func (m *MockGrantRepository) CompanyHasAccess(ctx context.Context, companyID, chatbotID string) (bool, error) {
	args := m.Called(ctx, companyID, chatbotID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGrantRepository) ListCompaniesByChatbot(ctx context.Context, chatbotID string) ([]model.Company, error) {
	args := m.Called(ctx, chatbotID)
	return args.Get(0).([]model.Company), args.Error(1)
}

type MockConversationRepository struct{ mock.Mock }

func NewMockConversationRepository(t testingT) *MockConversationRepository {
	m := &MockConversationRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockConversationRepository) GetActiveByUserAndChatbot(ctx context.Context, userID, chatbotID string) (model.Conversation, []model.Message, error) {
	args := m.Called(ctx, userID, chatbotID)
	return args.Get(0).(model.Conversation), args.Get(1).([]model.Message), args.Error(2)
}

func (m *MockConversationRepository) Create(ctx context.Context, conversation model.Conversation) error {
	return m.Called(ctx, conversation).Error(0)
}

func (m *MockConversationRepository) ArchiveActiveAndCreate(ctx context.Context, userID, chatbotID string, archivedAt time.Time, conversation model.Conversation) error {
	return m.Called(ctx, userID, chatbotID, archivedAt, conversation).Error(0)
}

func (m *MockConversationRepository) AppendMessages(ctx context.Context, messages ...model.Message) error {
	inputs := []any{ctx}
	for _, message := range messages {
		inputs = append(inputs, message)
	}
	return m.Called(inputs...).Error(0)
}

func (m *MockConversationRepository) ListMessages(ctx context.Context, conversationID string) ([]model.Message, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]model.Message), args.Error(1)
}

type MockIdentityAdmin struct{ mock.Mock }

func NewMockIdentityAdmin(t testingT) *MockIdentityAdmin {
	m := &MockIdentityAdmin{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockIdentityAdmin) InviteUser(ctx context.Context, email, redirectURL string) (ports.IdentityUser, error) {
	args := m.Called(ctx, email, redirectURL)
	return args.Get(0).(ports.IdentityUser), args.Error(1)
}

func (m *MockIdentityAdmin) DisableUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *MockIdentityAdmin) DeleteUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *MockIdentityAdmin) CreateUser(ctx context.Context, email, password string, emailConfirmed bool) (ports.IdentityUser, error) {
	args := m.Called(ctx, email, password, emailConfirmed)
	return args.Get(0).(ports.IdentityUser), args.Error(1)
}

type MockSessionAuthenticator struct{ mock.Mock }

func NewMockSessionAuthenticator(t testingT) *MockSessionAuthenticator {
	m := &MockSessionAuthenticator{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockSessionAuthenticator) Authenticate(ctx context.Context, bearer string) (ports.AuthenticatedUser, error) {
	args := m.Called(ctx, bearer)
	return args.Get(0).(ports.AuthenticatedUser), args.Error(1)
}

type MockObjectStorage struct{ mock.Mock }

func NewMockObjectStorage(t testingT) *MockObjectStorage {
	m := &MockObjectStorage{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockObjectStorage) PutObject(ctx context.Context, path, contentType string, reader io.Reader, size int64) error {
	return m.Called(ctx, path, contentType, reader, size).Error(0)
}

func (m *MockObjectStorage) DeleteObject(ctx context.Context, path string) error {
	return m.Called(ctx, path).Error(0)
}

type MockTextExtractor struct{ mock.Mock }

func NewMockTextExtractor(t testingT) *MockTextExtractor {
	m := &MockTextExtractor{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockTextExtractor) Extract(ctx context.Context, fileName, contentType string, data []byte) (ports.ExtractedDocument, error) {
	args := m.Called(ctx, fileName, contentType, data)
	return args.Get(0).(ports.ExtractedDocument), args.Error(1)
}

type MockTokenCounter struct{ mock.Mock }

func NewMockTokenCounter(t testingT) *MockTokenCounter {
	m := &MockTokenCounter{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockTokenCounter) CountText(text string) int {
	return m.Called(text).Int(0)
}

type MockAIChatClient struct{ mock.Mock }

func NewMockAIChatClient(t testingT) *MockAIChatClient {
	m := &MockAIChatClient{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockAIChatClient) Complete(ctx context.Context, req ports.AIChatRequest) (ports.AIChatResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(ports.AIChatResponse), args.Error(1)
}

type MockClock struct{ mock.Mock }

func NewMockClock(t testingT) *MockClock {
	m := &MockClock{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockClock) Now() time.Time {
	return m.Called().Get(0).(time.Time)
}

type MockIDGenerator struct{ mock.Mock }

func NewMockIDGenerator(t testingT) *MockIDGenerator {
	m := &MockIDGenerator{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockIDGenerator) NewID(prefix string) string {
	return m.Called(prefix).String(0)
}

type MockLogger struct{ mock.Mock }

func NewMockLogger(t testingT) *MockLogger {
	m := &MockLogger{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockLogger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	inputs := []any{ctx, msg}
	inputs = append(inputs, keysAndValues...)
	m.Called(inputs...)
}

func (m *MockLogger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	inputs := []any{ctx, msg}
	inputs = append(inputs, keysAndValues...)
	m.Called(inputs...)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	inputs := []any{ctx, msg}
	inputs = append(inputs, keysAndValues...)
	m.Called(inputs...)
}

func (m *MockLogger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	inputs := []any{ctx, msg}
	inputs = append(inputs, keysAndValues...)
	m.Called(inputs...)
}
