# Simple Vault Helm Chart

This Helm chart deploys the Simple Vault application on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- MongoDB (either external or using the included MongoDB deployment)

## Installation

### Quick Start

```bash
# Add your image registry if needed
# Build and push images first:
# docker build -t your-registry/simple-vault-api:latest ./api
# docker build -t your-registry/simple-vault-ui:latest ./ui

# Install with default values
helm install simple-vault ./helm/simple-vault

# Install with custom values
helm install simple-vault ./helm/simple-vault -f my-values.yaml

# Install with external MongoDB
helm install simple-vault ./helm/simple-vault \
  --set api.env.MONGODB_URI="mongodb://user:pass@mongodb-host:27017/vault?authSource=admin" \
  --set mongodb.enabled=false
```

### Using External MongoDB

If you're using an external MongoDB instance:

```yaml
# values-external-mongodb.yaml
mongodb:
  enabled: false

api:
  env:
    MONGODB_URI: "mongodb://user:password@mongodb-host:27017/vault?authSource=admin"
    # OR use individual components:
    # DB_HOST: "mongodb-host"
    # DB_PORT: "27017"
    # DB_USER: "user"
    # DB_PASSWORD: "password"
    # DB_NAME: "vault"
    AUTH_SERVICE_URL: "http://auth-service:8083"
```

Then install:
```bash
helm install simple-vault ./helm/simple-vault -f values-external-mongodb.yaml
```

## Configuration

The following table lists the configurable parameters and their default values:

### API Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `api.enabled` | Enable API deployment | `true` |
| `api.image.repository` | API image repository | `simple-vault-api` |
| `api.image.tag` | API image tag | `latest` |
| `api.replicaCount` | Number of API replicas | `2` |
| `api.service.type` | API service type | `ClusterIP` |
| `api.service.port` | API service port | `8080` |
| `api.env.MONGODB_URI` | MongoDB connection URI (takes precedence) | `""` |
| `api.env.DB_HOST` | MongoDB host | `mongodb` |
| `api.env.DB_PORT` | MongoDB port | `27017` |
| `api.env.DB_USER` | MongoDB username | `vault` |
| `api.env.DB_PASSWORD` | MongoDB password | `vault` |
| `api.env.DB_NAME` | MongoDB database name | `vault` |
| `api.env.AUTH_SERVICE_URL` | Auth service URL | `http://192.168.1.4:8083` |
| `api.resources` | API resource limits and requests | See values.yaml |

### UI Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ui.enabled` | Enable UI deployment | `true` |
| `ui.image.repository` | UI image repository | `simple-vault-ui` |
| `ui.image.tag` | UI image tag | `latest` |
| `ui.replicaCount` | Number of UI replicas | `2` |
| `ui.service.type` | UI service type | `ClusterIP` |
| `ui.service.port` | UI service port | `80` |
| `ui.resources` | UI resource limits and requests | See values.yaml |

### MongoDB Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `mongodb.enabled` | Enable MongoDB deployment | `false` |
| `mongodb.image.repository` | MongoDB image repository | `mongo` |
| `mongodb.image.tag` | MongoDB image tag | `7` |
| `mongodb.auth.enabled` | Enable MongoDB authentication | `true` |
| `mongodb.auth.rootUsername` | MongoDB root username | `vault` |
| `mongodb.auth.rootPassword` | MongoDB root password | `vault` |
| `mongodb.persistence.enabled` | Enable persistent storage | `true` |
| `mongodb.persistence.size` | Persistent volume size | `10Gi` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | Ingress TLS configuration | `[]` |

## Examples

### Example 1: Deploy with Internal MongoDB

```bash
helm install simple-vault ./helm/simple-vault \
  --set mongodb.enabled=true \
  --set mongodb.auth.rootPassword=mySecurePassword
```

### Example 2: Deploy with External MongoDB and Custom Auth Service

```bash
helm install simple-vault ./helm/simple-vault \
  --set mongodb.enabled=false \
  --set api.env.MONGODB_URI="mongodb://user:pass@external-mongodb:27017/vault" \
  --set api.env.AUTH_SERVICE_URL="http://auth-service.default.svc.cluster.local:8083"
```

### Example 3: Deploy with Ingress

```yaml
# values-ingress.yaml
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: vault.example.com
      paths:
        - path: /
          pathType: Prefix
          service: ui
        - path: /api
          pathType: Prefix
          service: api
  tls:
    - secretName: vault-tls
      hosts:
        - vault.example.com
```

```bash
helm install simple-vault ./helm/simple-vault -f values-ingress.yaml
```

### Example 4: Production Configuration

```yaml
# values-production.yaml
api:
  replicaCount: 3
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi

ui:
  replicaCount: 3
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi

mongodb:
  enabled: false  # Use external managed MongoDB

api:
  env:
    MONGODB_URI: "mongodb://prod-user:prod-pass@prod-mongodb:27017/vault?authSource=admin"
    AUTH_SERVICE_URL: "http://auth-service.production.svc.cluster.local:8083"

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: vault.production.com
      paths:
        - path: /
          pathType: Prefix
          service: ui
        - path: /api
          pathType: Prefix
          service: api
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade simple-vault ./helm/simple-vault -f my-values.yaml

# Upgrade with set flags
helm upgrade simple-vault ./helm/simple-vault \
  --set api.image.tag=v1.1.0 \
  --set ui.image.tag=v1.1.0
```

## Uninstalling

```bash
helm uninstall simple-vault
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=simple-vault
```

### Check Logs

```bash
# API logs
kubectl logs -l app.kubernetes.io/component=api

# UI logs
kubectl logs -l app.kubernetes.io/component=ui
```

### Check Services

```bash
kubectl get svc -l app.kubernetes.io/name=simple-vault
```

### Port Forward for Testing

```bash
# Forward API
kubectl port-forward svc/simple-vault-api 8080:8080

# Forward UI
kubectl port-forward svc/simple-vault-ui 3000:80
```

## Notes

- The UI build-time environment variable `VITE_API_URL` should be set during the Docker build process. The Helm chart assumes the image is already built with the correct API URL.
- If using an external MongoDB, ensure the `vcluster` database and `users` collection exist, as the API queries this for user authentication.
- The auth service URL should be accessible from within the Kubernetes cluster.

