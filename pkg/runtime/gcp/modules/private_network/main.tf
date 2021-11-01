

locals {
  private_network_name = "private-network-${var.environment}"
}
module "vpc-module" {
  source       = "terraform-google-modules/network/google"
  version      = ">= 3.4"
  project_id   = var.project # Replace this with your project ID in quotes
  network_name = local.private_network_name
  mtu          = 1460

  subnets = [
    {
      subnet_name   = "serverless-subnet-${var.environment}"
      subnet_ip     = "10.10.10.0/28"
      subnet_region = var.region
    }
  ]
} #module.vpc-module.network_self_link

# https://github.com/terraform-google-modules/terraform-google-network/blob/master/variables.tf
module "serverless-connector" {
  source     = "terraform-google-modules/network/google//modules/vpc-serverless-connector-beta"
  project_id = var.project
  vpc_connectors = [{
      name        = "serverless-${var.environment}"
      region      = var.region
      subnet_name = module.vpc-module.subnets["${var.region}/serverless-subnet-${var.environment}"].name

      machine_type  = var.instance_type # #f1-micro, e2-standard-4
      min_instances = var.min_instances
      max_instances = var.max_instances
    }
  ]
  depends_on = [
    module.vpc-module
  ]
}
