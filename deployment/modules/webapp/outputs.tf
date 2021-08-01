output "cloud_run_service_name" {
  value = google_cloud_run_service.default.name
}


output "cloud_run_endpoint" {
  value = google_cloud_run_service.default.status
}