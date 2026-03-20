package postgres

import "github.com/linkai/go-chatbot-api/domain/model"

func toDomainCompany(row Company) model.Company {
	return model.Company{
		ID:                 row.ID,
		Name:               row.Name,
		MonthlyTokenBudget: row.MonthlyTokenBudget,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func toCompanyRow(company model.Company) Company {
	return Company{
		ID:                 company.ID,
		Name:               company.Name,
		MonthlyTokenBudget: company.MonthlyTokenBudget,
		CreatedAt:          company.CreatedAt,
		UpdatedAt:          company.UpdatedAt,
	}
}

func toDomainUserProfile(row UserProfile) model.UserProfile {
	return model.UserProfile{
		ID:          row.ID,
		Email:       row.Email,
		Role:        model.Role(row.Role),
		Status:      model.UserStatus(row.Status),
		CompanyID:   row.CompanyID,
		InvitedAt:   row.InvitedAt,
		ActivatedAt: row.ActivatedAt,
		DisabledAt:  row.DisabledAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toUserProfileRow(profile model.UserProfile) UserProfile {
	return UserProfile{
		ID:          profile.ID,
		Email:       profile.Email,
		Role:        string(profile.Role),
		Status:      string(profile.Status),
		CompanyID:   profile.CompanyID,
		InvitedAt:   profile.InvitedAt,
		ActivatedAt: profile.ActivatedAt,
		DisabledAt:  profile.DisabledAt,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
	}
}

func toDomainChatbot(row Chatbot) model.Chatbot {
	return model.Chatbot{
		ID:                   row.ID,
		Name:                 row.Name,
		Description:          row.Description,
		SystemPrompt:         row.SystemPrompt,
		TotalKnowledgeTokens: int(row.TotalKnowledgeTokens),
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func toChatbotRow(chatbot model.Chatbot) Chatbot {
	return Chatbot{
		ID:                   chatbot.ID,
		Name:                 chatbot.Name,
		Description:          chatbot.Description,
		SystemPrompt:         chatbot.SystemPrompt,
		TotalKnowledgeTokens: int32(chatbot.TotalKnowledgeTokens),
		CreatedAt:            chatbot.CreatedAt,
		UpdatedAt:            chatbot.UpdatedAt,
	}
}

func toDomainKnowledgeFile(row ChatbotFile) model.KnowledgeFile {
	return model.KnowledgeFile{
		ID:              row.ID,
		ChatbotID:       row.ChatbotID,
		Name:            row.Name,
		ContentType:     row.ContentType,
		SizeBytes:       row.SizeBytes,
		Kind:            model.KnowledgeFileKind(row.Kind),
		StoragePath:     row.StoragePath,
		ExtractedText:   row.ExtractedText,
		ExtractedTokens: int(row.ExtractedTokens),
		CreatedAt:       row.CreatedAt,
	}
}

func toKnowledgeFileRow(file model.KnowledgeFile) ChatbotFile {
	return ChatbotFile{
		ID:              file.ID,
		ChatbotID:       file.ChatbotID,
		Name:            file.Name,
		ContentType:     file.ContentType,
		SizeBytes:       file.SizeBytes,
		Kind:            string(file.Kind),
		StoragePath:     file.StoragePath,
		ExtractedText:   file.ExtractedText,
		ExtractedTokens: int32(file.ExtractedTokens),
		CreatedAt:       file.CreatedAt,
	}
}

func toDomainConversation(row Conversation) model.Conversation {
	return model.Conversation{
		ID:         row.ID,
		UserID:     row.UserID,
		ChatbotID:  row.ChatbotID,
		Status:     model.ConversationStatus(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		ArchivedAt: row.ArchivedAt,
	}
}

func toConversationRow(conversation model.Conversation) Conversation {
	return Conversation{
		ID:         conversation.ID,
		UserID:     conversation.UserID,
		ChatbotID:  conversation.ChatbotID,
		Status:     string(conversation.Status),
		CreatedAt:  conversation.CreatedAt,
		UpdatedAt:  conversation.UpdatedAt,
		ArchivedAt: conversation.ArchivedAt,
	}
}

func toDomainMessage(row Message) model.Message {
	return model.Message{
		ID:             row.ID,
		ConversationID: row.ConversationID,
		Role:           model.ChatMessageRole(row.Role),
		Content:        row.Content,
		Sequence:       int(row.Sequence),
		InputTokens:    int(row.InputTokens),
		OutputTokens:   int(row.OutputTokens),
		CreatedAt:      row.CreatedAt,
	}
}

func toMessageRow(message model.Message) Message {
	return Message{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           string(message.Role),
		Content:        message.Content,
		Sequence:       int32(message.Sequence),
		InputTokens:    int32(message.InputTokens),
		OutputTokens:   int32(message.OutputTokens),
		CreatedAt:      message.CreatedAt,
	}
}

func toDomainUsage(row CompanyMonthUsage) model.CompanyMonthUsage {
	return model.CompanyMonthUsage{
		CompanyID:              row.CompanyID,
		MonthStart:             row.MonthStart,
		BudgetTokens:           row.BudgetTokens,
		InputTokens:            row.InputTokens,
		OutputTokens:           row.OutputTokens,
		ManualAdjustmentTokens: row.ManualAdjustmentTokens,
		UpdatedAt:              row.UpdatedAt,
	}
}

func toUsageRow(usage model.CompanyMonthUsage) CompanyMonthUsage {
	return CompanyMonthUsage{
		CompanyID:              usage.CompanyID,
		MonthStart:             usage.MonthStart,
		BudgetTokens:           usage.BudgetTokens,
		InputTokens:            usage.InputTokens,
		OutputTokens:           usage.OutputTokens,
		ManualAdjustmentTokens: usage.ManualAdjustmentTokens,
		UpdatedAt:              usage.UpdatedAt,
	}
}
