DO
$do$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'gophermart_db') THEN
            CREATE DATABASE gophermart_db;
        END IF;
    END
$do$;

DO
$do$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'accrual_db') THEN
            CREATE DATABASE accrual_db;
        END IF;
    END
$do$;
