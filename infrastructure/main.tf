terraform {
  required_version = ">=1.5.0"

  required_providers {
    google = {
        source = "hashicorp/google"
        version = "~> 5.0"
    }
  }
  backend "gcs" {
    prefix = "services/main-service"
  }
}

data "terraform_remote_state" "platform" {
    backend = "gcs"

    config = {
        bucket = var.tf_state_bucket
        prefix = "platform/prod"
    }
}

locals {
    # GCP project and region info
    project_id      = data.terraform_remote_state.platform.outputs.project_id
    project_number  = data.terraform_remote_state.platform.outputs.project_number
    region          = data.terraform_remote_state.platform.outputs.region

    # Service info
    service_account         = data.terraform_remote_state.platform.outputs.main_service_account_email
    artifact_repository_id  = data.terraform_remote_state.platform.outputs.artifact_repository_id
    service_name            = var.environment == "prod" ? "main-service" : "main-service-${var.environment}"
    image_name              = "${local.region}-docker.pkg.dev/${local.project_id}/${local.artifact_repository_id}/${local.service_name}:${var.image_tag}"

    #Pub/Sub info
    upload_status_subscription          = data.terraform_remote_state.platform.outputs.upload_status_subscription
    image_processing_topic              = data.terraform_remote_state.platform.outputs.image_processing_topic
    image_processing_dlq_topic          = data.terraform_remote_state.platform.outputs.image_processing_dlq_topic
    processing_completed_subscription   = data.terraform_remote_state.platform.outputs.processing_completed_subscription
    processing_completed_dlq_topic      = data.terraform_remote_state.platform.outputs.processing_completed_dlq_topic

    # Storage bucket info
    original_bucket_name  = data.terraform_remote_state.platform.outputs.original_bucket_name
    processed_bucket_name = data.terraform_remote_state.platform.outputs.processed_bucket_name
}

provider "google" {
  project = local.project_id
  region  = local.region
}

# ----------------------------------
# CLOUD RUN SERVICE
# ----------------------------------

resource "google_cloud_run_v2_service" "main_service" {
    name     = local.service_name
    location = local.region
    ingress  = "INGRESS_TRAFFIC_ALL"
    template {
        service_account = local.service_account
        scaling {
            min_instance_count = var.min_instances
            max_instance_count = var.max_instances
        }

        containers {
            image = local.image_name
            resources {
                limits = {
                    cpu     = var.cpu_limit
                    memory  = var.memory_limit
                }
            }

            ports {
                container_port = 8080
            }

            env {
                name  = "PROJECT_ID"
                value = local.project_id
            }

            env {
                name  = "PROJECT_NUMBER"
                value = local.project_number
            }

            env {
                name  = "REGION"
                value = local.region
            }

            env {
                name  = "ENV"
                value = var.environment == "prod" ? "PROD" : "DEV"
            }

            env {
                name = "GIN_MODE"
                value = var.environment == "prod" ? "release" : "debug"
            }

            env {
                name = "LOG_LEVEL"
                value = var.log_levels
            }

            env {
                name = "LOG_FORMAT"
                value = "json"
            }

            env  {
                name = "READ_TIMEOUT"
                value = "15s"
            }

            env {
                name  = "WRITE_TIMEOUT"
                value = "30s"
            }

            env {
                name  = "IDLE_TIMEOUT"
                value = "120s"
            }

            # --- Platform specific env variables ---
            env {
                name = "ORIGINAL_BUCKET_NAME"
                value = local.original_bucket_name
            }

            env {
                name = "PROCESSED_BUCKET_NAME"
                value = local.processed_bucket_name
            }

            env {
                name = "UPLOAD_STATUS_SUBSCRIPTION"
                value = local.upload_status_subscription
            }

            env {
                name = "IMAGE_PROCESSING_TOPIC"
                value = local.image_processing_topic
            }

            env {
                name = "IMAGE_PROCESSING_DLQ_TOPIC"
                value = local.image_processing_dlq_topic
            }

            env {
                name = "PROCESSING_COMPLETED_SUBSCRIPTION"
                value = local.processing_completed_subscription
            }

            env {
                name = "PROCESSING_COMPLETED_DLQ_TOPIC"
                value = local.processing_completed_dlq_topic
            }
        }
    }

    traffic {
        type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
        percent = 100
    }

    labels = {
        environment = var.environment
        service     = "main-service"
        managed_by  = "terraform"
    }
}

resource "google_cloud_run_v2_service_iam_member" "auth_service_access" {
  
  project  = google_cloud_run_v2_service.main_service.project
  location = google_cloud_run_v2_service.main_service.location
  name     = google_cloud_run_v2_service.main_service.name
  role     = "roles/run.invoker"
  
  member   = "serviceAccount:${data.terraform_remote_state.platform.outputs.auth_service_account_email}"
}