terraform {
  required_providers {
    hcloud = {
      source = "hetznercloud/hcloud"
      version = "1.34.3"
    }
  }
}

# Configure the Hetzner Cloud Provider
provider "hcloud" {
  token = var.hcloud_token
}

locals {
  # Common tags to be assigned to all resources
  observability_servers = var.multi_instance_observability ? 4 : 1
  consul_servers = var.separate_consul_servers ? var.server_count : 0 * var.server_count
  nomad_servers = var.server_count
  vault_servers = 0
  nomad_clients = var.client_count

}

resource "hcloud_network" "private_network" {
  name     = "${var.network_name}"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "consul_subnet" {
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.0.0/24"
}

resource "hcloud_network_subnet" "nomad_srv_subnet" {
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_network_subnet" "vault_subnet" {
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.2.0/24"
}

resource "hcloud_network_subnet" "nomad_client_subnet" {
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.3.0/24"
}

resource "hcloud_network_subnet" "observability_subnet" {
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.4.0/24"
}

resource "hcloud_placement_group" "server_placement_group" {
  name = "server_placement_spread_group"
  type = "spread"
}

resource "hcloud_placement_group" "client_placement_group" {
  name = "client_placement_spread_group"
  type = "spread"
}

resource "hcloud_firewall" "network_firewall" {
  name = "${var.firewall_name}"
  rule {
    direction = "in"
    protocol  = "tcp"
    port  = "1-10000"
    source_ips = var.allow_ips
  }

  rule {
    direction = "in"
    protocol      = "icmp"
    source_ips = [
      "0.0.0.0/0",
      "::/0"
    ]
  }
}

resource "hcloud_server" "consul_server_node" {
  count       = local.consul_servers
  name        = "${var.base_server_name}-consul-server-${count.index+1}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = var.location
  placement_group_id = hcloud_placement_group.server_placement_group.id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.consul_subnet
  ]

  ssh_keys = var.ssh_keys
}

resource "hcloud_server" "nomad_server_node" {
  count       = local.nomad_servers
  name        = "${var.base_server_name}-nomad-server-${count.index+1}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = var.location
  placement_group_id = hcloud_placement_group.server_placement_group.id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.nomad_srv_subnet
  ]

  ssh_keys = var.ssh_keys
}

resource "hcloud_server" "vault_server_node" {
  count       = local.vault_servers
  name        = "${var.base_server_name}-vault-${count.index+1}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = var.location
  placement_group_id = hcloud_placement_group.server_placement_group.id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.vault_subnet
  ]

  ssh_keys = var.ssh_keys
}

resource "hcloud_server" "nomad_client_node" {
  count       = local.nomad_clients
  name        = "${var.base_server_name}-client-${count.index+1}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = var.location
  placement_group_id = hcloud_placement_group.client_placement_group.id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.nomad_client_subnet
  ]

  ssh_keys = var.ssh_keys
}

resource "hcloud_server" "observability_node" {
  count       = local.observability_servers
  name        = "${var.base_server_name}-observability-${count.index+1}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = var.location
  placement_group_id = hcloud_placement_group.server_placement_group.id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.observability_subnet
  ]

  ssh_keys = var.ssh_keys
}

resource "hcloud_server_network" "consulsrvnetwork" {
  count       = local.consul_servers
  server_id  = hcloud_server.nomad_server_node[count.index].id
  network_id = hcloud_network.private_network.id
  ip         = "10.0.0.${count.index+2}"
}
resource "hcloud_server_network" "nomadsrvnetwork" {
  count       = local.nomad_servers
  server_id  = hcloud_server.nomad_server_node[count.index].id
  network_id = hcloud_network.private_network.id
  ip         = "10.0.1.${count.index+2}"
}
resource "hcloud_server_network" "vaultsrvnetwork" {
  count       = local.vault_servers
  server_id  = hcloud_server.vault_server_node[count.index].id
  network_id = hcloud_network.private_network.id
  ip         = "10.0.2.${count.index+2}"
}
resource "hcloud_server_network" "clientnetwork" {
  count       = local.nomad_clients
  server_id  = hcloud_server.nomad_client_node[count.index].id
  network_id = hcloud_network.private_network.id
  ip         = "10.0.3.${count.index+2}"
}
resource "hcloud_server_network" "observability" {
  count       = local.observability_servers
  server_id  = hcloud_server.observability_node[count.index].id
  network_id = hcloud_network.private_network.id
  ip         = "10.0.4.${count.index+2}"
}


output "consul_servers" {
  value = [
    for index, node in hcloud_server.consul_server_node : "${node.ipv4_address} host_name=${node.name} private_ip=10.0.0.${index+2}"
  ]
}
output "nomad_servers" {
  value = [
    for index, node in hcloud_server.nomad_server_node : "${node.ipv4_address} host_name=${node.name} private_ip=10.0.1.${index+2}"
  ]
}
output "vault_servers" {
  value = [
    for index, node in hcloud_server.vault_server_node : "${node.ipv4_address} host_name=${node.name} private_ip=10.0.2.${index+2}"
  ]
}

output "client_servers" {
  value = [
    for index, node in hcloud_server.nomad_client_node : "${node.ipv4_address} host_name=${node.name} private_ip=10.0.3.${index+2}"
  ]
}

output "o11y_servers" {
  value = [
    for index, node in hcloud_server.observability_node : "${node.ipv4_address} host_name=${node.name} private_ip=10.0.4.${index+2}"
  ]
}