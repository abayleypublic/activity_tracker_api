resource "random_password" "pw" {
  length  = 20
  special = false
}

resource "oci_vault_secret" "prod_roam_mongodb_uri" {
  compartment_id = var.compartment_id
  key_id         = var.key_id
  secret_name    = "prod_roam_mongodb_uri"
  vault_id       = var.vault_id

  secret_content {
    content_type = "BASE64"
    content      = base64encode("mongodb+srv://${var.mongodb_atlas_username}:${random_password.pw.result}@portfolio.gupwyyx.mongodb.net/?retryWrites=true&w=majority")
  }
}

resource "mongodbatlas_database_user" "prod_db_user" {
  project_id         = var.mongodb_atlas_project_id
  username           = var.mongodb_atlas_username
  password           = random_password.pw.result
  auth_database_name = "admin"

  roles {
    role_name     = "readWrite"
    database_name = "roam"
  }
}
