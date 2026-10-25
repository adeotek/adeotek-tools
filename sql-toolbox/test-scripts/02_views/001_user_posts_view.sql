CREATE VIEW user_posts AS
SELECT
    u.id AS user_id,
    u.username,
    u.email,
    p.id AS post_id,
    p.title,
    p.content,
    p.created_at AS post_created_at
FROM users u
LEFT JOIN posts p ON u.id = p.user_id;
