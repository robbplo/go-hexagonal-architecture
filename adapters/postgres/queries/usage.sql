-- name: InsertTokenUsageEvent :exec
INSERT INTO token_usage_events (
    id, company_id, conversation_id, user_id, chatbot_id, assistant_message_id, month_start, input_tokens, output_tokens, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
