package postgres

import "context"

const getCompanyMonthUsageSQL = `
SELECT company_id, month_start, budget_tokens, input_tokens, output_tokens, manual_adjustment_tokens, updated_at
FROM company_month_usage
WHERE company_id = $1 AND month_start = $2
`

const upsertCompanyMonthUsageSQL = `
INSERT INTO company_month_usage (
	company_id, month_start, budget_tokens, input_tokens, output_tokens, manual_adjustment_tokens, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (company_id, month_start) DO UPDATE
SET budget_tokens = EXCLUDED.budget_tokens,
	input_tokens = EXCLUDED.input_tokens,
	output_tokens = EXCLUDED.output_tokens,
	manual_adjustment_tokens = EXCLUDED.manual_adjustment_tokens,
	updated_at = EXCLUDED.updated_at
`

const insertTokenUsageEventSQL = `
INSERT INTO token_usage_events (
	id, company_id, conversation_id, user_id, chatbot_id, assistant_message_id, month_start, input_tokens, output_tokens, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

func (q *Queries) GetCompanyMonthUsage(ctx context.Context, companyID string, monthStart any) (CompanyMonthUsage, error) {
	row := q.db.QueryRow(ctx, getCompanyMonthUsageSQL, companyID, monthStart)
	var usage CompanyMonthUsage
	err := row.Scan(&usage.CompanyID, &usage.MonthStart, &usage.BudgetTokens, &usage.InputTokens, &usage.OutputTokens, &usage.ManualAdjustmentTokens, &usage.UpdatedAt)
	return usage, err
}

func (q *Queries) UpsertCompanyMonthUsage(ctx context.Context, arg CompanyMonthUsage) error {
	_, err := q.db.Exec(ctx, upsertCompanyMonthUsageSQL,
		arg.CompanyID,
		arg.MonthStart,
		arg.BudgetTokens,
		arg.InputTokens,
		arg.OutputTokens,
		arg.ManualAdjustmentTokens,
		arg.UpdatedAt,
	)
	return err
}

func (q *Queries) InsertTokenUsageEvent(ctx context.Context, companyID, conversationID, userID, chatbotID, assistantMessageID string, monthStart any, inputTokens, outputTokens int, eventID string, createdAt any) error {
	_, err := q.db.Exec(ctx, insertTokenUsageEventSQL, eventID, companyID, conversationID, userID, chatbotID, assistantMessageID, monthStart, inputTokens, outputTokens, createdAt)
	return err
}
