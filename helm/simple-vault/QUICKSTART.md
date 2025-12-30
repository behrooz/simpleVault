# Quick Start Guide

## Prerequisites

1. Build and push your Docker images:
```bash
# Build API image
cd /home/behrooz/Projects/simple-vault/simpleVault/api
docker build -t your-registry/simple-vault-api:latest .

# Build UI image (make sure VITE_API_URL is set correctly in build)
cd /home/behrooz/Projects/simple-vault/simpleVault/ui
docker build -t your-registry/simple-vault-ui:latest .

# Push to your registry
docker push your-registry/simple-vault-api:latest
docker push your-registry/simple-vault-ui:latest
```

## Installation Options

### Option 1: With External MongoDB (Recommended for Production)

```bash
helm install simple-vault ./helm/simple-vault \
  --set api.image.repository=your-registry/simple-vault-api \
  --set ui.image.repository=your-registry/simple-vault-ui \
  --set mongodb.enabled=false \
  --set api.env.MONGODB_URI="mongodb://user:password@mongodb-host:27017/vault?authSource=admin" \
  --set api.env.AUTH_SERVICE_URL="http://auth-service:8083"
```

### Option 2: With Internal MongoDB (For Testing)

```bash
helm install simple-vault ./helm/simple-vault \
  --set api.image.repository=your-registry/simple-vault-api \
  --set ui.image.repository=your-registry/simple-vault-ui \
  --set mongodb.enabled=true \
  --set mongodb.auth.rootPassword=your-secure-password \
  --set api.env.AUTH_SERVICE_URL="http://auth-service:8083"
```

### Option 3: Using Values File

Create `my-values.yaml`:
```yaml
api:
  image:
    repository: your-registry/simple-vault-api
    tag: latest
  env:
    MONGODB_URI: "mongodb://user:password@mongodb-host:27017/vault?authSource=admin"
    AUTH_SERVICE_URL: "http://auth-service:8083"

ui:
  image:
    repository: your-registry/simple-vault-ui
    tag: latest

mongodb:
  enabled: false
```

Then install:
```bash
helm install simple-vault ./helm/simple-vault -f my-values.yaml
```

## Verify Installation

```bash
# Check pods
kubectl get pods -l app.kubernetes.io/name=simple-vault

# Check services
kubectl get svc -l app.kubernetes.io/name=simple-vault

# Check logs
kubectl logs -l app.kubernetes.io/component=api
kubectl logs -l app.kubernetes.io/component=ui
```

## Access the Application

### Port Forward (for testing)

```bash
# Forward API
kubectl port-forward svc/simple-vault-api 8080:8080

# Forward UI
kubectl port-forward svc/simple-vault-ui 3000:80
```

Then access:
- UI: http://localhost:3000
- API: http://localhost:8080/api/v1/secrets

### Using Ingress

Enable ingress in your values file and configure your domain.

## Upgrade

```bash
helm upgrade simple-vault ./helm/simple-vault -f my-values.yaml
```

## Uninstall

```bash
helm uninstall simple-vault
```

