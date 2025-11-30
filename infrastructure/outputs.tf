output "service_url" {
  description = "The URL of the deployed main service"
  value       = google_cloud_run_v2_service.main_service.uri
}
output "processing_completed_topic" {
  description = "The Pub/Sub topic for image processing results"
  value       = google_pubsub_topic.image_processing_result.name
}