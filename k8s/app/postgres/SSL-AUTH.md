# PostgreSQL SSL Client Certificate Authentication

## Overview

The grpcserver authenticates to PostgreSQL using SSL client certificates instead of passwords.

## Certificate Setup

### Certificates Created
1. **postgres-server-tls** - PostgreSQL server certificate
   - Issued by: postgres-ca-issuer
   - DNS names: app-postgres-rw, app-postgres-r, app-postgres-ro (all variants)

2. **grpcserver-tls** - Client certificate for grpcserver
   - Issued by: grpcserver-ca-issuer
   - Common Name: grpcserver.app-namespace.svc.cluster.local
   - Usages: server auth, client auth

### Certificate Mounts in grpcserver
- `/mnt/client-certs/tls.crt` - Client certificate
- `/mnt/client-certs/tls.key` - Client private key
- `/mnt/postgres-ca/ca.crt` - PostgreSQL CA certificate (to verify server)

## PostgreSQL Configuration

### Cluster Configuration (cluster.yaml)
```yaml
# SSL configuration is automatic when certificates section is defined
# CNPG manages SSL parameters internally

certificates:
  serverTLSSecret: postgres-server-tls
  serverCASecret: postgres-ca-bundle
  clientCASecret: postgres-ca-bundle
  replicationTLSSecret: postgres-server-tls

# Certificate authentication rules
postgresql:
  pg_hba:
    - hostssl config dbmate 0.0.0.0/0 cert clientcert=verify-full
    - hostssl config grpcserver 0.0.0.0/0 cert clientcert=verify-full
```

## Application Code Integration

### Go PostgreSQL Connection String

```go
import (
    "fmt"
    "github.com/jackc/pgx/v5"
)

func createConnectionString() string {
    return fmt.Sprintf(
        "postgresql://%s@%s:%s/%s?"+
            "sslmode=verify-full&"+
            "sslcert=%s&"+
            "sslkey=%s&"+
            "sslrootcert=%s",
        "grpcserver",  // PostgreSQL user (matches cert CN)
        "app-postgres-rw.app-namespace.svc.cluster.local",
        "5432",
        "config",
        "/mnt/client-certs/tls.crt",
        "/mnt/client-certs/tls.key",
        "/mnt/postgres-ca/ca.crt",
    )
}
```

### Connection Options
- `sslmode=verify-full` - Verify server certificate and hostname
- `sslcert` - Path to client certificate
- `sslkey` - Path to client private key
- `sslrootcert` - Path to CA certificate to verify server

## PostgreSQL User Setup

### Required: Create PostgreSQL User

A PostgreSQL user must be created that matches the client certificate's Common Name:

```sql
-- Connect to postgres as superuser
CREATE USER grpcserver;
GRANT ALL PRIVILEGES ON DATABASE config TO grpcserver;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO grpcserver;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO grpcserver;
```

### Required: Configure pg_hba.conf for Certificate Auth

CNPG needs to be configured to require certificate authentication. This can be done via a custom `pg_hba.conf`:

```
# TYPE  DATABASE  USER         ADDRESS         METHOD
hostssl app       grpcserver   0.0.0.0/0       cert clientcert=verify-full
hostssl app       app          0.0.0.0/0       scram-sha-256
```

## Implementation Steps

1. ✅ Create PostgreSQL server certificate
2. ✅ Create grpcserver client certificate
3. ✅ Configure PostgreSQL cluster to use SSL
4. ✅ Mount certificates in grpcserver deployment
5. ⏳ Update grpcserver connection string to use SSL with client cert
6. ⏳ Create PostgreSQL user 'grpcserver'
7. ⏳ Configure pg_hba.conf for certificate authentication

## Testing

```bash
# From inside grpcserver pod
psql "postgresql://grpcserver@app-postgres-rw.app-namespace.svc.cluster.local:5432/app?sslmode=verify-full&sslcert=/mnt/client-certs/tls.crt&sslkey=/mnt/client-certs/tls.key&sslrootcert=/mnt/postgres-ca/ca.crt"
```

## Security Benefits

- ✅ No passwords stored or transmitted
- ✅ Certificate-based authentication
- ✅ Automatic rotation via cert-manager
- ✅ Per-service isolation via separate CAs
- ✅ Mutual TLS (server and client authentication)
