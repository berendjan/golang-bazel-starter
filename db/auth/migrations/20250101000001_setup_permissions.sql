-- migrate:up
-- Setup users and permissions for auth database

DO $$
DECLARE
    read_write_users TEXT[] := ARRAY[
        'kratos'
    ]::TEXT[];
    read_only_users TEXT[] := ARRAY[
        'grpcserver'
    ]::TEXT[];
    username TEXT;
BEGIN
    -- Create and grant permissions to read-write users
    FOREACH username IN ARRAY read_write_users LOOP
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = username) THEN
            EXECUTE format('CREATE ROLE %I WITH LOGIN', username);
        END IF;
        EXECUTE format('GRANT ALL PRIVILEGES ON DATABASE auth TO %I', username);
        EXECUTE format('GRANT ALL PRIVILEGES ON SCHEMA public TO %I', username);
        EXECUTE format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %I', username);
        EXECUTE format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %I', username);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %I', username);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %I', username);
    END LOOP;

    -- Create and grant permissions to read-only users
    FOREACH username IN ARRAY read_only_users LOOP
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = username) THEN
            EXECUTE format('CREATE ROLE %I WITH LOGIN', username);
        END IF;
        EXECUTE format('GRANT CONNECT ON DATABASE auth TO %I', username);
        EXECUTE format('GRANT USAGE ON SCHEMA public TO %I', username);
        EXECUTE format('GRANT SELECT ON ALL TABLES IN SCHEMA public TO %I', username);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO %I', username);
    END LOOP;
END
$$;

-- migrate:down
DO $$
DECLARE
    read_write_users TEXT[] := ARRAY['kratos'];
    read_only_users TEXT[] := ARRAY['grpcserver'];
    username TEXT;
BEGIN
    -- Revoke from read-only users
    FOREACH username IN ARRAY read_only_users LOOP
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT ON TABLES FROM %I', username);
        EXECUTE format('REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM %I', username);
        EXECUTE format('REVOKE USAGE ON SCHEMA public FROM %I', username);
        EXECUTE format('REVOKE CONNECT ON DATABASE auth FROM %I', username);
        EXECUTE format('DROP ROLE IF EXISTS %I', username);
    END LOOP;

    -- Revoke from read-write users
    FOREACH username IN ARRAY read_write_users LOOP
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON SEQUENCES FROM %I', username);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM %I', username);
        EXECUTE format('REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM %I', username);
        EXECUTE format('REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM %I', username);
        EXECUTE format('REVOKE ALL PRIVILEGES ON SCHEMA public FROM %I', username);
        EXECUTE format('REVOKE ALL PRIVILEGES ON DATABASE auth FROM %I', username);
        EXECUTE format('DROP ROLE IF EXISTS %I', username);
    END LOOP;
END
$$;
