# Certificate Management with cert-manager

This directory contains the certificate infrastructure for the cluster using cert-manager and trust-manager.

## Architecture

```
selfsigned-issuer (ClusterIssuer)
  ├─> registry-ca → registry-ca-issuer → registry certificates
  ├─> grpcserver-ca → grpcserver-ca-issuer → grpcserver certificates
  └─> postgres-ca → postgres-ca-issuer → postgres certificates
```

Each service has its own Certificate Authority (CA) for mTLS.

## Components

### Root Issuer
- **selfsigned-issuer**: ClusterIssuer that creates self-signed certificates (for dev)

### Per-Service CAs
- **registry-ca**: CA for registry certificates
- **grpcserver-ca**: CA for gRPC server certificates
- **postgres-ca**: CA for PostgreSQL certificates

Each CA is valid for 10 years and uses RSA 4096-bit keys.

### ClusterIssuers
- **registry-ca-issuer**: Issues certificates signed by registry-ca
- **grpcserver-ca-issuer**: Issues certificates signed by grpcserver-ca
- **postgres-ca-issuer**: Issues certificates signed by postgres-ca

### Trust Bundles
trust-manager distributes CA certificates to all namespaces:

- **all-service-cas**: ConfigMap with all service CAs (for full mTLS mesh)
- **registry-ca-bundle**: ConfigMap with just the registry CA

These ConfigMaps are automatically created in every namespace.

## Creating Certificates for Services

### Example: gRPC Server Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: grpcserver-tls
  namespace: app-namespace
spec:
  secretName: grpcserver-tls  # Kubernetes secret name
  duration: 2160h              # 90 days
  renewBefore: 360h            # Renew 15 days before expiry
  commonName: grpcserver.app-namespace.svc.cluster.local
  dnsNames:
    - grpcserver
    - grpcserver.app-namespace.svc.cluster.local
  usages:
    - server auth              # TLS server
    - client auth              # mTLS client
  issuerRef:
    name: grpcserver-ca-issuer
    kind: ClusterIssuer
```

## Using Certificates in Go Code

### 1. Load TLS Credentials

```go
import (
    "crypto/tls"
    "crypto/x509"
    "os"
)

// Load server certificate and key
cert, err := tls.LoadX509KeyPair(
    "/mnt/certs/tls.crt",   // Certificate
    "/mnt/certs/tls.key",   // Private key
)

// Load CA bundle for client verification
caCert, err := os.ReadFile("/mnt/ca-bundle/ca-bundle.crt")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

// Create TLS config
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ClientCAs:    caCertPool,
    ClientAuth:   tls.RequireAndVerifyClientCert,  // For mTLS
}
```

### 2. Mount Certificates in Deployment

```yaml
spec:
  containers:
    - name: grpcserver
      volumeMounts:
        - name: tls-certs
          mountPath: /mnt/certs
          readOnly: true
        - name: ca-bundle
          mountPath: /mnt/ca-bundle
          readOnly: true
  volumes:
    - name: tls-certs
      secret:
        secretName: grpcserver-tls
    - name: ca-bundle
      configMap:
        name: all-service-cas
```

### 3. Example: gRPC Server with mTLS

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

func startServer() error {
    // Load TLS config (from above)
    tlsConfig := loadTLSConfig()

    // Create gRPC credentials
    creds := credentials.NewTLS(tlsConfig)

    // Create server with mTLS
    server := grpc.NewServer(
        grpc.Creds(creds),
    )

    // Register services...
    return server.Serve(listener)
}
```

### 4. Example: gRPC Client with mTLS

```go
func createClient() (*grpc.ClientConn, error) {
    // Load client certificate
    cert, _ := tls.LoadX509KeyPair(
        "/mnt/certs/tls.crt",
        "/mnt/certs/tls.key",
    )

    // Load server CA
    caCert, _ := os.ReadFile("/mnt/ca-bundle/ca-bundle.crt")
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
    }

    creds := credentials.NewTLS(tlsConfig)

    return grpc.Dial(
        "grpcserver.app-namespace.svc.cluster.local:25000",
        grpc.WithTransportCredentials(creds),
    )
}
```

## Certificate Lifecycle

1. **Creation**: cert-manager automatically creates certificates when Certificate resources are applied
2. **Storage**: Certificates are stored in Kubernetes secrets
3. **Renewal**: cert-manager automatically renews certificates before expiry
4. **Rotation**: Update your pods to reload certificates (or use a sidecar)

## Production Considerations

### Replace Self-Signed CA

For production, replace the self-signed CA with a real CA:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
```

### Certificate Rotation

Implement graceful certificate reloading:

```go
// Watch certificate file for changes
watcher, _ := fsnotify.NewWatcher()
watcher.Add("/mnt/certs/tls.crt")

go func() {
    for event := range watcher.Events {
        if event.Op&fsnotify.Write == fsnotify.Write {
            // Reload TLS config
            reloadTLSConfig()
        }
    }
}()
```

## Troubleshooting

### Check certificate status:
```bash
kubectl get certificate -A
kubectl describe certificate grpcserver-tls -n app-namespace
```

### Check secret contents:
```bash
kubectl get secret grpcserver-tls -n app-namespace -o yaml
```

### Check CA bundles:
```bash
kubectl get configmap all-service-cas -n app-namespace -o yaml
```

### Debug cert-manager:
```bash
kubectl logs -n cert-manager -l app=cert-manager
```
