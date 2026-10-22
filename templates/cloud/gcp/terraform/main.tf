# Spec-Kit Google Cloud Platform Deployment
# Terraform configuration for GKE with constitutional monitoring

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.84"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 4.84"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
  
  backend "gcs" {
    # Configure your GCS backend
    bucket = "spec-kit-terraform-state"
    prefix = "infrastructure"
  }
}

# Variables
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "project_name" {
  description = "Name of the Spec-Kit project"
  type        = string
  default     = "spec-kit-app"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "node_count" {
  description = "Number of GKE nodes per zone"
  type        = number
  default     = 1
}

variable "node_machine_type" {
  description = "Machine type for GKE nodes"
  type        = string
  default     = "e2-medium"
}

variable "enable_constitutional_monitoring" {
  description = "Enable constitutional compliance monitoring"
  type        = bool
  default     = true
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = ""
}

# Configure providers
provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# Local values
locals {
  resource_prefix = "${var.project_name}-${var.environment}"
  
  common_labels = {
    project     = var.project_name
    environment = var.environment
    framework   = "spec-kit"
    managed_by  = "terraform"
  }
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "container.googleapis.com",
    "compute.googleapis.com",
    "cloudsql.googleapis.com",
    "redis.googleapis.com",
    "secretmanager.googleapis.com",
    "monitoring.googleapis.com",
    "logging.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "artifactregistry.googleapis.com"
  ])
  
  project = var.project_id
  service = each.value
  
  disable_on_destroy = false
}

# VPC Network
resource "google_compute_network" "main" {
  name                    = "${local.resource_prefix}-vpc"
  auto_create_subnetworks = false
  
  depends_on = [google_project_service.required_apis]
}

# Subnet
resource "google_compute_subnetwork" "main" {
  name          = "${local.resource_prefix}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  region        = var.region
  network       = google_compute_network.main.id
  
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "192.168.0.0/18"
  }
  
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "192.168.64.0/18"
  }
}

# Firewall rules
resource "google_compute_firewall" "allow_internal" {
  name    = "${local.resource_prefix}-allow-internal"
  network = google_compute_network.main.name
  
  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }
  
  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }
  
  allow {
    protocol = "icmp"
  }
  
  source_ranges = ["10.0.0.0/16", "192.168.0.0/18", "192.168.64.0/18"]
}

# Artifact Registry for container images
resource "google_artifact_registry_repository" "main" {
  location      = var.region
  repository_id = "${local.resource_prefix}-repo"
  description   = "Container registry for ${var.project_name} ${var.environment}"
  format        = "DOCKER"
  
  labels = local.common_labels
  
  depends_on = [google_project_service.required_apis]
}

# GKE Cluster
resource "google_container_cluster" "main" {
  name     = "${local.resource_prefix}-gke"
  location = var.region
  
  # Remove default node pool
  remove_default_node_pool = true
  initial_node_count       = 1
  
  network    = google_compute_network.main.name
  subnetwork = google_compute_subnetwork.main.name
  
  # Enable Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
  
  # Enable network policy
  network_policy {
    enabled = true
  }
  
  # IP allocation policy
  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }
  
  # Enable monitoring and logging
  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }
  
  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }
  
  # Enable addons
  addons_config {
    http_load_balancing {
      disabled = false
    }
    
    horizontal_pod_autoscaling {
      disabled = false
    }
    
    network_policy_config {
      disabled = false
    }
    
    gcp_filestore_csi_driver_config {
      enabled = true
    }
  }
  
  # Security settings
  master_auth {
    client_certificate_config {
      issue_client_certificate = false
    }
  }
  
  # Enable private nodes
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }
  
  resource_labels = local.common_labels
  
  depends_on = [google_project_service.required_apis]
}

# GKE Node Pool
resource "google_container_node_pool" "main" {
  name       = "${local.resource_prefix}-node-pool"
  location   = var.region
  cluster    = google_container_cluster.main.name
  node_count = var.node_count
  
  autoscaling {
    min_node_count = 1
    max_node_count = var.node_count * 3
  }
  
  management {
    auto_repair  = true
    auto_upgrade = true
  }
  
  node_config {
    preemptible  = var.environment != "prod"
    machine_type = var.node_machine_type
    disk_size_gb = 50
    disk_type    = "pd-ssd"
    
    # Enable Workload Identity
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    
    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
    
    labels = local.common_labels
    
    tags = ["${local.resource_prefix}-node"]
  }
}

# Cloud SQL PostgreSQL instance
resource "google_sql_database_instance" "main" {
  count               = var.environment == "prod" ? 1 : 0
  name                = "${local.resource_prefix}-postgres"
  database_version    = "POSTGRES_15"
  region              = var.region
  deletion_protection = var.environment == "prod"
  
  settings {
    tier              = "db-f1-micro"
    availability_type = var.environment == "prod" ? "REGIONAL" : "ZONAL"
    disk_size         = 20
    disk_type         = "PD_SSD"
    disk_autoresize   = true
    
    backup_configuration {
      enabled                        = true
      start_time                     = "03:00"
      point_in_time_recovery_enabled = var.environment == "prod"
      backup_retention_settings {
        retained_backups = 7
      }
    }
    
    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.main.id
      enable_private_path_for_google_cloud_services = true
    }
    
    user_labels = local.common_labels
  }
  
  depends_on = [
    google_project_service.required_apis,
    google_service_networking_connection.private_vpc_connection
  ]
}

# Private VPC connection for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  count         = var.environment == "prod" ? 1 : 0
  name          = "${local.resource_prefix}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.main.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  count                   = var.environment == "prod" ? 1 : 0
  network                 = google_compute_network.main.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address[0].name]
}

# Cloud SQL Database
resource "google_sql_database" "main" {
  count    = var.environment == "prod" ? 1 : 0
  name     = replace(var.project_name, "-", "_")
  instance = google_sql_database_instance.main[0].name
}

# Cloud SQL User
resource "google_sql_user" "main" {
  count    = var.environment == "prod" ? 1 : 0
  name     = "app_user"
  instance = google_sql_database_instance.main[0].name
  password = random_password.db_password[0].result
}

resource "random_password" "db_password" {
  count   = var.environment == "prod" ? 1 : 0
  length  = 16
  special = true
}

# Redis Memorystore instance
resource "google_redis_instance" "main" {
  name               = "${local.resource_prefix}-redis"
  memory_size_gb     = 1
  region             = var.region
  location_id        = var.zone
  redis_version      = "REDIS_6_X"
  tier               = "STANDARD_HA"
  auth_enabled       = true
  authorized_network = google_compute_network.main.id
  
  labels = local.common_labels
  
  depends_on = [google_project_service.required_apis]
}

# Cloud Storage bucket
resource "google_storage_bucket" "app_assets" {
  name          = "${local.resource_prefix}-assets-${random_id.bucket_suffix.hex}"
  location      = var.region
  force_destroy = var.environment != "prod"
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }
  
  encryption {
    default_kms_key_name = google_kms_crypto_key.bucket_key.id
  }
  
  labels = local.common_labels
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Cloud KMS for encryption
resource "google_kms_key_ring" "main" {
  name     = "${local.resource_prefix}-keyring"
  location = var.region
}

resource "google_kms_crypto_key" "bucket_key" {
  name     = "${local.resource_prefix}-bucket-key"
  key_ring = google_kms_key_ring.main.id
  purpose  = "ENCRYPT_DECRYPT"
  
  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }
  
  labels = local.common_labels
}

# Secret Manager secrets
resource "google_secret_manager_secret" "app_secrets" {
  secret_id = "${local.resource_prefix}-secrets"
  
  labels = local.common_labels
  
  replication {
    automatic = true
  }
  
  depends_on = [google_project_service.required_apis]
}

resource "google_secret_manager_secret_version" "app_secrets" {
  secret = google_secret_manager_secret.app_secrets.id
  
  secret_data = jsonencode({
    database_password = var.environment == "prod" ? random_password.db_password[0].result : "dev_password"
    redis_auth_string = google_redis_instance.main.auth_string
    jwt_secret        = random_password.jwt_secret.result
  })
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = false
}

# Service Account for Workload Identity
resource "google_service_account" "app" {
  account_id   = "${local.resource_prefix}-app"
  display_name = "Service account for ${var.project_name} ${var.environment}"
}

# IAM bindings for the service account
resource "google_project_iam_member" "app_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.app.email}"
}

resource "google_project_iam_member" "app_storage_admin" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.app.email}"
}

resource "google_project_iam_member" "app_monitoring_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.app.email}"
}

# Kubernetes Service Account with Workload Identity binding
resource "google_service_account_iam_binding" "workload_identity" {
  service_account_id = google_service_account.app.name
  role               = "roles/iam.workloadIdentityUser"
  
  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[spec-kit-app/spec-kit-app]"
  ]
}

# Constitutional monitoring service account
resource "google_service_account" "constitutional_monitor" {
  count        = var.enable_constitutional_monitoring ? 1 : 0
  account_id   = "${local.resource_prefix}-constitutional-monitor"
  display_name = "Constitutional monitoring service account"
}

resource "google_project_iam_member" "constitutional_monitor_monitoring" {
  count   = var.enable_constitutional_monitoring ? 1 : 0
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.constitutional_monitor[0].email}"
}

resource "google_project_iam_member" "constitutional_monitor_logging" {
  count   = var.enable_constitutional_monitoring ? 1 : 0
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.constitutional_monitor[0].email}"
}

# Outputs
output "cluster_name" {
  description = "GKE cluster name"
  value       = google_container_cluster.main.name
}

output "cluster_endpoint" {
  description = "GKE cluster endpoint"
  value       = google_container_cluster.main.endpoint
}

output "cluster_location" {
  description = "GKE cluster location"
  value       = google_container_cluster.main.location
}

output "artifact_registry_repository" {
  description = "Artifact Registry repository URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.main.repository_id}"
}

output "database_connection_string" {
  description = "Database connection string"
  value       = var.environment == "prod" ? "postgresql://app_user:${random_password.db_password[0].result}@${google_sql_database_instance.main[0].private_ip_address}:5432/${google_sql_database.main[0].name}" : "postgresql://localhost:5432/spec_kit_dev"
  sensitive   = true
}

output "redis_host" {
  description = "Redis instance host"
  value       = google_redis_instance.main.host
}

output "redis_port" {
  description = "Redis instance port"
  value       = google_redis_instance.main.port
}

output "storage_bucket_name" {
  description = "Cloud Storage bucket name"
  value       = google_storage_bucket.app_assets.name
}

output "app_service_account_email" {
  description = "Application service account email"
  value       = google_service_account.app.email
}

output "constitutional_monitor_service_account_email" {
  description = "Constitutional monitoring service account email"
  value       = var.enable_constitutional_monitoring ? google_service_account.constitutional_monitor[0].email : ""
}