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

  groups = {
    consul = {
      count = var.separate_consul_servers ? var.server_count : 0 * var.server_count, subnet = "0", group = 0
    }
    nomad-server = {
      count = var.server_count, subnet = "1", group = 0
    },
    vault = {
      count = 0, subnet = "2", group = 0
    },
    client = {
      count = var.client_count, subnet = "3", group = 1
    },
    observability = {
      count = var.multi_instance_observability ? 4 : 1, subnet = "4", group = 1
    }
  }

  servers = flatten([
    for name, value in local.groups : [
      for i in range(value.count) : {
        group_name = name, 
        private_ip = "10.0.${value.subnet}.${i + 2}", 
        name = "${var.base_server_name}-${name}-${i +1}", 
        group = value.group
      }
    ]
  ])

  placement_groups = 2
}

resource "hcloud_network" "private_network" {
  name     = "${var.network_name}"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "network_subnet" {
  for_each = local.groups
  network_id   = hcloud_network.private_network.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.${each.value.subnet}.0/24"
}

resource "hcloud_placement_group" "placement_group" {
  count = local.placement_groups
  name = "server_placement_spread_group-${count.index}"
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

resource "hcloud_server" "server_node" {
  for_each      = { for entry in local.servers: "${entry.name}" => entry }
  name        = "${each.value.name}"
  image       = "ubuntu-22.04"
  server_type = var.server_type
  location = var.location
  placement_group_id = hcloud_placement_group.placement_group[each.value.group].id
  firewall_ids = [hcloud_firewall.network_firewall.id]

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  depends_on = [
    hcloud_network_subnet.network_subnet["consul"],
    hcloud_network_subnet.network_subnet["nomad-server"],
    hcloud_network_subnet.network_subnet["vault"],
    hcloud_network_subnet.network_subnet["client"],
    hcloud_network_subnet.network_subnet["observability"],
  ]

  labels = {
    "group" = each.value.group_name
  }

  ssh_keys = var.ssh_keys
}

resource "hcloud_server_network" "network_binding" {
  for_each    = { for entry in local.servers: "${entry.name}" => entry }
  server_id  = hcloud_server.server_node[each.value.name].id
  network_id = hcloud_network.private_network.id
  ip         = each.value.private_ip
}


output "consul_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers: 
      "${node.ipv4_address} host_name=${node.name} private_ip=${server.private_ip}" if server.name == node.name 
    ] if node.labels["group"] == "consul"
  ])
}

output "nomad_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers: 
      "${node.ipv4_address} host_name=${node.name} private_ip=${server.private_ip}" if server.name == node.name 
    ] if node.labels["group"] == "nomad-server"
  ])
}

output "vault_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers: 
      "${node.ipv4_address} host_name=${node.name} private_ip=${server.private_ip}" if server.name == node.name 
    ] if node.labels["group"] == "vault"
  ])
}

output "client_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers: 
      "${node.ipv4_address} host_name=${node.name} private_ip=${server.private_ip}" if server.name == node.name 
    ] if node.labels["group"] == "client"
  ])
}

output "o11y_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers: 
      "${node.ipv4_address} host_name=${node.name} private_ip=${server.private_ip}" if server.name == node.name 
    ] if node.labels["group"] == "observability"
  ])
}
