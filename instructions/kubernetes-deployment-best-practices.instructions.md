---
applyTo: 'deployments/**/*.yaml, deployments/**/*.yml, k8s/**/*.yaml, k8s/**/*.yml'
description: 'Comprehensive best practices for deploying and managing applications on Kubernetes. Covers Pods, Deployments, Services, Ingress, ConfigMaps, Secrets, health checks, resource limits, scaling, and security contexts.'
---

# Kubernetes Deployment Best Practices

## Mission
Guide developers toward reliable, secure, efficient Kubernetes deployments - emphasize resilience, security, scalability.

## Core Kubernetes Concepts for Deployment

### 1. Pods
Smallest deployable unit - one running process instance.
- Run a single primary container (or tightly coupled sidecars)
- Define `resources` (requests/limits) for CPU/memory
- Implement `livenessProbe` and `readinessProbe`
- Avoid deploying Pods directly - use Deployments/StatefulSets

### 2. Deployments
Manages a set of identical Pods, handles rolling updates/rollbacks.
- Use for stateless applications
- Define `replicas`, `selector`, `template`
- Configure `strategy` (`rollingUpdate` with `maxSurge`/`maxUnavailable`)

**Example (Simple Deployment):**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment
  labels:
    app: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app-container
          image: my-repo/my-app:1.0.0
          ports:
            - containerPort: 8080
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
```

### 3. Services
Exposes an application running on a set of Pods as a network service.
- Provides stable network identity to Pods
- Choose `type` by exposure needs: ClusterIP (internal), NodePort, LoadBalancer (internet-facing in cloud), ExternalName
- Ensure `selector` matches Pod labels

### 4. Ingress
Manages external HTTP/HTTPS access into cluster services.
- Consolidates routing rules, manages TLS termination
- Configure for external web-app access
- Specify host, path, backend service

**Example (Ingress):**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
spec:
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app-service
                port:
                  number: 80
  tls:
    - hosts:
        - myapp.example.com
      secretName: my-app-tls-secret
```

## Configuration and Secrets Management

### 1. ConfigMaps
Non-sensitive key-value configuration data.
- Use for app config, env vars, CLI args; mount as files or inject as env vars
- **Caution**: not encrypted at rest - never store sensitive data here

### 2. Secrets
Store sensitive data securely.
- Use for API keys, passwords, DB credentials, TLS certs
- Encrypted at rest in etcd (if cluster configured for it)
- Mount as volumes (files), or inject as env vars with caution
- Production: integrate external secret managers (HashiCorp Vault, AWS/Azure Secrets Manager) via an External Secrets Operator

## Health Checks and Probes

### 1. Liveness Probe
Detects if a container is still running - failure triggers a restart.
- HTTP, TCP, or command-based probe
- Config: `initialDelaySeconds`, `periodSeconds`, `timeoutSeconds`, `failureThreshold`, `successThreshold`

### 2. Readiness Probe
Detects if a container is ready for traffic - failure removes the Pod from Service load balancing.
- HTTP, TCP, or command-based probe, ensuring dependencies are available
- Use to gracefully remove Pods during startup or temporary outages

## Resource Management

### 1. Resource Requests and Limits
Define CPU/memory requests+limits for every container.
- **Requests**: guaranteed minimum (for scheduling)
- **Limits**: hard maximum (prevents noisy neighbors/resource exhaustion)
- Set both for proper QoS class: `Guaranteed`, `Burstable`, `BestEffort`

### 2. Horizontal Pod Autoscaler (HPA)
Scales Pod replica count from CPU utilization or custom metrics.
- Recommend for stateless apps with fluctuating load
- Config: `minReplicas`, `maxReplicas`, `targetCPUUtilizationPercentage`

### 3. Vertical Pod Autoscaler (VPA)
Adjusts container CPU/memory requests/limits from usage history - recommend for optimizing individual Pod resource usage over time.

## Security Best Practices

### 1. Network Policies
Control Pod-to-Pod and Pod-to-external communication - deny by default, allow by exception, granular rules.

### 2. RBAC
Control who can do what in the cluster - define granular `Roles`/`ClusterRoles`, bind via `RoleBindings`/`ClusterRoleBindings`. Always apply least privilege.

### 3. Pod Security Context
- `runAsNonRoot: true`
- `allowPrivilegeEscalation: false`
- `readOnlyRootFilesystem: true` where possible
- Drop unneeded capabilities (`capabilities: drop: [ALL]`)

**Example (Pod Security Context):**
```yaml
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
    - name: my-app
      image: my-repo/my-app:1.0.0
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        capabilities:
          drop:
            - ALL
```

### 4. Image Security
- Trusted, minimal base images (distroless, alpine)
- Vulnerability scanning in CI (Trivy, Clair, Snyk)
- Image signing and verification

### 5. API Server Security
Strong auth (client certs, OIDC), enforced RBAC, API auditing enabled.

## Logging, Monitoring, and Observability

### 1. Centralized Logging
Apps log to `STDOUT`/`STDERR`; a logging agent (Fluentd, Logstash, Loki) ships to a central system (ELK, Splunk, Datadog).

### 2. Metrics Collection
Prometheus + `kube-state-metrics` + `node-exporter`; custom app-specific exporters; Grafana for visualization.

### 3. Alerting
Prometheus Alertmanager for rule-based alerts: high error rates, low resource availability, Pod restarts, unhealthy probes.

### 4. Distributed Tracing
OpenTelemetry or Jaeger/Zipkin for end-to-end request tracing across microservices.

## Deployment Strategies

### 1. Rolling Updates (Default)
Gradually replace old Pods with new ones. Default for Deployments - tune `maxSurge`/`maxUnavailable`. Minimal downtime.

### 2. Blue/Green Deployment
Run two identical environments, switch traffic completely. Zero-downtime releases; needs external LB/Ingress traffic-switching support.

### 3. Canary Deployment
Roll a new version to a small user subset before full rollout. Good for testing with real traffic; implement via Service Mesh (Istio, Linkerd) or traffic-splitting Ingress.

### 4. Rollback Strategy
Revert to a previous stable version quickly/safely - `kubectl rollout undo`; keep previous image versions available.

## Kubernetes Manifest Review Checklist

- [ ] Is `apiVersion` and `kind` correct for the resource?
- [ ] Is `metadata.name` descriptive and follows naming conventions?
- [ ] Are `labels` and `selectors` consistently used?
- [ ] Are `replicas` set appropriately for the workload?
- [ ] Are `resources` (requests/limits) defined for all containers?
- [ ] Are `livenessProbe` and `readinessProbe` correctly configured?
- [ ] Are sensitive configurations handled via Secrets (not ConfigMaps)?
- [ ] Is `readOnlyRootFilesystem: true` set where possible?
- [ ] Is `runAsNonRoot: true` and a non-root `runAsUser` defined?
- [ ] Are unnecessary `capabilities` dropped?
- [ ] Are `NetworkPolicies` considered for communication restrictions?
- [ ] Is RBAC configured with least privilege for ServiceAccounts?
- [ ] Are `ImagePullPolicy` and image tags (`:latest` avoided) correctly set?
- [ ] Is logging sent to `STDOUT`/`STDERR`?
- [ ] Are appropriate `nodeSelector` or `tolerations` used for scheduling?
- [ ] Is the `strategy` for rolling updates configured?
- [ ] Are `Deployment` events and Pod statuses monitored?

## Troubleshooting Common Kubernetes Issues

### 1. Pods Not Starting (Pending, CrashLoopBackOff)
- `kubectl describe pod <pod_name>` for events/errors
- Review logs: `kubectl logs <pod_name> -c <container_name>`
- Verify resource requests/limits aren't too low
- Check image pull errors (typo, repo access)
- Ensure required ConfigMaps/Secrets are mounted/accessible

### 2. Pods Not Ready (Service Unavailable)
- Check `readinessProbe` config
- Verify the app listens on the expected port
- `kubectl describe service <service_name>` to confirm endpoints connected

### 3. Service Not Accessible
- Verify Service `selector` matches Pod labels
- Check Service `type` (ClusterIP internal, LoadBalancer external)
- For Ingress: check controller logs and Ingress resource rules
- Review `NetworkPolicies` possibly blocking traffic

### 4. Resource Exhaustion (OOMKilled)
- Increase `memory.limits`
- Optimize app memory usage
- Use VPA to recommend optimal limits

### 5. Performance Issues
- Monitor CPU/memory: `kubectl top pod` or Prometheus
- Check logs for slow queries/operations
- Analyze distributed traces for bottlenecks
- Review database performance

## Conclusion
Follow these Pod/Deployment/Service/Ingress/config/security/observability guidelines to build resilient, scalable, secure cloud-native apps on Kubernetes - continuously monitor, troubleshoot, and refine deployments.

---

<!-- End of Kubernetes Deployment Best Practices Instructions -->
