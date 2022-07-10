service {
  name = "tempo"
  address = "10.0.0.7"
  port = 3200
  tagged_addresses {
    lan = {
      address = "10.0.0.7"
      port = 3200
    }
  }
}

