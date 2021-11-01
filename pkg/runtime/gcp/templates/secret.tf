variable "secret_{{.Name}}"{
  type = string
  sensitive = true
}

module "secret-{{.Name}}" {
  source = "../modules/secret_manager"
  secret_id = "{{.Name}}"
  secret_data = var.secret_{{.Name}}
  environment = var.environment
  project = var.project
}
