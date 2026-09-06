# Spec-Kit Cloud Integration README
# Comprehensive deployment templates for AWS, Azure, and Google Cloud Platform

## Overview

This directory contains Infrastructure as Code (IaC) templates for deploying Spec-Kit applications with constitutional compliance monitoring across major cloud platforms:

- **AWS**: Terraform + EKS + ECR + RDS + ElastiCache
- **Azure**: ARM Templates + AKS + ACR + PostgreSQL + Redis Cache  
- **Google Cloud**: Terraform + GKE + Artifact Registry + Cloud SQL + Memorystore

All deployments include:
- 🏛️ **Constitutional compliance monitoring**
- 📊 **Performance monitoring dashboards**
- 🔒 **Security scanning and vulnerability management**
- 🚀 **CI/CD integration with GitHub Actions**
- 📈 **Auto-scaling and high availability**

## Quick Start

### AWS Deployment

1. **Prerequisites**
   ```bash
   # Install required tools
   brew install terraform aws-cli kubectl helm
   
   # Configure AWS credentials
   aws configure
   ```

2. **Deploy Infrastructure**
   ```bash
   # Navigate to AWS templates
   cd .github/templates/cloud/aws/terraform
   
   # Initialize Terraform
   terraform init
   
   # Deploy infrastructure
   terraform apply -var="project_name=my-spec-kit-app" -var="environment=staging"
   ```

3. **Deploy Application**
   ```bash
   # Use GitHub Actions workflow
   gh workflow run deploy-aws.yml \
     --field environment=staging \
     --field force_deploy=false
   ```

### Azure Deployment

1. **Prerequisites**
   ```bash
   # Install Azure CLI
   curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
   
   # Login to Azure
   az login
   ```

2. **Deploy Infrastructure**
   ```bash
   # Create resource group
   az group create --name spec-kit-staging --location eastus
   
   # Deploy ARM template
   az deployment group create \
     --resource-group spec-kit-staging \
     --template-file .github/templates/cloud/azure/arm-template.json \
     --parameters projectName=my-spec-kit-app \
                  environment=staging \
                  databasePassword=SecurePassword123!
   ```

### Google Cloud Deployment

1. **Prerequisites**
   ```bash
   # Install Google Cloud SDK
   curl https://sdk.cloud.google.com | bash
   
   # Initialize and authenticate
   gcloud init
   gcloud auth application-default login
   ```

2. **Deploy Infrastructure**
   ```bash
   # Navigate to GCP templates
   cd .github/templates/cloud/gcp/terraform
   
   # Set project ID
   export TF_VAR_project_id=my-gcp-project-id
   
   # Deploy infrastructure
   terraform init
   terraform apply -var="project_name=my-spec-kit-app" -var="environment=staging"
   ```

## Architecture Overview

### AWS Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                            AWS Cloud                            │
├─────────────────────────────────────────────────────────────────┤
│  VPC (10.0.0.0/16)                                            │
│  ├── Public Subnets (ALB)                                      │
│  ├── Private Subnets (EKS Nodes)                              │
│  └── Database Subnets (RDS, ElastiCache)                      │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │       EKS       │  │       RDS       │  │   ElastiCache   │ │
│  │   (Kubernetes)  │  │  (PostgreSQL)   │  │     (Redis)     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │      ECR        │  │       S3        │  │ Secrets Manager │ │
│  │  (Containers)   │  │    (Assets)     │  │   (Secrets)     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Constitutional Compliance Components

Each deployment includes:

1. **Constitutional Monitor Sidecar**
   - Real-time TDD compliance checking
   - Library-first architecture validation
   - CLI interface monitoring
   - Performance metrics collection

2. **Performance Dashboard**
   - Constitutional compliance score
   - System performance metrics
   - Dependency vulnerability tracking
   - Quality gate monitoring

3. **Security Scanning**
   - Automated dependency vulnerability scanning
   - Container image security scanning
   - Infrastructure compliance checking
   - Secrets management validation

## Configuration

### Environment Variables

All deployments support these constitutional compliance configurations:

```yaml
# Constitutional compliance settings
CONSTITUTIONAL_MONITORING_ENABLED: "true"
TDD_ENFORCEMENT: "true"
LIBRARY_FIRST_VALIDATION: "true"
CLI_INTERFACE_REQUIRED: "true"

# Quality gate thresholds
TEST_COVERAGE_THRESHOLD: "85"
PERFORMANCE_CPU_THRESHOLD: "80"
SECURITY_VULNERABILITY_THRESHOLD: "0"
COMPLIANCE_SCORE_MINIMUM: "80"
```

### Infrastructure Customization

#### AWS (terraform.tfvars)
```hcl
project_name = "my-spec-kit-app"
environment = "staging"
aws_region = "us-west-2"
node_count = 3
node_instance_type = "t3.medium"
enable_constitutional_monitoring = true
domain_name = "app.example.com"
```

#### Azure (parameters.json)
```json
{
  "projectName": {"value": "my-spec-kit-app"},
  "environment": {"value": "staging"},
  "location": {"value": "eastus"},
  "nodeCount": {"value": 3},
  "nodeVMSize": {"value": "Standard_D2s_v3"},
  "enableConstitutionalMonitoring": {"value": true}
}
```

#### Google Cloud (terraform.tfvars)
```hcl
project_id = "my-gcp-project-id"
project_name = "my-spec-kit-app"
environment = "staging"
region = "us-central1"
node_count = 3
node_machine_type = "e2-medium"
enable_constitutional_monitoring = true
```

## CI/CD Integration

### GitHub Actions Workflows

All cloud platforms include GitHub Actions workflows with:

1. **Constitutional Compliance Check**
   - Pre-deployment validation
   - Quality gate enforcement
   - Compliance score calculation

2. **Infrastructure Deployment**
   - Terraform/ARM template execution
   - Resource provisioning
   - Configuration management

3. **Application Deployment**
   - Container image building
   - Kubernetes deployment
   - Health checking

4. **Post-Deployment Validation**
   - Security scanning
   - Performance verification
   - Constitutional compliance monitoring

### Workflow Example

```yaml
# Deploy to staging on push to main
name: Deploy to Cloud
on:
  push:
    branches: [main]

jobs:
  constitutional-compliance:
    # Check compliance before deployment
    
  build-and-deploy:
    needs: constitutional-compliance
    if: needs.constitutional-compliance.outputs.compliance-passed == 'true'
    # Deploy to cloud platform
    
  post-deployment-validation:
    needs: build-and-deploy
    # Verify deployment and compliance
```

## Monitoring and Observability

### Constitutional Compliance Dashboard

Access your constitutional compliance dashboard at:
- AWS: `https://your-alb-dns/constitutional-dashboard`
- Azure: `https://your-agw-ip/constitutional-dashboard`
- GCP: `https://your-ingress-ip/constitutional-dashboard`

Dashboard includes:
- 📊 **Compliance Score**: Real-time constitutional compliance percentage
- 🧪 **TDD Metrics**: Test coverage and execution time
- 🏗️ **Architecture Metrics**: Library count and modularity score
- 🖥️ **CLI Metrics**: Command-line interface availability
- 🔒 **Security Metrics**: Vulnerability count and severity
- ⚡ **Performance Metrics**: CPU, memory, and response time

### Prometheus Metrics

All deployments expose Prometheus metrics:

```
# Constitutional compliance metrics
spec_kit_constitutional_compliance_percentage
spec_kit_tdd_compliance
spec_kit_library_first_compliance
spec_kit_cli_interface_count
spec_kit_dependency_vulnerabilities_total
spec_kit_test_coverage_percentage

# Application performance metrics
spec_kit_http_requests_total
spec_kit_http_request_duration_seconds
spec_kit_cpu_usage_percentage
spec_kit_memory_usage_percentage
```

### Alerting Rules

Pre-configured alerts for constitutional violations:

```yaml
# Constitutional compliance below 80%
- alert: ConstitutionalComplianceBelow80Percent
  expr: spec_kit_constitutional_compliance_percentage < 80
  for: 5m
  
# TDD compliance failure
- alert: TDDComplianceFailure
  expr: spec_kit_tdd_compliance == 0
  for: 2m
  
# High severity vulnerabilities
- alert: DependencyVulnerabilitiesHigh
  expr: spec_kit_dependency_vulnerabilities_high > 0
  for: 1m
```

## Security

### Best Practices

All deployments implement security best practices:

1. **Network Security**
   - Private subnets for application components
   - Network policies and security groups
   - TLS encryption for all communications

2. **Identity and Access Management**
   - Least privilege access principles
   - Service account authentication
   - Workload identity integration

3. **Secrets Management**
   - Cloud-native secret stores (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager)
   - Encryption at rest and in transit
   - Secret rotation capabilities

4. **Container Security**
   - Vulnerability scanning for container images
   - Non-root container execution
   - Read-only root filesystems

### Compliance

Infrastructure templates include compliance controls for:
- **SOC 2**: Logging, monitoring, and access controls
- **GDPR**: Data encryption and privacy controls
- **HIPAA**: Enhanced security and audit logging
- **PCI DSS**: Network segmentation and encryption

## Cost Optimization

### Resource Sizing

Default configurations optimized for cost and performance:

| Environment | CPU | Memory | Storage | Estimated Monthly Cost |
|-------------|-----|--------|---------|----------------------|
| Development | 2 vCPU | 4 GB | 50 GB | $150-200 |
| Staging | 4 vCPU | 8 GB | 100 GB | $300-400 |
| Production | 8 vCPU | 16 GB | 200 GB | $600-800 |

### Auto-scaling

All deployments include auto-scaling:
- **Horizontal Pod Autoscaling**: Scale pods based on CPU/memory
- **Cluster Autoscaling**: Scale nodes based on demand
- **Vertical Pod Autoscaling**: Optimize resource requests

### Cost Monitoring

Built-in cost monitoring and alerting:
- Daily cost reports
- Budget alerts
- Resource utilization tracking
- Rightsizing recommendations

## Troubleshooting

### Common Issues

#### Constitutional Compliance Failures
```bash
# Check compliance status
kubectl logs -l app.kubernetes.io/name=spec-kit-constitutional-monitor

# View compliance metrics
curl http://dashboard-url/api/constitutional-compliance
```

#### Deployment Issues
```bash
# Check pod status
kubectl get pods -n spec-kit-app

# View deployment logs
kubectl logs deployment/spec-kit-app -n spec-kit-app

# Check ingress configuration
kubectl describe ingress spec-kit-ingress -n spec-kit-app
```

#### Infrastructure Issues
```bash
# AWS: Check EKS cluster status
aws eks describe-cluster --name cluster-name

# Azure: Check AKS cluster status
az aks show --name cluster-name --resource-group rg-name

# GCP: Check GKE cluster status
gcloud container clusters describe cluster-name --zone zone-name
```

### Support

For issues and questions:
1. Check the troubleshooting section above
2. Review deployment logs in your cloud console
3. Check constitutional compliance dashboard
4. Open an issue in the Spec-Kit repository

## Migration

### From Other Platforms

Migration guides available for:
- Docker Compose → Kubernetes
- Heroku → Cloud Platform
- Traditional VMs → Containers
- Monolith → Microservices (Library-first)

### Between Cloud Platforms

Cross-cloud migration support:
- Infrastructure state migration
- Data migration tools
- Configuration mapping
- DNS and certificate management

## Contributing

To add support for additional cloud platforms:

1. Create platform-specific directory structure
2. Implement constitutional compliance monitoring
3. Add CI/CD workflow integration
4. Include monitoring and alerting
5. Document configuration and troubleshooting
6. Add cost optimization features

See `CONTRIBUTING.md` for detailed guidelines.