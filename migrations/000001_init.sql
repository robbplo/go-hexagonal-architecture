CREATE TABLE companies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    monthly_token_budget BIGINT NOT NULL DEFAULT 0 CHECK (monthly_token_budget >= 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_profiles (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
    status TEXT NOT NULL CHECK (status IN ('invited', 'active', 'disabled')),
    company_id TEXT REFERENCES companies (id) ON DELETE RESTRICT,
    invited_at TIMESTAMPTZ,
    activated_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE chatbots (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    total_knowledge_tokens INTEGER NOT NULL DEFAULT 0 CHECK (total_knowledge_tokens >= 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE chatbot_files (
    id TEXT PRIMARY KEY,
    chatbot_id TEXT NOT NULL REFERENCES chatbots (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    kind TEXT NOT NULL CHECK (kind IN ('pdf', 'txt', 'docx', 'md')),
    storage_path TEXT NOT NULL UNIQUE,
    extracted_text TEXT NOT NULL,
    extracted_tokens INTEGER NOT NULL CHECK (extracted_tokens > 0),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE company_chatbot_access (
    company_id TEXT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    chatbot_id TEXT NOT NULL REFERENCES chatbots (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (company_id, chatbot_id)
);

CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES user_profiles (id) ON DELETE CASCADE,
    chatbot_id TEXT NOT NULL REFERENCES chatbots (id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('active', 'archived')),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX conversations_user_chatbot_active_uidx
    ON conversations (user_id, chatbot_id)
    WHERE status = 'active';

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    input_tokens INTEGER NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
    output_tokens INTEGER NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (conversation_id, sequence)
);

CREATE TABLE company_month_usage (
    company_id TEXT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    month_start DATE NOT NULL,
    budget_tokens BIGINT NOT NULL CHECK (budget_tokens >= 0),
    input_tokens BIGINT NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
    output_tokens BIGINT NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
    manual_adjustment_tokens BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (company_id, month_start)
);

CREATE TABLE token_usage_events (
    id TEXT PRIMARY KEY,
    company_id TEXT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    conversation_id TEXT NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES user_profiles (id) ON DELETE CASCADE,
    chatbot_id TEXT NOT NULL REFERENCES chatbots (id) ON DELETE CASCADE,
    assistant_message_id TEXT NOT NULL REFERENCES messages (id) ON DELETE CASCADE,
    month_start DATE NOT NULL,
    input_tokens INTEGER NOT NULL CHECK (input_tokens >= 0),
    output_tokens INTEGER NOT NULL CHECK (output_tokens >= 0),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE OR REPLACE FUNCTION prevent_user_company_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.role = 'user' AND COALESCE(OLD.company_id, '') <> COALESCE(NEW.company_id, '') THEN
        RAISE EXCEPTION 'user company is immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_company_immutable
BEFORE UPDATE ON user_profiles
FOR EACH ROW
EXECUTE FUNCTION prevent_user_company_change();
