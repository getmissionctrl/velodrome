job "demo" {

  datacenters = ["hetzner"]
  type = "service"

  update {
    max_parallel = 1
    min_healthy_time = "10s"
    healthy_deadline = "3m"
    progress_deadline = "10m"
    auto_revert = false
    canary = 0
  }

  migrate {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "10s"
    healthy_deadline = "5m"
  }
  group "app" {

    count = 1

    network {
      port "http" {
        to = 8080
      }

      port "prometheus" {
        to = 8090
      }
    }

    service {
      name = "demo-app-http"
      tags = ["global", "public"]
      port = "http"
      # check {
      #   name     = "alive"
      #   type     = "tcp"
      #   interval = "10s"
      #   timeout  = "2s"
      # }
    }

    service {
      name = "demo-app-prometheus"
      tags = ["prometheus"]
      port = "prometheus"
      # check {
      #   name     = "alive"
      #   type     = "tcp"
      #   interval = "10s"
      #   timeout  = "2s"
      # }
    }

    restart {
      attempts = 2
      interval = "30m"
      delay = "15s"
      mode = "fail"
    }

    task "demo-app" {
      driver = "docker"

      config {
        image = "wfaler/demo-app:v10"
        ports = ["http", "prometheus"]
      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 256 # 256MB
      }

    }
  }
}
