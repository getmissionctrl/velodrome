variable "hcloud_token" {
  sensitive = true
  type = string
}

variable "server_count" {
  type = number
  default = 3
}

variable "client_count" {
  type = number
  default = 2
}

variable "separate_consul_servers"{
  type = bool
  default = false
}

variable "multi_instance_observability" {
  type = bool
  default = false
}

variable "ssh_keys" {
  type = list
  default = ["wille.faler@gmail.com"]
}

variable "base_server_name" {
  type = string
  default = "nomad-srv"
}

variable "firewall_name" {
  type = string
  default = "dev_firewall"
}

variable "network_name" {
  type = string
  default = "dev_network"
}

variable "allow_ips" {
  type = list
  default = [
      "85.4.84.201/32"
    ]
}

variable "server_type"{
  type = string
  default = "cx21"
}
variable "location"{
  type = string
  default = "nbg1" #nbg1, fsn1, hel1 or ash
}