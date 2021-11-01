output "network_self_link" {
  value       = module.vpc-module.network_self_link
  description = "The URI of the VPC being created"
}


output "serverless_connector" {
  value       = tolist(module.serverless-connector.connector_ids)[0]
  description = "The URI of the VPC being created"
}