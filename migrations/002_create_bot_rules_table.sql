CREATE TABLE IF NOT EXISTS bot_rules (
    id SERIAL PRIMARY KEY,
    rule_type VARCHAR(50) NOT NULL, -- IP, UA, ASN
    value TEXT NOT NULL,
    action VARCHAR(50) NOT NULL, -- BLOCK, CHALLENGE, ALLOW
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bot_rules_value ON bot_rules(value);
