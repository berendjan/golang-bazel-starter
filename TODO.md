# TODO List

## Documentation
- [ ] Document dev/prod overlay pattern for registry (HTTP for dev, TLS for prod)
- [ ] Document certificate architecture (per-service CAs, ClusterIssuers, trust-manager)
- [ ] Update dev/README.md with current deployment flow and init-cluster steps
- [ ] Document deploy.sh usage and quick iteration workflow

## Database & Server
- [ ] Fix grpcserver PostgreSQL connection in Kind cluster
- [ ] Configure grpcserver to use cert-manager SSL certificate for PostgreSQL authentication
- [ ] Create PostgreSQL role/user for grpcserver that authenticates via SSL certificate
- [ ] Create separate golang binary for database migrations (independent from grpcserver)
- [ ] Remove migration logic from grpcserver and delegate to migration binary
- [ ] Adjust the code generation so it generates the messenger with only the specified handlers, should throw on illegal config

## Authentication
- [ ] Add Ory Kratos for user authentication
- [ ] Users information should be added to postgresql itself

## Message Queue & Background Jobs
- [ ] Implement custom pgqueue (PostgreSQL-based message queue)
- [ ] Create golang api to process pgqueue jobs
- [ ] Add job scheduling and retry logic to pgqueue

## Monitoring & Observability
- [ ] Add Prometheus for metrics collection
- [ ] Add Grafana for metrics visualization and dashboards
- [ ] Configure ServiceMonitors for application metrics
- [ ] Create Grafana dashboards for gRPC server, PostgreSQL, and infrastructure

## Frontend
- [ ] Set up React frontend with Bazel (rules_js/aspect_rules_js)
- [ ] Configure TypeScript for type safety
- [ ] Create OCI image build for React frontend
- [ ] Add Kubernetes deployment for frontend with nginx
- [ ] Configure ingress for frontend routing

## CI/CD
- [ ] Set up GitHub Actions workflow for continuous integration
- [ ] Add linting and formatting checks (gofmt, buildifier)
- [ ] Add automated testing (unit tests, integration tests)
- [ ] Add Bazel remote caching for faster CI builds
- [ ] Add container image building and pushing to registry
- [ ] Add terraform validation and linting (terraform fmt, terraform validate)
- [ ] Add ansible linting (ansible-lint)
- [ ] Add automated writing to manifest repository after successful builds
- [ ] Add Flux for GitOps deployments from manifest repository

## Infrastructure as Code
- [ ] Create terraform directory structure for infrastructure provisioning
- [ ] Write terraform configurations for provisioning cloud resources

## Configuration Management
- [ ] Create ansible directory structure for Kubernetes installation
- [ ] Write ansible playbooks to install Kubernetes with kubeadm on provisioned resources
