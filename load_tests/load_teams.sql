SELECT json_agg(json_build_object(
        'team_name', t.name,
        'members', (
            SELECT json_agg(json_build_object(
                    'user_id', u.user_id,
                    'username', u.name,
                    'is_active', u.is_active))
            FROM users u
            WHERE u.team_id = t.id
        )
))
FROM teams t
WHERE EXISTS (SELECT 1 FROM users u WHERE u.team_id = t.id);
