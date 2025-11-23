SELECT setval(pg_get_serial_sequence('teams', 'id'), COALESCE((SELECT MAX(id) FROM teams), 0) + 1);

SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE((SELECT MAX(id) FROM users), 0) + 1);

SELECT setval(pg_get_serial_sequence('pull_requests', 'id'), COALESCE((SELECT MAX(id) FROM pull_requests), 0) + 1);
