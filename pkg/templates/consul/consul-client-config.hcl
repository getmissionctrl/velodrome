datacenter = "dc1"
data_dir = "/opt/consul"
encrypt = "{{ CONSUL_GOSSIP_KEY }}"
verify_incoming = true
verify_outgoing = true
verify_server_hostname = true

server = false

bind_addr = "{{ private_ip }}"

ca_file = "/etc/consul.d/certs/consul-agent-ca.pem"

auto_encrypt {
  tls = true
}

ports {
  grpc = 8502
}

connect {
  enabled = true
}

retry_join = [join_servers]

acl {
  tokens {
    agent  = "{{agent_token}}"
  }
}