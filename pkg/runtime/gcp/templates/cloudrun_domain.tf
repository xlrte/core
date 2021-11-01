module "cloudrun_network-{{.ServiceName}}" {
  source = "../modules/cloudrun_network"
  project = var.project
  region = var.region
  environment = var.environment
  cloud_run_service_name = "{{.ServiceName}}"
  domain = "{{.NetworkConfig.Domain.Name}}"
  dns_name = "{{.NetworkConfig.DNSName}}"
  managed_zone = "{{.NetworkConfig.Domain.DNSZone}}"
}
