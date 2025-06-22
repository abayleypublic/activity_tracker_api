variable "gcp_project_id" {
  description = "The ID of the GCP project in which the resources will be created."
  type        = string
  default     = "portfolio-463406"
}

variable "mongodb_atlas_project_id" {
  description = "Project ID for the MongoDB instance."
  type        = string
}

variable "mongodb_atlas_public_key" {
  description = "Public key for MongoDB Atlas API."
  type        = string
}

variable "mongodb_atlas_private_key" {
  description = "Private key for MongoDB Atlas API."
  type        = string
}

variable "mongodb_atlas_org_id" {
  description = "MongoDB Atlas organization ID."
  type        = string
}

variable "mongodb_atlas_idp_id" {
  description = "ID of the MongoDB Atlas Identity Provider for Kubernetes OIDC integration."
  type        = string
}

variable "tenancy_ocid" {
  description = "Oracle Cloud tenancy OCID."
  type        = string
}

variable "user_ocid" {
  description = "Oracle Cloud user OCID."
  type        = string
}

variable "fingerprint" {
  description = "Fingerprint of the Oracle Cloud user."
  type        = string
}

variable "private_key_path" {
  description = "Oracle Cloud private key."
  type        = string
}

variable "region" {
  description = "Oracle Cloud region."
  type        = string
  default     = "uk-london-1"
}
