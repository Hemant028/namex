CREATE TABLE IF NOT EXISTS dns_records (
    id SERIAL PRIMARY KEY,
    domain_id INT REFERENCES domains(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL, -- A, CNAME, TXT, MX, NS
    name VARCHAR(255) NOT NULL, -- subdomain or @
    content TEXT NOT NULL,
    ttl INT DEFAULT 300,
    priority INT DEFAULT 0, -- For MX
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dns_records_domain_id ON dns_records(domain_id);
CREATE INDEX idx_dns_records_name ON dns_records(name);
