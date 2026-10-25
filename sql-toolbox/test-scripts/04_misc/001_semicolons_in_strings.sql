-- Values containing semicolons must survive statement splitting
INSERT INTO posts (user_id, title, content) VALUES
    (1, 'Semicolons', 'one; two; three');

UPDATE posts
SET content = 'updated; still one statement'
WHERE title = 'Semicolons';
