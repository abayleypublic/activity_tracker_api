# ==========
# Staging
# ==========


resource "random_password" "stg_pw" {
  length  = 20
  special = false
}

resource "oci_vault_secret" "stg_roam_mongodb_uri" {
  compartment_id = var.compartment_id
  key_id         = var.key_id
  secret_name    = "stg_roam_mongodb_uri"
  vault_id       = var.vault_id

  secret_content {
    content_type = "BASE64"
    content      = base64encode("mongodb+srv://${var.stg_mongodb_atlas_username}:${random_password.stg_pw.result}@portfolio.gupwyyx.mongodb.net/?retryWrites=true&w=majority")
  }
}

resource "mongodbatlas_database_user" "stg_db_user" {
  project_id         = var.mongodb_atlas_project_id
  username           = var.stg_mongodb_atlas_username
  password           = random_password.stg_pw.result
  auth_database_name = "admin"

  roles {
    role_name     = "readWrite"
    database_name = "stg_activity_tracker"
  }
}

# ==========
# Production
# ==========

resource "random_password" "prod_pw" {
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
    content      = base64encode("mongodb+srv://${var.prod_mongodb_atlas_username}:${random_password.prod_pw.result}@portfolio.gupwyyx.mongodb.net/?retryWrites=true&w=majority")
  }
}

resource "mongodbatlas_database_user" "prod_db_user" {
  project_id         = var.mongodb_atlas_project_id
  username           = var.prod_mongodb_atlas_username
  password           = random_password.prod_pw.result
  auth_database_name = "admin"

  roles {
    role_name     = "readWrite"
    database_name = "prod_activity_tracker"
  }
}
