CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR UNIQUE NOT NULL,
    roles VARCHAR[] NOT NULL, 
    hashed_password BYTEA NOT NULL,
    verified BOOL NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    refresh_token VARCHAR NOT NULL,
    user_agent VARCHAR NOT NULL,
    client_ip VARCHAR NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS verifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    email VARCHAR NOT NULL,
    code VARCHAR NOT NULL,
    used BOOL NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    expired_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '15 minutes'),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS topics (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    topic_id UUID NOT NULL,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    topic_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT unique_sub_topic_user UNIQUE (topic_id, user_id)
);