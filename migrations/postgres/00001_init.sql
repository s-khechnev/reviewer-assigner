-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
    team_name VARCHAR(128) PRIMARY KEY
);

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    external_id VARCHAR(128) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    team_name VARCHAR(128) NOT NULL REFERENCES teams(team_name) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE pull_requests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    external_id VARCHAR(128) UNIQUE NOT NULL,
    name VARCHAR(512) NOT NULL,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE pull_request_reviewers (
    pull_request_id BIGINT NOT NULL REFERENCES pull_requests(id) ON DELETE RESTRICT,
    reviewer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    PRIMARY KEY (pull_request_id, reviewer_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pull_request_reviewers;
DROP TABLE pull_requests;
DROP TYPE pr_status;
DROP TABLE users;
DROP TABLE teams;
-- +goose StatementEnd
