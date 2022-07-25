variable "hcloud_token" {
  sensitive = true
  type = string
}

variable "server_count" {
  type = number
  default = {{.ClusterConfig.ServerCount}}
}

variable "client_count" {
  type = number
  default = {{.ClusterConfig.ClientCount}}
}

variable "separate_consul_servers"{
  type = bool
  default = {{.ClusterConfig.SeparateConsulServers}}
}

variable "multi_instance_observability" {
  type = bool
  default = {{.ClusterConfig.MultiInstanceO11y}}
}

variable "ssh_keys" {
  type = list
  default = ["wille.faler@gmail.com"]
}

variable "base_server_name" {
  type = string
  default = "{{.ProviderConfig.ResourceNames.BaseServerName}}"
}

variable "firewall_name" {
  type = string
  default = "{{.ProviderConfig.ResourceNames.FirewallName}}"
}

variable "network_name" {
  type = string
  default = "{{.ProviderConfig.ResourceNames.NetworkName}}"
}

variable "allow_ips" {
  type = list
  default = [
      "85.4.84.201/32"
    ]
}

variable "server_type"{
  type = string
  default = "{{.ProviderConfig.ServerType}}"
}
variable "location"{
  type = string
  default = "{{.ProviderConfig.Location}}" #nbg1, fsn1, hel1 or ash
}