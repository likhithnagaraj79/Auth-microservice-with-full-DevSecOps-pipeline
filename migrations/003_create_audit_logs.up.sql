CREATE TABLE IF NOT EXISTS audit_logs (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action     VARCHAR(100) NOT NULL,
    resource   VARCHAR(255) NOT NULL,
    ip_address VARCHAR(50) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    success    BOOLEAN NOT NULL DEFAULT TRUE,
    details    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id    ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_action     ON audit_logs(action);
