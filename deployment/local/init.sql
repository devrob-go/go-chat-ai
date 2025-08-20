-- Create the `auth_db` database if it doesn't exist
\ connect postgres
SELECT
    'CREATE DATABASE auth_db'
WHERE
    NOT EXISTS (
        SELECT
        FROM
            pg_database
        WHERE
            datname = 'auth_db'
    ) \ gexec -- Create the `chat_db` database if it doesn't exist
SELECT
    'CREATE DATABASE chat_db'
WHERE
    NOT EXISTS (
        SELECT
        FROM
            pg_database
        WHERE
            datname = 'chat_db'
    ) \ gexec