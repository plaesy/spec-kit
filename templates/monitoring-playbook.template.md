# Monitoring & Observability Playbook Template

## Service Overview
- **Service Name**: [Service/Application Name]
- **Version**: [Current Version]
- **Owner Team**: [Team Name]
- **Contact**: [On-call rotation/contact info]
- **Business Criticality**: [Critical/High/Medium/Low]

## Architecture Context
- **Dependencies**: [List of upstream/downstream services]
- **Data Stores**: [Databases, caches, message queues]
- **External APIs**: [Third-party integrations]
- **Infrastructure**: [Cloud provider, regions, compute resources]

## Service Level Objectives (SLOs)

### Availability SLO
- **Target**: 99.9% uptime (8.76 hours downtime/year)
- **Measurement Window**: 30 days rolling
- **Error Budget**: 0.1% (43.2 minutes/month)

### Latency SLO
- **Target**: 95th percentile < 200ms
- **Measurement Window**: 5 minutes
- **Endpoints**: [List critical endpoints]

### Throughput SLO
- **Target**: Support X requests/second sustained
- **Peak Capacity**: Y requests/second
- **Measurement**: Average over 1-minute windows

## Key Metrics & Dashboards

### Golden Signals
1. **Latency**: Response time distribution
   - Metric: `http_request_duration_seconds`
   - Dashboard: [Link to latency dashboard]

2. **Traffic**: Request rate
   - Metric: `http_requests_total`
   - Dashboard: [Link to traffic dashboard]

3. **Errors**: Error rate and types
   - Metric: `http_requests_total{status=~"5.."}`
   - Dashboard: [Link to errors dashboard]

4. **Saturation**: Resource utilization
   - Metrics: CPU, memory, disk, network
   - Dashboard: [Link to infrastructure dashboard]

### Business Metrics
- **User Signups**: Rate and conversion
- **Transaction Volume**: Revenue impact metrics
- **Feature Adoption**: Usage of key features
- **Data Quality**: Completeness and accuracy metrics

### Custom Application Metrics
- **Queue Depth**: Background job processing
- **Cache Hit Rate**: Performance optimization
- **Database Connections**: Resource utilization
- **External API Latency**: Dependency health

## Alerting Rules

### Critical Alerts (Page immediately)
```yaml
- alert: ServiceDown
  expr: up{job="[service-name]"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Service [service-name] is down"

- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "High error rate: {{ $value }}% errors"

- alert: HighLatency
  expr: histogram_quantile(0.95, http_request_duration_seconds) > 0.5
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High latency: {{ $value }}s at 95th percentile"
```

### Warning Alerts (Notify but don't page)
```yaml
- alert: ErrorBudgetExhaustion
  expr: error_budget_remaining < 0.25
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Error budget 75% consumed"

- alert: HighCPUUsage
  expr: cpu_usage_percent > 80
  for: 15m
  labels:
    severity: warning
  annotations:
    summary: "High CPU usage: {{ $value }}%"
```

## Runbook Procedures

### Alert Response Matrix
| Alert | Severity | Response Time | Escalation |
|-------|----------|---------------|------------|
| ServiceDown | Critical | Immediate | On-call → Team Lead → Manager |
| HighErrorRate | Critical | 5 minutes | On-call → Senior Engineer |
| HighLatency | Critical | 10 minutes | On-call → Performance Team |
| ErrorBudget | Warning | 30 minutes | Team notification |

### Common Incident Procedures

#### Service Down Response
1. **Check service health endpoint**: `curl [health-check-url]`
2. **Verify infrastructure**: Check cloud provider status
3. **Check dependencies**: Verify database, external APIs
4. **Review recent deployments**: Check for recent changes
5. **Scale horizontally**: Add more instances if needed
6. **Rollback**: If deployment-related, initiate rollback

#### High Error Rate Response
1. **Identify error patterns**: Check error logs and types
2. **Check external dependencies**: Verify third-party APIs
3. **Resource constraints**: Monitor CPU, memory, connections
4. **Recent changes**: Review recent code/config deployments
5. **Circuit breaker**: Enable if external dependency issues

#### Performance Degradation
1. **Resource utilization**: Check CPU, memory, disk I/O
2. **Database performance**: Query execution times, locks
3. **Cache efficiency**: Hit rates and cache evictions
4. **Network latency**: Inter-service communication
5. **Auto-scaling**: Trigger horizontal scaling if needed

## Log Analysis

### Structured Logging Format
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "[service-name]",
  "version": "[version]",
  "trace_id": "[uuid]",
  "user_id": "[user-id]",
  "endpoint": "/api/users",
  "method": "GET",
  "status_code": 200,
  "duration_ms": 150,
  "message": "User profile retrieved successfully"
}
```

### Key Log Queries
```bash
# Error rate by endpoint
grep "ERROR" logs.json | jq -r '.endpoint' | sort | uniq -c

# Slow requests (>1s)
grep "duration_ms" logs.json | jq 'select(.duration_ms > 1000)'

# User error patterns
grep "user_id" logs.json | jq 'select(.level == "ERROR")' | jq -r '.user_id' | sort | uniq -c
```

### Log Retention Policy
- **Application Logs**: 30 days hot, 90 days cold storage
- **Access Logs**: 90 days hot, 1 year archive
- **Error Logs**: 90 days hot, 1 year archive
- **Audit Logs**: 7 years (compliance requirement)

## Performance Baselines

### Response Time Baselines
| Endpoint | 50th %ile | 95th %ile | 99th %ile |
|----------|-----------|-----------|-----------|
| GET /api/users | 50ms | 150ms | 300ms |
| POST /api/users | 100ms | 250ms | 500ms |
| GET /api/users/{id} | 25ms | 75ms | 150ms |

### Throughput Baselines
- **Normal Load**: 1,000 RPS sustained
- **Peak Load**: 5,000 RPS for 10 minutes
- **Maximum Tested**: 10,000 RPS burst capacity

### Resource Baselines
- **CPU Usage**: 40-60% normal, 80% peak
- **Memory Usage**: 60-70% normal, 85% peak
- **Database Connections**: 20-30 normal, 80 peak
- **Disk I/O**: <70% utilization

## Capacity Planning

### Growth Projections
- **Traffic Growth**: 20% per quarter
- **Data Growth**: 15% per quarter
- **User Growth**: 25% per quarter

### Scaling Triggers
- **CPU > 70%** for 10 minutes → Add instances
- **Memory > 80%** for 5 minutes → Add instances
- **Queue depth > 1000** → Add workers
- **DB connections > 70%** → Scale read replicas

### Infrastructure Scaling Plan
1. **Horizontal Scaling**: Auto-scaling groups 2-10 instances
2. **Database Scaling**: Read replicas, connection pooling
3. **Cache Scaling**: Redis cluster mode, memory optimization
4. **CDN Scaling**: Geographic distribution, cache policies

## Security Monitoring

### Security Metrics
- **Failed Authentication Rate**: <1% of total attempts
- **Suspicious Activity**: Unusual access patterns
- **API Rate Limiting**: Threshold breaches
- **Data Access Patterns**: Unusual data queries

### Security Alerts
- **Brute Force Attacks**: >10 failed logins/minute
- **Data Exfiltration**: Large data downloads
- **Privilege Escalation**: Unauthorized access attempts
- **Configuration Changes**: Security setting modifications

## Data Quality Monitoring

### Data Freshness
- **ETL Pipeline Lag**: <15 minutes
- **Real-time Streams**: <1 minute delay
- **Batch Processing**: Complete within SLA windows

### Data Accuracy
- **Validation Rules**: Check business logic constraints
- **Referential Integrity**: Foreign key constraints
- **Schema Compliance**: Data type and format validation
- **Duplicate Detection**: Identify and flag duplicates

## Cost Monitoring

### Cost Metrics
- **Compute Costs**: Per service, per environment
- **Storage Costs**: Database, file storage, backups
- **Network Costs**: Data transfer, CDN usage
- **Third-party APIs**: Usage-based pricing

### Cost Optimization
- **Right-sizing**: Match resources to actual usage
- **Reserved Instances**: Long-term capacity planning
- **Automated Scaling**: Scale down during low usage
- **Data Lifecycle**: Archive or delete old data

## Incident Communication

### Communication Channels
- **Critical Incidents**: PagerDuty → Slack #incidents
- **Status Updates**: Status page, stakeholder emails
- **Post-mortem**: Dedicated Slack channel, documentation

### Escalation Matrix
1. **L1 Response**: On-call engineer (0-15 minutes)
2. **L2 Escalation**: Senior engineer (15-30 minutes)
3. **L3 Escalation**: Team lead (30-60 minutes)
4. **Management**: Director level (60+ minutes)

### Status Page Updates
- **Investigating**: Within 5 minutes of detection
- **Identified**: When root cause is known
- **Monitoring**: During fix implementation
- **Resolved**: When service fully restored

## Maintenance & Updates

### Planned Maintenance Windows
- **Schedule**: Second Tuesday of month, 2-4 AM UTC
- **Notification**: 7 days advance notice
- **Duration**: Maximum 2 hours
- **Rollback Plan**: 30-minute rollback window

### Monitoring Health Checks
- **Dashboard Review**: Weekly team review
- **Alert Tuning**: Monthly false positive analysis
- **SLO Review**: Quarterly business alignment
- **Runbook Updates**: After each incident

### Tool Maintenance
- **Metric Retention**: Configure appropriate retention periods
- **Dashboard Cleanup**: Remove unused dashboards
- **Alert Cleanup**: Remove obsolete alerts
- **Access Review**: Quarterly permission audit

## Documentation Links
- **Architecture Diagrams**: [Link to technical documentation]
- **API Documentation**: [Link to API specs]
- **Deployment Guide**: [Link to deployment procedures]
- **Security Procedures**: [Link to security documentation]
- **Business Context**: [Link to business requirements]

## Emergency Contacts
- **On-call Rotation**: [PagerDuty/OpsGenie schedule]
- **Team Lead**: [Contact information]
- **Product Owner**: [Contact information]
- **Infrastructure Team**: [Contact information]
- **Security Team**: [Contact information]

---
**Document Version**: 1.0
**Last Updated**: [Date]
**Next Review**: [Date + 3 months]
**Owner**: [Team Name]