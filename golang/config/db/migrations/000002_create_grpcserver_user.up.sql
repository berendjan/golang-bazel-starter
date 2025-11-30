-- Create grpcserver user for SSL certificate authentication
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'grpcserver') THEN
        CREATE USER grpcserver;
    END IF;
END
$$;

-- Grant privileges on database
GRANT ALL PRIVILEGES ON DATABASE app TO grpcserver;

-- Grant privileges on schema
GRANT ALL PRIVILEGES ON SCHEMA public TO grpcserver;

-- Grant privileges on existing tables and sequences
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO grpcserver;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO grpcserver;

-- Grant privileges on future tables and sequences
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO grpcserver;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO grpcserver;
