# Functional Requirements — AI Chatbot Platform
**Version:** 0.1 (MVP)
**Status:** Draft
**Based on:** Sales briefing (NL) + extrapolated scope

---

## 1. Overview

A single-tenant web platform that allows an administrator to configure AI-powered chatbots (each with a custom system prompt and optional knowledge files) and grant access to those chatbots on a per-company basis. End users belong to one company and can only interact with chatbots their company has been granted access to.

---

## 2. Roles & Permissions

| Role | Description |
|---|---|
| **Admin** | Single superuser role. Full control over the platform. |
| **User** | End user. Belongs to exactly one company. Accesses only assigned chatbots. |

> **Out of scope (MVP):** Company-level admin roles, manager roles, or role customization.

---

## 3. Functional Requirements

### 3.1 Authentication

| ID | Requirement |
|---|---|
| AUTH-01 | Users can log in with email and password. |
| AUTH-02 | Users can request a password reset via email (forgot password flow). |
| AUTH-03 | Password reset tokens expire after 1 hour. |
| AUTH-04 | Session management via Supabase Auth (JWT-based). |
| AUTH-05 | Unauthenticated users are redirected to the login page. |

> **Extrapolated:** OAuth / SSO is out of scope for MVP. Magic link login is not required but can be enabled via Supabase without additional development effort — decision deferred.

---

### 3.2 User Invitation & Account Creation

| ID | Requirement |
|---|---|
| USR-01 | Admin can invite a user by entering their email address and selecting a company. |
| USR-02 | An invitation email is sent with a time-limited sign-up link (expire after 48 hours). |
| USR-03 | Invited user clicks the link, sets a password, and activates their account. |
| USR-04 | Each user is associated with exactly one company at account creation time. |
| USR-05 | Admin can view a list of all users, their company, and their account status (invited / active / disabled). |
| USR-06 | Admin can disable or delete a user account. Disabled users cannot log in. |
| USR-07 | Admin cannot change a user's associated company after creation (MVP constraint). |

---

### 3.3 Company Management

| ID | Requirement |
|---|---|
| COM-01 | Admin can create a company with a name. |
| COM-02 | Admin can view, edit the name of, or delete a company. |
| COM-03 | A company cannot be deleted if it still has active users (admin must reassign or remove users first). |
| COM-04 | Each company has a list of chatbots it has access to (see §3.5). |

---

### 3.4 Chatbot Management (Admin)

| ID | Requirement |
|---|---|
| BOT-01 | Admin can create a chatbot with a name, description, and system prompt. |
| BOT-02 | Admin can upload one or more files to a chatbot. These files are injected into every conversation context. |
| BOT-03 | Supported file types: PDF, TXT, DOCX, MD (plain text extraction). File size limit: 10 MB per file, max 5 files per chatbot. |
| BOT-04 | Admin can edit the name, description, system prompt, and files of an existing chatbot. |
| BOT-05 | Admin can delete a chatbot. Active access grants for that chatbot are also removed. |
| BOT-06 | Admin can view all chatbots and their current company assignments. |

> **Extrapolated:** Files are stored in Supabase Storage and their text content is extracted at upload time and prepended to the system prompt at inference time. No vector search / RAG in MVP — full file content is injected directly.

---

### 3.5 Chatbot Access Control

| ID | Requirement |
|---|---|
| ACC-01 | Admin can grant a company access to one or more chatbots. |
| ACC-02 | Admin can revoke a company's access to a chatbot at any time. |
| ACC-03 | All users belonging to a company automatically inherit access to the chatbots assigned to that company. |
| ACC-04 | Users can only see and interact with chatbots their company has been granted access to. |

> **Out of scope (MVP):** Per-user chatbot access overrides within a company.

---

### 3.6 Token Limits & Usage Management

| ID | Requirement |
|---|---|
| TOK-01 | Admin can set a monthly token budget per company (input + output tokens combined). |
| TOK-02 | Token usage is tracked per conversation turn and aggregated per company per calendar month. |
| TOK-03 | When a company reaches its monthly token limit, users of that company receive an error message and cannot continue chatting until the limit is reset or raised. |
| TOK-04 | Admin can view current token usage per company for the current month. |
| TOK-05 | Admin can manually reset or adjust the token counter for a company. |
| TOK-06 | Token counters reset automatically on the 1st of each calendar month. |

> **Extrapolated:** Token counts are obtained from the API response (`usage.input_tokens`, `usage.output_tokens`) and persisted to Supabase after each turn. No real-time enforcement mid-message in MVP — limit is checked before each new turn is submitted.

---

### 3.7 Chat Interface (User)

| ID | Requirement |
|---|---|
| CHAT-01 | After login, a user sees a list of chatbots available to their company. |
| CHAT-02 | A user can open a chatbot and send messages in a standard chat interface. |
| CHAT-03 | The system prompt (and any extracted file content) is always prepended to each conversation; users do not see it. |
| CHAT-04 | Conversation history is maintained across turns within a conversation (multi-turn). |
| CHAT-05 | The most recent conversation history is persisted between sessions and restored when the user reopens the chatbot. |
| CHAT-06 | Users can start a new / fresh conversation at any time. |

> **Out of scope (MVP):** Conversation history browser, export, sharing, or search.

---

## 4. Technical Constraints

| Area | Decision |
|---|---|
| **Database** | Supabase (Postgres) |
| **Auth** | Supabase Auth |
| **File storage** | Supabase Storage |
| **AI provider** | TBD — assumed OpenAI-compatible API (e.g. OpenAI, Azure OpenAI, or Anthropic) |
| **Multi-tenancy** | None — single admin, multiple companies as data-level separation only |
| **Deployment** | TBD |

---

## 5. Out of Scope (MVP)

- SSO / OAuth login
- Company-level admin or manager roles
- Per-user chatbot access overrides
- Vector search / RAG (file content is injected directly)
- Usage billing or invoicing
- White-labeling or custom domains
- API access for external integrations
- Audit logging

---

## 6. Open Questions

| # | Question | Owner |
|---|---|---|
| OQ-01 | Which AI provider and model(s) will be used? Is the model configurable per chatbot? | Sales / Client |
| OQ-02 | Is there a preferred frontend framework (e.g. Next.js)? | Dev |
| OQ-03 | Should the token limit be a hard block or a soft warning? | Client |
| OQ-04 | Are invitation emails sent via Supabase's built-in email or a custom provider (e.g. Resend)? | Dev |
| OQ-05 | Is there a target deployment environment (Hetzner/Kamal, Vercel, etc.)? | Dev |
