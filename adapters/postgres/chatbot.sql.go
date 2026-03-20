package postgres

import "context"

const insertChatbotSQL = `
INSERT INTO chatbots (id, name, description, system_prompt, total_knowledge_tokens, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

const getChatbotByIDSQL = `
SELECT id, name, description, system_prompt, total_knowledge_tokens, created_at, updated_at
FROM chatbots
WHERE id = $1
`

const listChatbotsSQL = `
SELECT id, name, description, system_prompt, total_knowledge_tokens, created_at, updated_at
FROM chatbots
ORDER BY name ASC
`

const listChatbotsByCompanySQL = `
SELECT c.id, c.name, c.description, c.system_prompt, c.total_knowledge_tokens, c.created_at, c.updated_at
FROM chatbots c
INNER JOIN company_chatbot_access cca ON cca.chatbot_id = c.id
WHERE cca.company_id = $1
ORDER BY c.name ASC
`

const updateChatbotSQL = `
UPDATE chatbots
SET name = $2, description = $3, system_prompt = $4, total_knowledge_tokens = $5, updated_at = $6
WHERE id = $1
`

const deleteChatbotSQL = `
DELETE FROM chatbots
WHERE id = $1
`

const listChatbotFilesSQL = `
SELECT id, chatbot_id, name, content_type, size_bytes, kind, storage_path, extracted_text, extracted_tokens, created_at
FROM chatbot_files
WHERE chatbot_id = $1
ORDER BY created_at ASC
`

const countChatbotFilesSQL = `
SELECT count(*) FROM chatbot_files WHERE chatbot_id = $1
`

const insertChatbotFileSQL = `
INSERT INTO chatbot_files (
	id, chatbot_id, name, content_type, size_bytes, kind, storage_path, extracted_text, extracted_tokens, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const getChatbotFileSQL = `
SELECT id, chatbot_id, name, content_type, size_bytes, kind, storage_path, extracted_text, extracted_tokens, created_at
FROM chatbot_files
WHERE chatbot_id = $1 AND id = $2
`

const deleteChatbotFileSQL = `
DELETE FROM chatbot_files
WHERE chatbot_id = $1 AND id = $2
RETURNING id, chatbot_id, name, content_type, size_bytes, kind, storage_path, extracted_text, extracted_tokens, created_at
`

const setChatbotKnowledgeTokensSQL = `
UPDATE chatbots
SET total_knowledge_tokens = $2, updated_at = NOW()
WHERE id = $1
`

const grantAccessSQL = `
INSERT INTO company_chatbot_access (company_id, chatbot_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (company_id, chatbot_id) DO NOTHING
`

const revokeAccessSQL = `
DELETE FROM company_chatbot_access
WHERE company_id = $1 AND chatbot_id = $2
`

const companyHasAccessSQL = `
SELECT EXISTS(
	SELECT 1
	FROM company_chatbot_access
	WHERE company_id = $1 AND chatbot_id = $2
)
`

const listCompaniesByChatbotSQL = `
SELECT c.id, c.name, c.monthly_token_budget, c.created_at, c.updated_at
FROM companies c
INNER JOIN company_chatbot_access cca ON cca.company_id = c.id
WHERE cca.chatbot_id = $1
ORDER BY c.name ASC
`

func (q *Queries) InsertChatbot(ctx context.Context, arg Chatbot) error {
	_, err := q.db.Exec(ctx, insertChatbotSQL,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.SystemPrompt,
		arg.TotalKnowledgeTokens,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

func (q *Queries) GetChatbotByID(ctx context.Context, id string) (Chatbot, error) {
	row := q.db.QueryRow(ctx, getChatbotByIDSQL, id)
	var chatbot Chatbot
	err := row.Scan(&chatbot.ID, &chatbot.Name, &chatbot.Description, &chatbot.SystemPrompt, &chatbot.TotalKnowledgeTokens, &chatbot.CreatedAt, &chatbot.UpdatedAt)
	return chatbot, err
}

func (q *Queries) ListChatbots(ctx context.Context) ([]Chatbot, error) {
	rows, err := q.db.Query(ctx, listChatbotsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Chatbot
	for rows.Next() {
		var chatbot Chatbot
		if err := rows.Scan(&chatbot.ID, &chatbot.Name, &chatbot.Description, &chatbot.SystemPrompt, &chatbot.TotalKnowledgeTokens, &chatbot.CreatedAt, &chatbot.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, chatbot)
	}
	return out, rows.Err()
}

func (q *Queries) ListChatbotsByCompany(ctx context.Context, companyID string) ([]Chatbot, error) {
	rows, err := q.db.Query(ctx, listChatbotsByCompanySQL, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Chatbot
	for rows.Next() {
		var chatbot Chatbot
		if err := rows.Scan(&chatbot.ID, &chatbot.Name, &chatbot.Description, &chatbot.SystemPrompt, &chatbot.TotalKnowledgeTokens, &chatbot.CreatedAt, &chatbot.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, chatbot)
	}
	return out, rows.Err()
}

func (q *Queries) UpdateChatbot(ctx context.Context, arg Chatbot) error {
	_, err := q.db.Exec(ctx, updateChatbotSQL, arg.ID, arg.Name, arg.Description, arg.SystemPrompt, arg.TotalKnowledgeTokens, arg.UpdatedAt)
	return err
}

func (q *Queries) DeleteChatbot(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteChatbotSQL, id)
	return err
}

func (q *Queries) ListChatbotFiles(ctx context.Context, chatbotID string) ([]ChatbotFile, error) {
	rows, err := q.db.Query(ctx, listChatbotFilesSQL, chatbotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChatbotFile
	for rows.Next() {
		var file ChatbotFile
		if err := rows.Scan(&file.ID, &file.ChatbotID, &file.Name, &file.ContentType, &file.SizeBytes, &file.Kind, &file.StoragePath, &file.ExtractedText, &file.ExtractedTokens, &file.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, file)
	}
	return out, rows.Err()
}

func (q *Queries) CountChatbotFiles(ctx context.Context, chatbotID string) (int, error) {
	row := q.db.QueryRow(ctx, countChatbotFilesSQL, chatbotID)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) InsertChatbotFile(ctx context.Context, arg ChatbotFile) error {
	_, err := q.db.Exec(ctx, insertChatbotFileSQL,
		arg.ID,
		arg.ChatbotID,
		arg.Name,
		arg.ContentType,
		arg.SizeBytes,
		arg.Kind,
		arg.StoragePath,
		arg.ExtractedText,
		arg.ExtractedTokens,
		arg.CreatedAt,
	)
	return err
}

func (q *Queries) GetChatbotFile(ctx context.Context, chatbotID, fileID string) (ChatbotFile, error) {
	row := q.db.QueryRow(ctx, getChatbotFileSQL, chatbotID, fileID)
	var file ChatbotFile
	err := row.Scan(&file.ID, &file.ChatbotID, &file.Name, &file.ContentType, &file.SizeBytes, &file.Kind, &file.StoragePath, &file.ExtractedText, &file.ExtractedTokens, &file.CreatedAt)
	return file, err
}

func (q *Queries) DeleteChatbotFile(ctx context.Context, chatbotID, fileID string) (ChatbotFile, error) {
	row := q.db.QueryRow(ctx, deleteChatbotFileSQL, chatbotID, fileID)
	var file ChatbotFile
	err := row.Scan(&file.ID, &file.ChatbotID, &file.Name, &file.ContentType, &file.SizeBytes, &file.Kind, &file.StoragePath, &file.ExtractedText, &file.ExtractedTokens, &file.CreatedAt)
	return file, err
}

func (q *Queries) SetChatbotKnowledgeTokens(ctx context.Context, chatbotID string, totalTokens int) error {
	_, err := q.db.Exec(ctx, setChatbotKnowledgeTokensSQL, chatbotID, totalTokens)
	return err
}

func (q *Queries) GrantAccess(ctx context.Context, companyID, chatbotID string, createdAt any) error {
	_, err := q.db.Exec(ctx, grantAccessSQL, companyID, chatbotID, createdAt)
	return err
}

func (q *Queries) RevokeAccess(ctx context.Context, companyID, chatbotID string) error {
	_, err := q.db.Exec(ctx, revokeAccessSQL, companyID, chatbotID)
	return err
}

func (q *Queries) CompanyHasAccess(ctx context.Context, companyID, chatbotID string) (bool, error) {
	row := q.db.QueryRow(ctx, companyHasAccessSQL, companyID, chatbotID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func (q *Queries) ListCompaniesByChatbot(ctx context.Context, chatbotID string) ([]Company, error) {
	rows, err := q.db.Query(ctx, listCompaniesByChatbotSQL, chatbotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		if err := rows.Scan(&company.ID, &company.Name, &company.MonthlyTokenBudget, &company.CreatedAt, &company.UpdatedAt); err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, rows.Err()
}
