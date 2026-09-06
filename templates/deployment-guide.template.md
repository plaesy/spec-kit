# Deployment Guide Template

## Deployment Guide for: [Application/Service Name]

**Version**: [Version Number]  
**Date**: [Date]  
**Environment**: [Development/Staging/Production]  
**Prepared By**: [DevOps Engineer Name]

---

## Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Prerequisites](#prerequisites)
3. [Environment Setup](#environment-setup)
4. [Pre-Deployment Checklist](#pre-deployment-checklist)
5. [Deployment Steps](#deployment-steps)
6. [Post-Deployment Verification](#post-deployment-verification)
7. [Rollback Procedures](#rollback-procedures)
8. [Monitoring and Alerting](#monitoring-and-alerting)
9. [Troubleshooting](#troubleshooting)
10. [Support and Contacts](#support-and-contacts)

---

## Deployment Overview

### Application Information
- **Application Name**: [Name]
- **Version**: [Version Number]
- **Git Commit**: [Commit Hash]
- **Build Number**: [Build Number]
- **Release Date**: [Date]

### Deployment Strategy
- **Strategy Type**: [Blue-Green/Rolling/Canary/Recreate]
- **Downtime Expected**: [Duration or "Zero downtime"]
- **Traffic Routing**: [Method for traffic management]
- **Rollback Time**: [Time to complete rollback]

### Architecture Overview
```
[Include architecture diagram showing deployment topology]

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │────│   App Servers   │────│    Database     │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Environment Details
- **Target Environment**: [Production/Staging/Development]
- **Infrastructure**: [Cloud Provider/On-premises]
- **Deployment Tool**: [Tool name and version]
- **Container Runtime**: [Docker/Kubernetes/etc.]

---

## Prerequisites

### System Requirements
**Minimum Hardware Requirements**:
- CPU: [Requirements]
- Memory: [Requirements]
- Storage: [Requirements]
- Network: [Bandwidth requirements]

**Software Dependencies**:
- Operating System: [Version requirements]
- Runtime Environment: [Version requirements]
- Database: [Type and version]
- Required Services: [List of dependent services]

### Access Requirements
**Personnel Access**:
- [ ] Deployment team has necessary server access
- [ ] Database administrator available for schema changes
- [ ] Network administrator available for configuration changes
- [ ] Security team approval for production deployment

**Service Accounts**:
- [ ] Application service account configured
- [ ] Database connection credentials available
- [ ] External service API keys and certificates ready
- [ ] SSL certificates valid and configured

### External Dependencies
**Third-Party Services**:
- [ ] External API services available and tested
- [ ] Payment gateway integration tested
- [ ] Email service provider configured
- [ ] CDN and static asset hosting ready

**Network Dependencies**:
- [ ] Firewall rules configured
- [ ] DNS records updated
- [ ] Load balancer configuration validated
- [ ] SSL/TLS certificates installed

---

## Environment Setup

### Infrastructure Configuration

#### Server Configuration
```bash
# Server specifications
Servers: [Number] x [Instance Type]
CPU: [Specifications]
Memory: [Amount]
Storage: [Type and amount]
Operating System: [OS and version]
```

#### Network Configuration
```bash
# Network settings
VPC/Network: [Network ID]
Subnets: [Subnet configurations]
Security Groups: [Security group IDs]
Load Balancer: [Load balancer configuration]
```

#### Database Configuration
```sql
-- Database setup
Database Type: [PostgreSQL/MySQL/MongoDB/etc.]
Version: [Version number]
Instance Size: [Size specifications]
Storage: [Storage configuration]
Backup Configuration: [Backup strategy]
```

### Container Configuration

#### Dockerfile
```dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000

CMD ["npm", "start"]
```

#### Docker Compose (for development/staging)
```yaml
version: '3.8'

services:
  app:
    image: your-app:${VERSION}
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      - database
      
  database:
    image: postgres:15
    environment:
      - POSTGRES_DB=${DB_NAME}
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - db_data:/var/lib/postgresql/data

volumes:
  db_data:
```

#### Kubernetes Manifests
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-app
  labels:
    app: your-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: your-app
  template:
    metadata:
      labels:
        app: your-app
    spec:
      containers:
      - name: your-app
        image: your-registry/your-app:${VERSION}
        ports:
        - containerPort: 3000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 30
```

---

## Pre-Deployment Checklist

### Code and Build Verification
- [ ] **Code Review**: All code changes reviewed and approved
- [ ] **Build Status**: Build passes all automated tests
- [ ] **Security Scan**: Security vulnerabilities addressed
- [ ] **Performance Test**: Performance benchmarks met
- [ ] **Integration Test**: All integration tests pass

### Environment Preparation
- [ ] **Environment Health**: Target environment healthy and ready
- [ ] **Database Backup**: Current database backup completed
- [ ] **Configuration**: Environment variables and secrets configured
- [ ] **Dependencies**: All dependencies available and compatible
- [ ] **Monitoring**: Monitoring and alerting systems active

### Communication and Approval
- [ ] **Stakeholder Approval**: Deployment approved by product owner
- [ ] **Change Management**: Change request approved (if required)
- [ ] **Team Notification**: Development and operations teams notified
- [ ] **Customer Communication**: Customer-facing communications prepared
- [ ] **Rollback Plan**: Rollback procedures documented and tested

### Maintenance Window
- [ ] **Maintenance Window**: Scheduled maintenance window (if required)
- [ ] **User Notification**: Users notified of potential downtime
- [ ] **Support Team**: Support team prepared for post-deployment issues
- [ ] **Documentation**: Deployment documentation updated

---

## Deployment Steps

### Step 1: Pre-Deployment Setup
```bash
# 1. Connect to deployment environment
ssh deploy@production-server

# 2. Navigate to deployment directory
cd /opt/deployments

# 3. Create deployment directory
mkdir deployment-$(date +%Y%m%d-%H%M%S)
cd deployment-$(date +%Y%m%d-%H%M%S)

# 4. Download deployment artifacts
wget https://releases.yourcompany.com/app-v1.2.3.tar.gz
tar -xzf app-v1.2.3.tar.gz
```

### Step 2: Database Migration (if required)
```bash
# 1. Backup current database
pg_dump production_db > backup-$(date +%Y%m%d-%H%M%S).sql

# 2. Run database migrations
./migrate.sh up

# 3. Verify migration success
./migrate.sh status
```

### Step 3: Application Deployment

#### Traditional Deployment
```bash
# 1. Stop current application
sudo systemctl stop your-app

# 2. Update application files
sudo cp -r app/* /opt/your-app/

# 3. Update configuration
sudo cp config/production.conf /opt/your-app/config/

# 4. Install/update dependencies
cd /opt/your-app && npm install --production

# 5. Start application
sudo systemctl start your-app

# 6. Verify application status
sudo systemctl status your-app
```

#### Docker Deployment
```bash
# 1. Pull new Docker image
docker pull your-registry/your-app:v1.2.3

# 2. Stop current container
docker stop your-app-current

# 3. Run new container
docker run -d --name your-app-new \
  -p 3000:3000 \
  -e DATABASE_URL=$DATABASE_URL \
  your-registry/your-app:v1.2.3

# 4. Health check
curl http://localhost:3000/health

# 5. Switch traffic to new container
# (Update load balancer configuration)

# 6. Remove old container
docker rm your-app-current
docker rename your-app-new your-app-current
```

#### Kubernetes Deployment
```bash
# 1. Update image version in deployment
kubectl set image deployment/your-app \
  your-app=your-registry/your-app:v1.2.3

# 2. Monitor rollout status
kubectl rollout status deployment/your-app

# 3. Verify pods are running
kubectl get pods -l app=your-app

# 4. Check application logs
kubectl logs -l app=your-app --tail=100
```

### Step 4: Configuration Update
```bash
# 1. Update environment variables
echo "APP_VERSION=v1.2.3" >> /etc/environment

# 2. Update configuration files
sudo cp config/production.conf /etc/your-app/

# 3. Reload configuration
sudo systemctl reload your-app

# 4. Verify configuration
curl http://localhost:3000/config
```

### Step 5: Traffic Routing
```bash
# For Blue-Green Deployment
# 1. Route traffic to new environment
aws elbv2 modify-target-group --target-group-arn $TARGET_GROUP_ARN \
  --targets Id=new-instance-id,Port=3000

# 2. Monitor traffic and metrics
# (Check monitoring dashboards)

# 3. Remove old environment from load balancer
aws elbv2 deregister-targets --target-group-arn $TARGET_GROUP_ARN \
  --targets Id=old-instance-id,Port=3000
```

---

## Post-Deployment Verification

### Health Checks
```bash
# 1. Application health endpoint
curl -f http://your-app.com/health
# Expected: HTTP 200 with health status

# 2. Database connectivity
curl -f http://your-app.com/health/database
# Expected: HTTP 200 with database status

# 3. External service connectivity
curl -f http://your-app.com/health/external
# Expected: HTTP 200 with external service status

# 4. API functionality test
curl -X POST http://your-app.com/api/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
# Expected: HTTP 200 with valid response
```

### Smoke Tests
- [ ] **User Authentication**: Login functionality works
- [ ] **Core Features**: Primary user workflows function correctly
- [ ] **API Endpoints**: Critical API endpoints respond correctly
- [ ] **Database Operations**: Read/write operations successful
- [ ] **External Integrations**: Third-party services accessible

### Performance Verification
```bash
# 1. Response time check
curl -w "Total time: %{time_total}s\n" -o /dev/null -s http://your-app.com/

# 2. Load test (basic)
ab -n 100 -c 10 http://your-app.com/

# 3. Memory usage check
docker stats your-app-container

# 4. CPU usage monitoring
top -p $(pgrep your-app)
```

### Monitoring Dashboard Review
- [ ] **Application Metrics**: Response times, error rates, throughput
- [ ] **Infrastructure Metrics**: CPU, memory, disk usage
- [ ] **Business Metrics**: User activity, transaction volumes
- [ ] **Error Tracking**: Error logs and exception reporting
- [ ] **Security Monitoring**: Security events and alerts

---

## Rollback Procedures

### Rollback Decision Criteria
**Immediate Rollback Required**:
- Critical functionality broken
- High error rates (>5% of requests)
- Severe performance degradation (>50% slower)
- Security vulnerabilities exposed
- Data corruption detected

**Rollback Process Approval**:
- Product Owner approval required
- Technical Lead confirmation
- Operations team coordination

### Rollback Steps

#### Traditional Rollback
```bash
# 1. Stop current application
sudo systemctl stop your-app

# 2. Restore previous version
sudo cp -r /opt/your-app-backup/* /opt/your-app/

# 3. Restore database (if needed)
psql production_db < backup-previous.sql

# 4. Start application
sudo systemctl start your-app

# 5. Verify rollback success
curl http://localhost:3000/health
```

#### Docker Rollback
```bash
# 1. Find previous image
docker images your-registry/your-app

# 2. Run previous version
docker stop your-app-current
docker run -d --name your-app-rollback \
  -p 3000:3000 \
  your-registry/your-app:v1.2.2

# 3. Update load balancer
# (Point traffic to rollback container)
```

#### Kubernetes Rollback
```bash
# 1. Check rollout history
kubectl rollout history deployment/your-app

# 2. Rollback to previous version
kubectl rollout undo deployment/your-app

# 3. Monitor rollback
kubectl rollout status deployment/your-app

# 4. Verify pods
kubectl get pods -l app=your-app
```

### Post-Rollback Actions
- [ ] Verify system functionality
- [ ] Update monitoring dashboards
- [ ] Notify stakeholders of rollback
- [ ] Document rollback reason and lessons learned
- [ ] Plan fix for rolled-back issue

---

## Monitoring and Alerting

### Key Metrics to Monitor
**Application Metrics**:
- Response time (95th percentile < 200ms)
- Error rate (< 0.1%)
- Throughput (requests per second)
- Active users and sessions

**Infrastructure Metrics**:
- CPU utilization (< 80%)
- Memory usage (< 85%)
- Disk space (< 80% full)
- Network latency and packet loss

**Business Metrics**:
- Transaction success rate
- User registration/conversion rates
- Revenue impact metrics
- Feature adoption rates

### Alert Configuration
```yaml
# Example alert rules (Prometheus/AlertManager)
groups:
- name: application-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
    for: 2m
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} requests per second"

  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.2
    for: 5m
    annotations:
      summary: "High response time detected"

  - alert: DatabaseConnectionFailure
    expr: up{job="database"} == 0
    for: 1m
    annotations:
      summary: "Database connection failed"
```

### Monitoring Dashboard URLs
- **Application Dashboard**: [URL]
- **Infrastructure Dashboard**: [URL]
- **Business Metrics Dashboard**: [URL]
- **Error Tracking**: [URL]
- **Log Aggregation**: [URL]

---

## Troubleshooting

### Common Issues and Solutions

#### Application Won't Start
**Symptoms**: Service fails to start, error in logs
**Possible Causes**:
- Configuration file errors
- Missing environment variables
- Port conflicts
- Dependency issues

**Investigation Steps**:
```bash
# Check service status
sudo systemctl status your-app

# Check application logs
sudo journalctl -u your-app -f

# Check port availability
netstat -tulpn | grep :3000

# Verify configuration
your-app --config-check
```

#### High Response Times
**Symptoms**: Slow API responses, user complaints
**Possible Causes**:
- Database performance issues
- Memory leaks
- External service delays
- High traffic load

**Investigation Steps**:
```bash
# Check application performance
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:3000/api/test

# Monitor system resources
top
htop
iostat 1

# Check database performance
EXPLAIN ANALYZE SELECT * FROM slow_query;

# Review application logs for bottlenecks
tail -f /var/log/your-app/performance.log
```

#### Database Connection Issues
**Symptoms**: Database errors in logs, connection timeouts
**Possible Causes**:
- Database server down
- Network connectivity issues
- Connection pool exhausted
- Authentication failures

**Investigation Steps**:
```bash
# Test database connectivity
pg_isready -h database-host -p 5432

# Check database logs
sudo tail -f /var/log/postgresql/postgresql.log

# Verify connection settings
echo $DATABASE_URL

# Test manual connection
psql $DATABASE_URL
```

### Escalation Procedures
**Level 1 - Operations Team**:
- Monitor alerts and dashboards
- Perform initial troubleshooting
- Implement known fixes

**Level 2 - Development Team**:
- Code-related issues
- Complex debugging
- Performance optimization

**Level 3 - Architecture Team**:
- System design issues
- Major performance problems
- Scalability concerns

### Emergency Contacts
**On-Call Rotation**: [Link to on-call schedule]
**Escalation Matrix**:
- L1 Support: [Contact information]
- L2 Support: [Contact information]
- L3 Support: [Contact information]
- Management: [Contact information]

---

## Support and Contacts

### Deployment Team
**DevOps Lead**: [Name, Email, Phone]
**Operations Engineer**: [Name, Email, Phone]
**Database Administrator**: [Name, Email, Phone]
**Security Engineer**: [Name, Email, Phone]

### Development Team
**Technical Lead**: [Name, Email, Phone]
**Backend Developer**: [Name, Email, Phone]
**Frontend Developer**: [Name, Email, Phone]

### Business Team
**Product Owner**: [Name, Email, Phone]
**Project Manager**: [Name, Email, Phone]
**Business Analyst**: [Name, Email, Phone]

### Communication Channels
**Slack Channels**:
- #deployments - Deployment notifications
- #alerts - System alerts and monitoring
- #incidents - Incident response coordination

**Email Lists**:
- deployments@yourcompany.com
- on-call@yourcompany.com
- management@yourcompany.com

### Documentation Links
- **Runbook**: [Link to operational runbook]
- **Architecture Documentation**: [Link to architecture docs]
- **API Documentation**: [Link to API docs]
- **Monitoring Dashboards**: [Link to dashboards]
- **Incident Response Procedures**: [Link to incident procedures]

---

**Deployment Sign-off**

**DevOps Engineer**: _________________________ Date: _________

**Technical Lead**: __________________________ Date: _________

**Product Owner**: __________________________ Date: _________

**Deployment Completed**: ___________________ Date: _________
