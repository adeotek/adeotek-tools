-- Complex stored procedure with dollar-quoted strings
CREATE OR REPLACE FUNCTION complex_procedure(user_id INT, message TEXT)
RETURNS VOID AS $$
BEGIN
    -- Create a temporary table
    CREATE TEMP TABLE temp_data (
        id INT,
        data TEXT
    );

    -- Insert some data with embedded semicolons
    INSERT INTO temp_data (id, data) VALUES
        (1, 'Hello; World'),
        (2, 'Test; Data; With; Semicolons');

    -- Another procedure with nested dollar quoting
    CREATE OR REPLACE FUNCTION nested_func() RETURNS TEXT AS $inner$
    BEGIN
        RETURN 'This is a nested function; with semicolons';
    END;
    $inner$ LANGUAGE plpgsql;

    -- Log the operation
    INSERT INTO audit_log (user_id, action, details)
    VALUES (user_id, 'complex_procedure', 'Executed with message: ' || message);

    -- Clean up
    DROP TABLE temp_data;
END;
$$ LANGUAGE plpgsql;

GRANT EXECUTE ON PROCEDURE public.complex_procedure() TO PUBLIC;

-- Another function with different dollar quotes
CREATE OR REPLACE FUNCTION another_function()
RETURNS TEXT AS $func$
DECLARE
    result TEXT;
BEGIN
    result := 'This function contains; semicolons in strings';
    RETURN result;
END;
$func$ LANGUAGE plpgsql;

GRANT EXECUTE ON PROCEDURE public.another_function() TO PUBLIC;
