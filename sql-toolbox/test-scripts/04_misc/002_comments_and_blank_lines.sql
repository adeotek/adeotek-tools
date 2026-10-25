-- Comments and blank lines around statements must be ignored by the splitter

-- A statement preceded by a comment
SELECT 'dollar $ sign inside a string; and a semicolon';

-- A statement after a blank line

SELECT count(*) FROM users;
