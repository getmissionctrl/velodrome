agent_prefix "" {
  policy = "read"
}

node_prefix "" {
  policy = "read"
}

service_prefix "" {
  policy = "write"
}

acl = "write"

// consul acl policy create \
//   -name "nomad-server" \
//   -description "Nomad Server Policy" \
//   -rules @nomad-server-policy.hcl