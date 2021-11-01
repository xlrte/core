module "private_network-network" {
  source = "../modules/private_network"
  project = var.project
  region = var.region
  environment = var.environment
  min_instances = {{.MinInstances}}
  max_instances = {{.MaxInstances}}
  instance_type = "{{.InstanceType}}"
}
