-- name: InsertConversation :exec
INSERT INTO conversations (id, user_id, chatbot_id, status, created_at, updated_at, archived_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: InsertMessage :exec
INSERT INTO messages (id, conversation_id, role, content, sequence, input_tokens, output_tokens, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
