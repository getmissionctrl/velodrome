variable "hcloud_token" {
  sensitive = true
  type = string
}

variable "server_count" {
  type = number
  default = {{.ClusterConfig.Servers}}
}

variable "client_count" {
  type = number
  default = {{.ClusterConfig.Clients}}
}

variable "separate_consul_servers"{
  type = bool
  default = {{.ClusterConfig.SeparateConsulServers}}
}

variable "multi_instance_observability" {
  type = bool
  default = {{.ObservabilityConfig.MultiInstance}}
}

variable "ssh_keys" {
  type = list
  default = [{{ range $key, $value := .CloudProviderConfig.ProviderSettings.ssh_keys}}
   "{{ $value }}",{{ end }}
  ]
}

variable "base_server_name" {
  type = string
  default = "{{.CloudProviderConfig.ProviderSettings.resource_names.base_server_name}}"
}

variable "firewall_name" {
  type = string
  default = "{{.CloudProviderConfig.ProviderSettings.resource_names.firewall_name}}"
}

variable "network_name" {
  type = string
  default = "{{.CloudProviderConfig.ProviderSettings.resource_names.network_name}}"
}

variable "allow_ips" {
  type = list
  default = [{{ range $key, $value := .CloudProviderConfig.ProviderSettings.allowed_ips}}
   "{{ $value }}",{{ end }}
  ]
}

variable "server_type"{
  type = string
  default = "{{.CloudProviderConfig.ProviderSettings.server_type}}"
}
variable "location"{
  type = string
  default = "{{.CloudProviderConfig.ProviderSettings.location}}" #nbg1, fsn1, hel1 or ash
}