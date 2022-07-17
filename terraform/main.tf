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
  observability_instances = var.multi_instance_observability ? 4 : 1
  consul_servers = var.separate_consul_servers ? var.server_count : 0 * var.server_count
  nomad_servers = var.server_count
  nomad_clients = var.client_count
  server_instances = local.observability_instances + local.consul_servers + local.nomad_servers + local.nomad_clients
}

resource "hcloud_network" "private_network" {
  name     = "${var.network_name}"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "private_subnet" {
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.0.0/24"
}

resource "hcloud_placement_group" "placement_group" {
  name = "placement_spread_group"
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

resource "hcloud_server" "cluster_node" {
  count       = local.server_instances
  name        = "${var.base_server_name}${count.index+1}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = vars.location
  placement_group_id = hcloud_placement_group.placement_group.id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.private_subnet
  ]

  ssh_keys = var.ssh_keys
}

resource "hcloud_server_network" "srvnetwork" {
  count       = local.server_instances
  server_id  = hcloud_server.cluster_node[count.index].id
  network_id = hcloud_network.private_network.id
  ip         = "10.0.0.${count.index+2}"
}

output "servers" {
  value = [
    for node in hcloud_server.cluster_node : "${node.ipv4_address} host_name=${node.name}"
  ]
}
# private_ip=10.0.0.2 host_name=nomad-cluster-nbg-1