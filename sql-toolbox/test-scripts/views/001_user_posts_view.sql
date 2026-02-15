CREATE VIEW user_posts AS
SELECT
    u.id as user_id,
    u.username,
    u.email,
    p.id as post_id,
    p.title,
    p.content,
    p.created_at as post_created_at
FROM users u
LEFT JOIN posts p ON u.id = p.user_id;
