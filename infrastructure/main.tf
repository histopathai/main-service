terraform {
  required_version = ">=1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
  backend "gcs" {}
}

data "terraform_remote_state" "platform" {
  backend = "gcs"

  config = {
    bucket = var.tf_state_bucket
    prefix = "platform/prod"
  }
}
data "terraform_remote_state" "image_processing" {
  backend = "gcs"

  config = {
    bucket = var.tf_state_bucket
    prefix = "services/image-processing-service"
  }
}



locals {
  # GCP project and region info
  project_id     = data.terraform_remote_state.platform.outputs.project_id
  project_number = data.terraform_remote_state.platform.outputs.project_number
  region         = data.terraform_remote_state.platform.outputs.region

  # Service info
  service_account        = data.terraform_remote_state.platform.outputs.main_service_account_email
  artifact_repository_id = data.terraform_remote_state.platform.outputs.artifact_repository_id
  service_name           = var.environment == "prod" ? "main-service" : "main-service-${var.environment}"

  # Environment-based prefix for PubSub resources
  # In dev environment, all topics/subscriptions get "dev-" prefix for isolation
  pubsub_prefix = var.environment == "dev" ? "dev-" : ""

  # Construct the full image path
  image_name = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_registry_repo}/main-service:${var.image_tag}"

  # Storage bucket info
  original_bucket_name  = data.terraform_remote_state.platform.outputs.original_bucket_name
  processed_bucket_name = data.terraform_remote_state.platform.outputs.processed_bucket_name

  # Cloud Run job names with environment suffix
  # In dev environment, jobs have "-dev" suffix (e.g., image-processing-job-small-dev)
  job_suffix = var.environment == "dev" ? "-dev" : ""
  job_small  = "${data.terraform_remote_state.image_processing.outputs.job_ids["small"]}${local.job_suffix}"
  job_medium = "${data.terraform_remote_state.image_processing.outputs.job_ids["medium"]}${local.job_suffix}"
  job_large  = "${data.terraform_remote_state.image_processing.outputs.job_ids["large"]}${local.job_suffix}"
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
          cpu    = var.cpu_limit
          memory = var.memory_limit
        }
        cpu_idle = true
      }

      ports {
        container_port = 8080
      }

      env {
        name  = "PROJECT_ID"
        value = local.project_id
      }

      env {
        name  = "REGION"
        value = local.region
      }

      env {
        name  = "PROJECT_NUMBER"
        value = local.project_number
      }

      env {
        name  = "ENV"
        value = var.environment == "prod" ? "PROD" : "DEV"
      }

      env {
        name  = "GIN_MODE"
        value = var.environment == "prod" ? "release" : "debug"
      }

      env {
        name  = "LOG_LEVEL"
        value = var.log_levels
      }

      env {
        name  = "LOG_FORMAT"
        value = var.log_format
      }

      env {
        name  = "READ_TIMEOUT"
        value = var.read_timeout
      }

      env {
        name  = "WRITE_TIMEOUT"
        value = var.write_timeout
      }

      env {
        name  = "IDLE_TIMEOUT"
        value = var.idle_timeout
      }

      env {
        name  = "FIRESTORE_DATABASE"
        value = var.environment == "prod" ? "(default)" : "dev-database"
      }

      # --- Platform specific env variables ---
      env {
        name  = "ORIGINAL_BUCKET_NAME"
        value = local.original_bucket_name
      }

      env {
        name  = "PROCESSED_BUCKET_NAME"
        value = local.processed_bucket_name
      }

      # --- Pub/Sub Configuration ---
      # Upload Status
      env {
        name  = "UPLOAD_STATUS_SUBSCRIPTION"
        value = google_pubsub_subscription.upload_status_sub.name
      }

      env {
        name  = "UPLOAD_STATUS_TOPIC"
        value = google_pubsub_topic.upload_status.name
      }

      # Image Processing Request
      env {
        name  = "IMAGE_PROCESSING_REQUEST_TOPIC"
        value = google_pubsub_topic.image_processing_request.name
      }

      env {
        name  = "IMAGE_PROCESSING_REQUEST_SUB"
        value = google_pubsub_subscription.image_processing_request_sub.name
      }

      env {
        name  = "IMAGE_PROCESSING_REQUEST_DLQ"
        value = google_pubsub_topic.image_processing_request_dlq.name
      }

      # Image Processing Result
      env {
        name  = "IMAGE_PROCESSING_RESULT_TOPIC"
        value = google_pubsub_topic.image_processing_result.name
      }

      env {
        name  = "IMAGE_PROCESSING_RESULT_SUB"
        value = google_pubsub_subscription.image_processing_result_sub.name
      }

      env {
        name  = "IMAGE_PROCESSING_RESULT_DLQ"
        value = google_pubsub_topic.image_processing_result_dlq.name
      }

      # Image Deletion
      env {
        name  = "IMAGE_DELETION_TOPIC"
        value = google_pubsub_topic.image_deletion.name
      }

      env {
        name  = "IMAGE_DELETION_SUB"
        value = google_pubsub_subscription.image_deletion_sub.name
      }

      env {
        name  = "IMAGE_DELETION_DLQ"
        value = google_pubsub_topic.image_deletion_dlq.name
      }

      # Image Process DLQ
      env {
        name  = "IMAGE_PROCESS_DLQ_TOPIC"
        value = google_pubsub_topic.image_process_dlq.name
      }

      env {
        name  = "IMAGE_PROCESS_DLQ_SUB"
        value = google_pubsub_subscription.image_process_dlq_sub.name
      }

      # Worker Config Env Vars
      env {
        name  = "WORKER_TYPE"
        value = "cloudrun"
      }
      env {
        name  = "CLOUD_RUN_JOB_SMALL"
        value = local.job_small
      }
      env {
        name  = "CLOUD_RUN_JOB_MEDIUM"
        value = local.job_medium
      }
      env {
        name  = "CLOUD_RUN_JOB_LARGE"
        value = local.job_large
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
    version     = "2" # Force update to apply FIRESTORE_DATABASE env var
  }
}

resource "google_cloud_run_v2_service_iam_member" "auth_service_access" {
  project  = google_cloud_run_v2_service.main_service.project
  location = google_cloud_run_v2_service.main_service.location
  name     = google_cloud_run_v2_service.main_service.name
  role     = "roles/run.invoker"

  member = "serviceAccount:${data.terraform_remote_state.platform.outputs.auth_service_account_email}"
}

resource "google_pubsub_topic_iam_member" "main_service_publishers" {
  for_each = {
    (google_pubsub_topic.image_processing_request.name) = google_pubsub_topic.image_processing_request.name
    (google_pubsub_topic.image_processing_result.name)  = google_pubsub_topic.image_processing_result.name
    (google_pubsub_topic.image_deletion.name)           = google_pubsub_topic.image_deletion.name
    (google_pubsub_topic.image_process_dlq.name)        = google_pubsub_topic.image_process_dlq.name
  }

  topic  = each.value
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:${local.service_account}"
}

resource "google_project_iam_member" "main_service_job_runner" {
  project = local.project_id
  role    = "roles/run.developer"
  member  = "serviceAccount:${local.service_account}"
}
