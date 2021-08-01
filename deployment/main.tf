locals{
  environment = "${terraform.workspace == "default" ? "prod" : terraform.workspace}"
}

module "webapp" {
  source = "./modules/webapp"
  region = var.region
  project = var.project
  image_id = var.image_id
  traffic = var.traffic
  service_name = var.service_name
}

module "network" {
  source = "./modules/network"
  region = var.region
  project = var.project
  cloud_run_service_name = module.webapp.cloud_run_service_name

  domain = var.domain
  dns_name = var.dns_name
  managed_zone = var.managed_zone
  network_name = var.network_name

}


output "cloud-run-url"{
  value = module.webapp.cloud_run_endpoint[0].url
}

output "environment"{
  value = "${terraform.workspace == "default" ? "prod" : terraform.workspace}"
}