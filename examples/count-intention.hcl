Kind = "service-intentions"
Name = "count-api"
Sources = [
  {
    Name   = "count-dashboard"
    Action = "allow"
  }
]

// consul config write intention-config.hcl