service {
  name = "grafana"
  address = "HOST"
  port = 3000
  tagged_addresses {
    lan = {
      address = "HOST"
      port = 3000
    }
  }
}