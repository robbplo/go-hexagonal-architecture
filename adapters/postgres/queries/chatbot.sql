-- name: InsertChatbot :exec
INSERT INTO chatbots (id, name, description, system_prompt, total_knowledge_tokens, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetChatbotByID :one
SELECT id, name, description, system_prompt, total_knowledge_tokens, created_at, updated_at
FROM chatbots
WHERE id = $1;

-- name: ListChatbots :many
SELECT id, name, description, system_prompt, total_knowledge_tokens, created_at, updated_at
FROM chatbots
ORDER BY name ASC;

-- name: UpdateChatbot :exec
UPDATE chatbots
SET name = $2, description = $3, system_prompt = $4, total_knowledge_tokens = $5, updated_at = $6
WHERE id = $1;

-- name: DeleteChatbot :exec
DELETE FROM chatbots
WHERE id = $1;
