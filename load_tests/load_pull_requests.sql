SELECT json_agg(json_build_object(
        'pull_request_id', pr.pull_request_id,
        'pull_request_name', pr.name,
        'author_id', pr.author_id,
        'status', pr.status,
        'reviewer_ids', (
            SELECT json_agg(u.user_id)
            FROM pull_request_reviewers prr
            JOIN users u ON u.id = prr.reviewer_id
            WHERE prr.pull_request_id = pr.id
        )
))
FROM pull_requests pr;
