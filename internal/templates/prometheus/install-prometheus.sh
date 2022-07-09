#!/bin/bash
rm -rf prom*
wget https://github.com/prometheus/prometheus/releases/download/v2.36.2/prometheus-2.36.2.linux-amd64.tar.gz
tar xvf prometheus-2.36.2.linux-amd64.tar.gz
cp prometheus-2.36.2.linux-amd64/prometheus /usr/local/bin/prometheus 
cp prometheus-2.36.2.linux-amd64/promtool /usr/local/bin/promtool
  