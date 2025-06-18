resource "mongodbatlas_database_user" "prod_db_user" {
  project_id         = var.mongodb_atlas_project_id
  username           = "${var.mongodb_atlas_idp_id}/system:serviceaccount:roam:roam-sa"
  auth_database_name = "$external"
  oidc_auth_type     = "USER"
  roles {
    role_name     = "readWrite"
    database_name = "roam"
  }
}
