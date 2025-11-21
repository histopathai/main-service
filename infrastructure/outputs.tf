output "service_url" {
  description = "The URL of the deployed main service"
  value       = google_cloud_run_v2_service.main_service.uri
}