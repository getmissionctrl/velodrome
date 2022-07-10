#!/bin/bash

rm tempo*
wget https://github.com/grafana/tempo/releases/download/v1.4.1/tempo_1.4.1_linux_amd64.tar.gz
gunzip tempo_1.4.1_linux_amd64.tar.gz
tar xvf tempo_1.4.1_linux_amd64.tar
mv tempo /usr/local/bin/tempo
chmod 755 /usr/local/bin/tempo

