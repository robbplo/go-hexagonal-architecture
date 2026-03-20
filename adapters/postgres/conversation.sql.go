package postgres

import "context"

const getActiveConversationByUserAndChatbotSQL = `
SELECT id, user_id, chatbot_id, status, created_at, updated_at, archived_at
FROM conversations
WHERE user_id = $1 AND chatbot_id = $2 AND status = 'active'
`

const insertConversationSQL = `
INSERT INTO conversations (id, user_id, chatbot_id, status, created_at, updated_at, archived_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

const archiveActiveConversationSQL = `
UPDATE conversations
SET status = 'archived', updated_at = $3, archived_at = $3
WHERE user_id = $1 AND chatbot_id = $2 AND status = 'active'
`

const insertMessageSQL = `
INSERT INTO messages (id, conversation_id, role, content, sequence, input_tokens, output_tokens, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

const listMessagesSQL = `
SELECT id, conversation_id, role, content, sequence, input_tokens, output_tokens, created_at
FROM messages
WHERE conversation_id = $1
ORDER BY sequence ASC
`

func (q *Queries) GetActiveConversationByUserAndChatbot(ctx context.Context, userID, chatbotID string) (Conversation, error) {
	row := q.db.QueryRow(ctx, getActiveConversationByUserAndChatbotSQL, userID, chatbotID)
	var conversation Conversation
	err := row.Scan(&conversation.ID, &conversation.UserID, &conversation.ChatbotID, &conversation.Status, &conversation.CreatedAt, &conversation.UpdatedAt, &conversation.ArchivedAt)
	return conversation, err
}

func (q *Queries) InsertConversation(ctx context.Context, arg Conversation) error {
	_, err := q.db.Exec(ctx, insertConversationSQL, arg.ID, arg.UserID, arg.ChatbotID, arg.Status, arg.CreatedAt, arg.UpdatedAt, arg.ArchivedAt)
	return err
}

func (q *Queries) ArchiveActiveConversation(ctx context.Context, userID, chatbotID string, archivedAt any) error {
	_, err := q.db.Exec(ctx, archiveActiveConversationSQL, userID, chatbotID, archivedAt)
	return err
}

func (q *Queries) InsertMessage(ctx context.Context, arg Message) error {
	_, err := q.db.Exec(ctx, insertMessageSQL, arg.ID, arg.ConversationID, arg.Role, arg.Content, arg.Sequence, arg.InputTokens, arg.OutputTokens, arg.CreatedAt)
	return err
}

func (q *Queries) ListMessages(ctx context.Context, conversationID string) ([]Message, error) {
	rows, err := q.db.Query(ctx, listMessagesSQL, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.ConversationID, &message.Role, &message.Content, &message.Sequence, &message.InputTokens, &message.OutputTokens, &message.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, message)
	}
	return out, rows.Err()
}
