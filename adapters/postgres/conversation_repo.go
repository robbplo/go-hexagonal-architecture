package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type ConversationRepo struct {
	db *pgx.Conn
	q  *Queries
}

var _ ports.ConversationRepository = (*ConversationRepo)(nil)

func NewConversationRepo(db DBTX) *ConversationRepo {
	return &ConversationRepo{q: New(db)}
}

func (r *ConversationRepo) GetActiveByUserAndChatbot(ctx context.Context, userID, chatbotID string) (model.Conversation, []model.Message, error) {
	conversationRow, err := r.q.GetActiveConversationByUserAndChatbot(ctx, userID, chatbotID)
	if err != nil {
		return model.Conversation{}, nil, fmt.Errorf("get active conversation user=%s chatbot=%s: %w", userID, chatbotID, mapError(err))
	}
	messages, err := r.ListMessages(ctx, conversationRow.ID)
	if err != nil {
		return model.Conversation{}, nil, err
	}
	return toDomainConversation(conversationRow), messages, nil
}

func (r *ConversationRepo) Create(ctx context.Context, conversation model.Conversation) error {
	if err := r.q.InsertConversation(ctx, toConversationRow(conversation)); err != nil {
		return fmt.Errorf("create conversation %s: %w", conversation.ID, mapError(err))
	}
	return nil
}

func (r *ConversationRepo) ArchiveActiveAndCreate(ctx context.Context, userID, chatbotID string, archivedAt time.Time, conversation model.Conversation) error {
	tx, err := beginTx(ctx, r.q.db)
	if err != nil {
		return fmt.Errorf("archive/create conversation: %w", err)
	}
	defer tx.Rollback(ctx)

	q := New(tx)
	if err := q.ArchiveActiveConversation(ctx, userID, chatbotID, archivedAt); err != nil {
		return fmt.Errorf("archive/create conversation: %w", mapError(err))
	}
	if err := q.InsertConversation(ctx, toConversationRow(conversation)); err != nil {
		return fmt.Errorf("archive/create conversation: %w", mapError(err))
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("archive/create conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepo) AppendMessages(ctx context.Context, messages ...model.Message) error {
	tx, err := beginTx(ctx, r.q.db)
	if err != nil {
		return fmt.Errorf("append messages: %w", err)
	}
	defer tx.Rollback(ctx)

	q := New(tx)
	for _, message := range messages {
		if err := q.InsertMessage(ctx, toMessageRow(message)); err != nil {
			return fmt.Errorf("append messages: %w", mapError(err))
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("append messages: %w", err)
	}
	return nil
}

func (r *ConversationRepo) ListMessages(ctx context.Context, conversationID string) ([]model.Message, error) {
	rows, err := r.q.ListMessages(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list messages %s: %w", conversationID, mapError(err))
	}
	out := make([]model.Message, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainMessage(row))
	}
	return out, nil
}

func beginTx(ctx context.Context, db DBTX) (pgx.Tx, error) {
	beginner, ok := db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("db does not support transactions")
	}
	return beginner.Begin(ctx)
}
