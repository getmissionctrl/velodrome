# Setup data dir
data_dir = "/opt/nomad"

# Enable the server
server {
  enabled = true

  # Self-elect, should be 3 or 5 for production
  bootstrap_expect = 3
}

# Require TLS
tls {
  http = true
  rpc  = true

  ca_file   = "/etc/nomad.d/certs/nomad-ca.pem"
  cert_file = "/etc/nomad.d/certs/server.pem"
  key_file  = "/etc/nomad.d/certs/server-key.pem"

  verify_server_hostname = true
  verify_https_client    = true
}