# Auth Microservice — Full DevSecOps Pipeline

A production-grade authentication microservice demonstrating security-aware backend development, enterprise CI/CD pipelines, and cloud-native deployment patterns.

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| Framework | Gin |
| Auth | JWT (HS256) + OAuth2 (Google) |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| RBAC | Custom (admin / user / viewer) |
| Container | Docker (scratch image) |
| Orchestration | Kubernetes + Helm (dev / staging / prod) |
| CI/CD | GitHub Actions |
| SAST | gosec → SARIF → GitHub Security tab |
| Secrets scan | TruffleHog |
| Image scan | Trivy |
| Lint | golangci-lint |

## API Endpoints

```
POST   /api/v1/auth/register              Register new user
POST   /api/v1/auth/login                 Login → JWT pair
POST   /api/v1/auth/refresh               Rotate refresh token
POST   /api/v1/auth/logout                Revoke refresh token
GET    /api/v1/auth/oauth/google          Get Google OAuth URL
GET    /api/v1/auth/oauth/google/callback OAuth callback

GET    /api/v1/me                         Current user profile  [auth]
GET    /api/v1/me/permissions             Current user role     [auth]

GET    /api/v1/admin/users                List all users        [admin]
GET    /api/v1/admin/users/:id            Get user by ID        [admin]
PATCH  /api/v1/admin/users/:id/role       Change user role      [admin]
DELETE /api/v1/admin/users/:id            Delete user           [admin]
GET    /api/v1/admin/audit-logs           Audit log stream      [admin]

GET    /health                            Liveness probe
```

## RBAC Permissions

| Permission | admin | user | viewer |
|---|---|---|---|
| users:read | ✓ | ✓ | ✓ |
| users:write | ✓ | | |
| users:delete | ✓ | | |
| logs:read | ✓ | | |
| roles:manage | ✓ | | |

## Quick Start

```bash
# 1. Clone & setup
git clone https://github.com/likhithnagaraj79/Auth-microservice-with-full-DevSecOps-pipeline.git
cd Auth-microservice-with-full-DevSecOps-pipeline

# 2. Configure (edit secrets!)
cp .env.example .env

# 3. Start infra + app
docker-compose up -d

# 4. Open frontend
open frontend/public/index.html
```

## DevSecOps Pipeline

```
push → lint (golangci-lint)
     → SAST (gosec → SARIF)
     → secrets scan (TruffleHog)
     → unit tests (race detector + coverage)
     → Docker build (multi-stage, scratch base)
     → Trivy image scan (HIGH/CRITICAL → fail)
     → Helm deploy to K8s (prod, main branch only)
```

## Helm Deployment

```bash
# Dev
helm upgrade --install auth-service ./deployments/helm/auth-service \
  --values ./deployments/helm/auth-service/values/dev.yaml

# Staging
helm upgrade --install auth-service ./deployments/helm/auth-service \
  --values ./deployments/helm/auth-service/values/staging.yaml

# Production
helm upgrade --install auth-service ./deployments/helm/auth-service \
  --values ./deployments/helm/auth-service/values/prod.yaml \
  --set secrets.JWT_ACCESS_SECRET="$JWT_ACCESS_SECRET"
```

## Integration (API Gateway Pattern)

Other services authenticate by forwarding the `Authorization: Bearer <token>` header to this service's `/api/v1/me` endpoint. A 200 response confirms the identity; the JSON body contains the user's role for local RBAC decisions.

## Security Decisions

- **bcrypt cost 12** — balances brute-force resistance vs. latency
- **Refresh token rotation** — each refresh issues a new token and revokes the old one
- **Scratch Docker base** — zero OS attack surface, no shell
- **Read-only root filesystem** — enforced in Helm securityContext
- **Non-root UID 65534** — nobody user in pod spec
- **SARIF upload** — gosec and Trivy findings surface in GitHub Security tab

## Running Tests

```bash
go test -v -race ./...
```
