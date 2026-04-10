CREATE TABLE operator_sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login         VARCHAR(255) NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE proxies (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host       VARCHAR(255) NOT NULL,
    port       INT NOT NULL,
    username   VARCHAR(255),
    password   VARCHAR(255),
    status     VARCHAR(50) NOT NULL DEFAULT 'healthy',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE accounts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone             VARCHAR(20) UNIQUE NOT NULL,
    proxy_id          UUID REFERENCES proxies(id) ON DELETE SET NULL,
    status            VARCHAR(50) NOT NULL DEFAULT 'WARMING_UP',
    channel           VARCHAR(20) NOT NULL DEFAULT 'telegram',
    daily_check_count INT NOT NULL DEFAULT 0,
    daily_send_count  INT NOT NULL DEFAULT 0,
    credentials       TEXT,
    cooldown_until    TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
    account_id    UUID PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    session_data  BYTEA,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE account_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    payload    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_account_events_account_id ON account_events(account_id);
CREATE INDEX idx_account_events_created_at ON account_events(created_at DESC);

CREATE TABLE campaigns (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    status       VARCHAR(50) NOT NULL DEFAULT 'draft',
    scheduled_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE message_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    channel     VARCHAR(20) NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_message_templates_campaign_channel ON message_templates(campaign_id, channel);

CREATE TABLE contacts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash   TEXT UNIQUE NOT NULL,
    phone        TEXT NOT NULL,
    name         TEXT,
    extra_data   TEXT,
    has_replied  BOOLEAN NOT NULL DEFAULT FALSE,
    replied_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contact_replies (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    channel    VARCHAR(20) NOT NULL DEFAULT 'telegram',
    message    TEXT,
    replied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_contact_replies_contact_id ON contact_replies(contact_id);
CREATE INDEX idx_contact_replies_replied_at ON contact_replies(replied_at DESC);

CREATE TABLE contact_channel_preferences (
    contact_id        UUID PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
    preferred_channel VARCHAR(20) NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE campaign_contacts (
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    status      VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (campaign_id, contact_id)
);

CREATE TABLE send_attempts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key UUID UNIQUE NOT NULL,
    contact_id      UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    campaign_id     UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    proxy_id        UUID REFERENCES proxies(id) ON DELETE SET NULL,
    channel         VARCHAR(20) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    error_code      VARCHAR(100),
    error_message   TEXT,
    latency_ms      BIGINT,
    attempt_number  INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_send_attempts_in_progress ON send_attempts(status, updated_at) WHERE status = 'in_progress';
CREATE INDEX idx_send_attempts_contact_campaign ON send_attempts(contact_id, campaign_id);
