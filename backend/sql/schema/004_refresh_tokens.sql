-- +goose Up
CREATE TABLE
    refresh_tokens (
        token TEXT PRIMARY KEY,
        expires_at TIMESTAMP NOT NULL,
        revoket_at TIMESTAMP NULL,
        user_id UUID NOT NULL REFERENCES users (id),
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );

-- +goose Down
DROP TABLE users;