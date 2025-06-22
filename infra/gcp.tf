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
  compartment_id = "ocid1.compartment.oc1..aaaaaaaaeef3gmatoiwjkxw65hot2n3lb7qkxjuw5k3vebozv7wl2od7cwvq"
  key_id         = "ocid1.key.oc1.uk-london-1.eruckcfxaaarq.abwgiljtucsyrvx76zf5uwpb5flj5vyqgjqldzwdn25fwo2ssycgisrvvywq"
  secret_name    = "prod_roam_maps_api_key"
  vault_id       = "ocid1.vault.oc1.uk-london-1.eruckcfxaaarq.abwgiljsf5ckladyqmxmnjo2zgc6sdymeqsic3ydbgrhuj7l2ulhl7k7mgda"

  secret_content {
    content_type = "BASE64"
    content      = base64encode(google_apikeys_key.maps.key_string)
  }
}
