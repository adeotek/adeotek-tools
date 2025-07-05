-- This is an intentionally bad SQL script to test error handling
CREATE TABLE bad_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT bad_constraint FOREIGN KEY (nonexistent_column) REFERENCES nonexistent_table(id)
);
