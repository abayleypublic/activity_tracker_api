resource "google_apikeys_key" "maps" {
  name         = "roam-maps-frontend"
  display_name = "Maps API Key for Roam Frontend"

  restrictions {
    browser_key_restrictions {
      allowed_referrers = [
        "https://roam.austinbayley.co.uk/*",
      ]
    }

    api_targets {
      service = "geocoding-backend.googleapis.com"
    }
  }
}

# I would like to use the Google Secret Manager to store application secrets but the External
# Secrets Operator does not support WIF outside of GKE, so I will use OCI Vault instead.
# TODO: revisit once ESO supports WIF for GCP outside of GKE.

# resource "google_secret_manager_secret" "maps_key" {
#   secret_id = "prod-roam-maps-api-key"
#   replication {
#     user_managed {
#       replicas {
#         location = "europe-west1"
#       }
#     }
#   }
# }

# resource "google_secret_manager_secret_version" "maps_key" {
#   secret      = google_secret_manager_secret.maps_key.id
#   secret_data = google_apikeys_key.maps.key_string
# }

resource "oci_vault_secret" "prod_maps_api_key" {
  compartment_id = var.compartment_id
  key_id         = var.key_id
  secret_name    = "prod_roam_maps_api_key"
  vault_id       = var.vault_id

  secret_content {
    content_type = "BASE64"
    content      = base64encode(google_apikeys_key.maps.key_string)
  }
}
