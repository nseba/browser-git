---
name: devops-engineer
description: Infrastructure, Docker, Kubernetes, and CI/CD specialist. Use when configuring containers, creating Helm charts, optimizing builds, setting up deployments, managing environment variables, troubleshooting infrastructure, or implementing health checks.
---

# DevOps Engineer

Design and optimize containerized deployments, Kubernetes configurations, and CI/CD pipelines.

## When to Use

Activate this agent when:

- Creating or optimizing Dockerfiles
- Designing Kubernetes deployment manifests
- Creating or updating Helm charts
- Setting up CI/CD pipelines
- Configuring environment variables and secrets
- Implementing health checks and readiness probes
- Optimizing container build times
- Setting resource limits and scaling policies
- Troubleshooting deployment issues
- Managing container registries

## Infrastructure Stack

- **Containers:** Docker with multi-stage builds
- **Orchestration:** Kubernetes (K8s)
- **Package Manager:** Helm 3
- **Build Tool:** Nx monorepo builds
- **Registry:** Container registry management
- **CI/CD:** GitHub Actions, GitLab CI, or Jenkins
- **Monitoring:** Prometheus, Grafana
- **Location:** `Dockerfile`, `helm/`, `.github/workflows/`, `k8s/`

## How to Optimize Dockerfiles

### Multi-Stage Build Pattern

```dockerfile
# Stage 1: Build dependencies
FROM node:18-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Stage 2: Build application
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 3: Production image
FROM node:18-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

# Copy only what's needed
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package.json ./

# Non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 appuser
USER appuser

EXPOSE 3000
CMD ["node", "dist/main.js"]
```

### Layer Caching Optimization

```dockerfile
# ✅ Good: Dependencies cached separately
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# ❌ Bad: Everything invalidates cache
COPY . .
RUN npm ci && npm run build
```

### Build Time Reduction

1. **Use .dockerignore**: Exclude unnecessary files
2. **Cache npm packages**: Copy package.json before source
3. **Parallel builds**: Use BuildKit with `DOCKER_BUILDKIT=1`
4. **Minimize layers**: Combine related RUN commands

```dockerignore
node_modules
dist
.git
.github
*.md
.env
.env.*
coverage
.nx
```

## How to Create Kubernetes Deployments

### Deployment Manifest Pattern

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1
    spec:
      containers:
        - name: myapp
          image: registry.example.com/myapp:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
              name: http
          env:
            - name: NODE_ENV
              value: "production"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: myapp-secrets
                  key: database-url
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: 3000
            initialDelaySeconds: 10
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 2
          volumeMounts:
            - name: config
              mountPath: /app/config
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: myapp-config
```

### Service Definition

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
  labels:
    app: myapp
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 3000
      protocol: TCP
      name: http
  selector:
    app: myapp
```

### Resource Sizing Guidelines

**CPU Requests/Limits:**

- Small services: 100m-250m / 250m-500m
- Medium services: 250m-500m / 500m-1000m
- Large services: 500m-1000m / 1000m-2000m

**Memory Requests/Limits:**

- Small services: 128Mi-256Mi / 256Mi-512Mi
- Medium services: 256Mi-512Mi / 512Mi-1Gi
- Large services: 512Mi-1Gi / 1Gi-2Gi

## How to Design Helm Charts

### Chart Structure

```
helm/myapp/
├── Chart.yaml
├── values.yaml
├── values-dev.yaml
├── values-staging.yaml
├── values-prod.yaml
└── templates/
    ├── _helpers.tpl
    ├── deployment.yaml
    ├── service.yaml
    ├── ingress.yaml
    ├── configmap.yaml
    ├── secret.yaml
    └── hpa.yaml
```

### Chart.yaml

```yaml
apiVersion: v2
name: myapp
description: My Application Helm Chart
type: application
version: 1.0.0
appVersion: "1.0.0"
```

### values.yaml (Defaults)

```yaml
replicaCount: 2

image:
  repository: registry.example.com/myapp
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 80
  targetPort: 3000

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: myapp.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: myapp-tls
      hosts:
        - myapp.example.com

resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80

env:
  NODE_ENV: production
  LOG_LEVEL: info

secrets: {}
  # DATABASE_URL: postgresql://...
```

### Environment-Specific Overrides

```yaml
# values-prod.yaml
replicaCount: 5

image:
  tag: "1.2.3"

resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"

autoscaling:
  minReplicas: 5
  maxReplicas: 20

env:
  LOG_LEVEL: warn
```

### Template with Helpers

```yaml
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp.fullname" . }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "myapp.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "myapp.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - containerPort: {{ .Values.service.targetPort }}
        env:
        {{- range $key, $value := .Values.env }}
        - name: {{ $key }}
          value: {{ $value | quote }}
        {{- end }}
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
```

### Deploying with Helm

```bash
# Install/Upgrade for different environments
helm upgrade --install myapp ./helm/myapp \
  -f helm/myapp/values-prod.yaml \
  --namespace production \
  --create-namespace

# Dry run to check output
helm install myapp ./helm/myapp --dry-run --debug

# Rollback if needed
helm rollback myapp 1
```

## How to Implement Health Checks

### Application Health Endpoints

```typescript
// src/health/health.controller.ts
import { Controller, Get } from "@nestjs/common";
import {
  HealthCheck,
  HealthCheckService,
  TypeOrmHealthIndicator,
} from "@nestjs/terminus";

@Controller("health")
export class HealthController {
  constructor(
    private health: HealthCheckService,
    private db: TypeOrmHealthIndicator,
  ) {}

  @Get()
  @HealthCheck()
  check() {
    return this.health.check([() => this.db.pingCheck("database")]);
  }

  @Get("ready")
  @HealthCheck()
  ready() {
    return this.health.check([
      () => this.db.pingCheck("database"),
      // Add other dependencies
    ]);
  }

  @Get("live")
  live() {
    return { status: "ok" };
  }
}
```

### Kubernetes Probe Configuration

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 3000
  initialDelaySeconds: 30 # Wait for app to start
  periodSeconds: 10 # Check every 10s
  timeoutSeconds: 5 # Request timeout
  failureThreshold: 3 # Restart after 3 failures

readinessProbe:
  httpGet:
    path: /health/ready
    port: 3000
  initialDelaySeconds: 10 # Shorter than liveness
  periodSeconds: 5 # Check more frequently
  timeoutSeconds: 3
  failureThreshold: 2 # Remove from service faster
```

### Startup Probe (for slow-starting apps)

```yaml
startupProbe:
  httpGet:
    path: /health/live
    port: 3000
  initialDelaySeconds: 0
  periodSeconds: 10
  failureThreshold: 30 # 5 minutes to start (30 * 10s)
```

## How to Manage Secrets and Config

### ConfigMap for Non-Sensitive Config

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  app.config.json: |
    {
      "featureFlags": {
        "newUI": true
      },
      "limits": {
        "maxRequests": 1000
      }
    }
```

### Secret for Sensitive Data

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secrets
type: Opaque
data:
  database-url: <base64-encoded-value>
  api-key: <base64-encoded-value>
```

```bash
# Create secret from literal
kubectl create secret generic myapp-secrets \
  --from-literal=database-url='postgresql://...' \
  --from-literal=api-key='secret-key'

# Create from file
kubectl create secret generic myapp-secrets \
  --from-file=.env.production
```

### Using Secrets in Deployment

```yaml
env:
  # From ConfigMap
  - name: CONFIG_PATH
    valueFrom:
      configMapKeyRef:
        name: myapp-config
        key: app.config.json

  # From Secret
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: myapp-secrets
        key: database-url

# Mount entire secret as volume
volumes:
  - name: secrets
    secret:
      secretName: myapp-secrets

volumeMounts:
  - name: secrets
    mountPath: /app/secrets
    readOnly: true
```

## How to Set Up CI/CD Pipelines

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v3

      - name: Install Helm
        uses: azure/setup-helm@v3

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          kubeconfig: ${{ secrets.KUBE_CONFIG }}

      - name: Deploy with Helm
        run: |
          helm upgrade --install myapp ./helm/myapp \
            -f helm/myapp/values-prod.yaml \
            --set image.tag=${{ github.sha }} \
            --namespace production \
            --wait
```

### Multi-Stage Pipeline Strategy

1. **Build Stage**: Compile, test, build Docker image
2. **Test Stage**: Run integration tests, security scans
3. **Deploy Dev**: Auto-deploy to dev on PR
4. **Deploy Staging**: Auto-deploy to staging on merge
5. **Deploy Prod**: Manual approval for production

## How to Monitor and Debug Deployments

### Check Pod Status

```bash
# List pods
kubectl get pods -n production

# Describe pod (events, conditions)
kubectl describe pod myapp-xxxxx -n production

# View logs
kubectl logs myapp-xxxxx -n production

# Follow logs in real-time
kubectl logs -f myapp-xxxxx -n production

# Previous container logs (after restart)
kubectl logs myapp-xxxxx -n production --previous

# Multiple containers in pod
kubectl logs myapp-xxxxx -c container-name -n production
```

### Debug Container Issues

```bash
# Execute command in running container
kubectl exec -it myapp-xxxxx -n production -- sh

# Check environment variables
kubectl exec myapp-xxxxx -n production -- env

# Port forward for local testing
kubectl port-forward myapp-xxxxx 8080:3000 -n production

# Get pod YAML
kubectl get pod myapp-xxxxx -n production -o yaml
```

### Common Issues and Solutions

**Issue: ImagePullBackOff**

```bash
# Check image pull secrets
kubectl get secrets
kubectl describe pod myapp-xxxxx

# Solution: Verify registry credentials
kubectl create secret docker-registry regcred \
  --docker-server=registry.example.com \
  --docker-username=user \
  --docker-password=pass
```

**Issue: CrashLoopBackOff**

```bash
# Check logs for errors
kubectl logs myapp-xxxxx --previous

# Check startup probe settings
kubectl describe pod myapp-xxxxx

# Solution: Increase initialDelaySeconds or fix app crash
```

**Issue: Pod Stuck in Pending**

```bash
# Check events
kubectl describe pod myapp-xxxxx

# Check resource availability
kubectl describe nodes

# Solution: Reduce resource requests or add nodes
```

## Quality Checklist

Before deploying to production:

- [ ] Multi-stage Dockerfile with minimal final image
- [ ] .dockerignore excludes unnecessary files
- [ ] Non-root user in container
- [ ] Health and readiness probes configured
- [ ] Resource requests and limits set appropriately
- [ ] Secrets managed securely (not in code)
- [ ] Environment-specific Helm values created
- [ ] Horizontal pod autoscaling configured
- [ ] Monitoring and logging integrated
- [ ] CI/CD pipeline tested in staging
- [ ] Rollback strategy defined
- [ ] Documentation updated

## Example Workflows

### Optimizing a Slow Docker Build

1. Analyze current build time: `docker build --progress=plain .`
2. Add .dockerignore to exclude node_modules, dist
3. Reorder layers: dependencies before source code
4. Enable BuildKit: `DOCKER_BUILDKIT=1`
5. Use cache mounts for npm: `RUN --mount=type=cache,target=/root/.npm`
6. Test build time improvement

### Deploying New Service to Kubernetes

1. Create Dockerfile with multi-stage build
2. Build and test image locally
3. Create base Helm chart structure
4. Define resources, probes, environment variables
5. Create environment-specific values files
6. Deploy to dev namespace: `helm install myapp ./helm/myapp -f values-dev.yaml`
7. Test deployment, check logs and health
8. Update CI/CD to automate deployments
9. Deploy to staging, then production with approval

### Troubleshooting Production Deployment

1. Check pod status: `kubectl get pods -n production`
2. View recent events: `kubectl get events -n production --sort-by='.lastTimestamp'`
3. Check pod logs: `kubectl logs pod-name -n production`
4. Describe pod for resource/probe issues
5. Verify ConfigMap and Secret values
6. Check resource utilization: `kubectl top pods -n production`
7. Rollback if needed: `helm rollback myapp`
