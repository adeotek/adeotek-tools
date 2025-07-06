-- Test script with various dollar sign scenarios
INSERT INTO products (name, price) VALUES ('Product $19.99', 19.99);

-- Column names with dollar signs
SELECT column$1, field$name FROM table$test;

-- Complex function with dollar-quoted strings containing dollars
CREATE OR REPLACE FUNCTION price_formatter(price DECIMAL) RETURNS TEXT AS $$
BEGIN
    -- Format price with dollar sign
    RETURN '$' || CAST(price AS TEXT);
END;
$$ LANGUAGE plpgsql;

-- Another function with named dollar quoting
CREATE OR REPLACE FUNCTION discount_calculator() RETURNS TEXT AS $body$
DECLARE
    message TEXT;
BEGIN
    message := 'Save $5 on orders over $50!';
    RETURN message;
END;
$body$ LANGUAGE plpgsql;
