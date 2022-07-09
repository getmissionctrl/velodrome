#!/bin/bash
LINES=$(ls prom* | wc -l)

if [ "$LINES" == "1" ]; then
  wget https://github.com/grafana/loki/releases/download/v2.5.0/promtail-linux-amd64.zip
  unzip promtail-linux-amd64.zip
  mv promtail-linux-amd64 /usr/local/bin/promtail
fi
