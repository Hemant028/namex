CREATE TABLE IF NOT EXISTS requests (
    timestamp DateTime,
    domain_id Int32,
    client_ip String,
    user_agent String,
    method String,
    path String,
    status Int32,
    duration Int64,
    action_taken String
) ENGINE = MergeTree()
ORDER BY (domain_id, timestamp);
