# ========================================
# PUB/SUB TOPICS & SUBSCRIPTIONS
# ========================================
data "google_storage_project_service_account" "gcs_account" {
}

# ----------------------------------------
# 1. UPLOAD STATUS (GCS Notifications)
# ----------------------------------------
resource "google_pubsub_topic" "upload_status" {
    name = "upload-status"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "86400s" # 1 day

}

resource "google_pubsub_subscription" "upload_status_sub" {
    name = "upload-status-sub"
    topic = google_pubsub_topic.upload_status.name

    ack_deadline_seconds = 600 # 10 minutes

    retry_policy {
        minimum_backoff = "10s"
        maximum_backoff = "600s"
    }

    labels = {
      service = "main-service"
      managed_by = "terraform"
    }
}

# Allow GCS to publish to upload status topic
resource "google_pubsub_topic_iam_member" "gcs_publisher" {
  topic  = google_pubsub_topic.upload_status.id
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"
}

# GCS Notification for original bucket
resource "google_storage_notification" "original_bucket_upload" {
  bucket         = local.original_bucket_name
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.upload_status.id
  event_types    = ["OBJECT_FINALIZE"]

  depends_on = [google_pubsub_topic_iam_member.gcs_publisher]
}

# ----------------------------------------
# 2. IMAGE PROCESSING
# ----------------------------------------
resource "google_pubsub_topic" "image_processing_request" {
    name = "image-processing-request"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "86400s" # 1 day
}

resource "google_pubsub_topic" "image_processing_request_dlq" {
    name = "image-processing-request-dlq"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "604800s" # 7 days
}

resource "google_pubsub_subscription" "image_processing_request_sub" {
    name = "image-processing-request-sub"
    topic = google_pubsub_topic.image_processing_request.name

    ack_deadline_seconds = 600 # 10 minutes

    retry_policy {
        minimum_backoff = "10s"
        maximum_backoff = "600s"
    }
    dead_letter_policy {
        dead_letter_topic     = google_pubsub_topic.image_processing_request_dlq.id
        max_delivery_attempts = 5
    }

    labels = {
        service     = "main-service"
        managed_by  = "terraform"
    }
}

# -----------------------------------------
# 3. IMAGE PROCESSING RESULT
# -----------------------------------------
resource "google_pubsub_topic" "image_processing_result" {
    name = "image-processing-results"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "86400s" # 1 day
}

resource "google_pubsub_topic" "image_processing_result_dlq" {
    name = "image-processing-results-dlq"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "604800s" # 7 days
}

resource "google_pubsub_subscription" "image_processing_result_sub" {
    name = "image-processing-results-sub"
    topic = google_pubsub_topic.image_processing_result.name

    ack_deadline_seconds = 600 # 10 minutes

    retry_policy {
        minimum_backoff = "10s"
        maximum_backoff = "600s"
    }
    dead_letter_policy {
        dead_letter_topic     = google_pubsub_topic.image_processing_result_dlq.id
        max_delivery_attempts = 5
    }

    labels = {
        service     = "main-service"
        managed_by  = "terraform"
    }
}


# -----------------------------------------
# 4. IMAGE DELETION
# -----------------------------------------
resource "google_pubsub_topic" "image_deletion" {
    name = "image-deletion-requests"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "86400s" # 1 day
}

resource "google_pubsub_topic" "image_deletion_dlq" {
    name = "image-deletion-requests-dlq"

    labels = {
        service = "main-service"
        managed_by = "terraform"
    }

    message_retention_duration = "604800s" # 7 days
}

resource "google_pubsub_subscription" "image_deletion_sub" {
    name = "image-deletion-requests-sub"
    topic = google_pubsub_topic.image_deletion.name

    ack_deadline_seconds =  600 # 10 minutes

    retry_policy {
        minimum_backoff = "10s"
        maximum_backoff = "600s"
    }

    dead_letter_policy {
        dead_letter_topic     = google_pubsub_topic.image_deletion_dlq.id
        max_delivery_attempts = 5
    }

    labels = {
        service     = "main-service"
        managed_by  = "terraform"
    }
}

# ----------------------------------------
# 5. IMAGE PROCESS DLQ
# ----------------------------------------
resource "google_pubsub_topic" "image_process_dlq" {
  name = "image-process-dlq"

  labels = {
    service     = "main-service"
    managed_by  = "terraform"
  }

  message_retention_duration = "604800s" # 7 days
}

resource "google_pubsub_subscription" "image_process_dlq_sub" {
  name  = "image-process-dlq-sub"
  topic = google_pubsub_topic.image_process_dlq.name

  ack_deadline_seconds = 600 # 10 minutes

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = {
    service     = "main-service"
    managed_by  = "terraform"
  }
}