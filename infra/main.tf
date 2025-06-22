resource "random_id" "default" {
  byte_length = 8
}

resource "google_storage_bucket" "state" {
  name     = "${random_id.default.hex}-roam-api-terraform-state"
  project  = var.gcp_project_id
  location = "US-EAST1"

  force_destroy               = false
  public_access_prevention    = "enforced"
  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}

resource "local_file" "default" {
  file_permission = "0644"
  filename        = "${path.module}/backend.tf"

  content = templatefile("${path.module}/backend.tftpl", {
    bucket_name = google_storage_bucket.state.name
  })
}

terraform {
  required_providers {
    mongodbatlas = {
      source  = "mongodb/mongodbatlas"
      version = "1.36.0"
    }

    oci = {
      source  = "oracle/oci"
      version = ">= 7.0.0"
    }
  }
}

provider "mongodbatlas" {
  public_key  = var.mongodb_atlas_public_key
  private_key = var.mongodb_atlas_private_key
}

provider "google" {
  project               = var.gcp_project_id
  billing_project       = var.gcp_project_id
  user_project_override = true
  region                = "europe-west1"
}

provider "oci" {
  region           = var.region
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
}
