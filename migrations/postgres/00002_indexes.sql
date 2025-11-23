-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_prr_pull_request_id ON pull_request_reviewers(pull_request_id);
CREATE INDEX idx_prr_reviewer_id ON pull_request_reviewers(reviewer_id);

CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);

CREATE INDEX idx_users_team_id ON users(team_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_prr_pull_request_id;
DROP INDEX IF EXISTS idx_prr_reviewer_id;

DROP INDEX IF EXISTS idx_pull_requests_author_id;

DROP INDEX IF EXISTS idx_users_team_id;
-- +goose StatementEnd
