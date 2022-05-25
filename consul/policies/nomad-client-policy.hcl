agent_prefix "" {
  policy = "read"
}

node_prefix "" {
  policy = "read"
}

service_prefix "" {
  policy = "write"
}

# uncomment if using Consul KV with Consul Template
key_prefix "" {
   policy = read
}

// consul acl policy create \
//   -name "nomad-client" \
//   -description "Nomad Client Policy" \
//   -rules @nomad-client-policy.hcl