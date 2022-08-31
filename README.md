# Venue Cluster
Sets up Consul & Nomad Servers & Clients given an inventory.

## Instructions for repo
Your machine/operator node will need the following pre-installed:
* `nomad`
* `consul`
* `vault`
* `ansible`
* `cfssl` & `cfssljson`


`git secret`-usage:
```
# add file
git secret add [file]
# make available to user
git secret tell [email of user in gnupg keychain]
# hide secrets
git secret hide
# show secret
git secret reveal
# other to see commands
git secret
```


## Instructions on new server
When a new server is added:
* Add to appropriate inventory place
* Add SSH key:

```
ssh-copy-id root@[ip of server]
```

* Disable password login:

```
vi /etc/ssh/sshd_config
```

Find following sections and set to no:

```
ChallengeResponseAuthentication no
PasswordAuthentication no
UsePAM no
```

Reload ssh:
```
/etc/init.d/ssh reload
sudo systemctl reload ssh
```

Run ansible:

```
 ansible-playbook setup.yml -i datacenters/contabo/inventory -u root -e @secrets/secrets.yml
```

## TODO
- [x] Harden servers
    - [x] Add SSH Key login
    - [x] Setup UFW firewall rules
    - [x] Template to allow hosts in cluster access to all ports
    - [x] Restart firewall
    - [x] Disable password login
    - [x] Run firewall script
- [x] Install all required software
- [x] Consul setup
    - [x] Setup cluster secrets
    - [x] Template configs
    - [x] Add configs to cluster
    - [x] Systemctl script & startup
    - [x] Verify cluster setup
    - [x] Automate consul ACL bootstrap
    - [x] Allow anonymous DNS access and default Consul as DNS for `.consul` domains
- [x] Nomad setup
    - [x] Setup cluster secrets
    - [x] Template configs
    - [x] Add configs to cluster
    - [x] Systemctl scripts and startup
- [x] Nomad & consul bootstrap expects based on inventory 
- [x] Vault setup
    - [x] setup cluster secrets
    - [x] template configs
    - [x] Systemctl script & startup
    - [x] Auto-unlock with script/ansible/terraform
    - [x] Integrate with Nomad
- [x] Observability
    - [x] Server health
        - [x] CPU monitor
        - [x] RAM usage monitor
        - [x] HD usage monitor  
    - [x] Nomad metrics
    - [x] Consul metrics
    - [x] Log aggregation of jobs
    - [x] Metrics produced by jobs
    - [x] Job tracing
    - [x] Host monitoring (disk, cpu, memory)
- [x] Networking
    - [x] Understand service mesh/ingress etc from consul
    - [x] Ingress to outside world with http/https
    - [x] Use consul as DNS
    - [x] Pull private docker images
    - [x] Observability ingress
    - [x] Auto-accept server signatures on first time connect
- [x] Overall setup
    - [x] Terraform var generation
    - [x] Generate Ansible inventory from Terraform output
- [ ] Grafana/Dashboards
    - [ ] Dashboards
        - [ ] Consul health
        - [ ] Nomad health
        - [ ] Vault health
        - [x] Host health
    - [ ] SLO templates
        - [ ] Web/api service
        - [ ] Headless backend service
    - [ ] Alerts
        - [ ] Consul health
        - [ ] Nomad health
        - [ ] Vault health
        - [ ] Host health (CPU, memory, disk)

# Kill orphaned nomad mounts

```
export NOMAD_DATA_ROOT=«Path to your Nomad data_dir»

for ALLOC in `ls -d $NOMAD_DATA_ROOT/alloc/*`; do for JOB in `ls ${ALLOC}| grep -v alloc`; do umount ${ALLOC}/${JOB}/secrets; umount ${ALLOC}/${JOB}/dev; umount ${ALLOC}/${JOB}/proc; umount ${ALLOC}/${JOB}/alloc; done; done
```
# Cert for client access with browser.
```
openssl pkcs12 -inkey consul-agent-ca-key.pem -in consul-agent-ca.pem -export -out consul.pfx
```
# Notes

### Scrape consul
```
    - job_name: integrations/consul
      metrics_path: /v1/agent/metrics
      params:
        format:
        - prometheus
      static_configs:
      - targets:
        - {{ private_ip }}:8500
```

# O11y setup

## Tempo search in Grafana

Edit `/etc/grafana/grafana.ini`, add:

```
[feature_toggles]
enable = tempoSearch tempoBackendSearch

```

## Link Loki to Tempo traces

Add derived fields:

```
Name: trace_id
Regex: "trace_id":"([A-Za-z0-9]+)" // this is for json format
Query: ${__value.raw}
Url label: Trace
Internal link: Tempo

```


## Hetzner cloud volume pattern
```
/mnt/HC_Volume_21865747

```
numbers are the assigned volume ID, you can get it in terraform.



openssl req -new -newkey rsa:4096 -x509 -sha256 -nodes -out vault.crt -keyout vault.key 



openssl genrsa -aes256 -out vaultCA.key 2048


openssl req -key vaultCA.key -new -out domain.csr



openssl req -key vaultCA.key rsa:2048 -nodes -keyout domain.key -x509 -days 365 -out domain.crt

Must use Homebrew openssl:
openssl req -out tls.crt -new -keyout tls.key -newkey rsa:4096 -nodes -sha256 -x509 -subj "/O=HashiCorp/CN=Vault" -addext "subjectAltName = IP:0.0.0.0,DNS:vault.service.consul,DNS:venue-vault-1,DNS:venue-vault-2" -days 3650

88.99.172.159 host_name=venue-vault-1 private_ip=10.0.2.2
78.46.128.124 host_name=venue-vault-2 private_ip=10.0.2.3


# Vault setup
Generate TLS keys - must be with homebrew or Nix openssl

```
openssl req -out tls.crt -new -keyout tls.key -newkey rsa:4096 -nodes -sha256 -x509 -subj "/O=HashiCorp/CN=Vault" -addext "subjectAltName = IP:0.0.0.0,DNS:vault.service.consul,DNS:venue-vault-1,DNS:venue-vault-2"
```

run `vault operator init`

run `vault operator unseal` on each vault node  

To be able to store secrets:
`vault secrets enable -path=secret/ kv`

Danach:

` vault kv put secret/hello foo=world  `
` vault kv get secret/hello`



https://www.cloudflare.com/ips-v6
https://www.cloudflare.com/ips-v4


```
vault policy write nomad-server internal/templates/vault/nomad-server-policy.hcl

vault token create -policy nomad-server -period 72h -orphan > nomad.txt

vault write /auth/token/roles/nomad-cluster @internal/templates/vault/token-role.json
```

token into `nomad_root_token` of secrets


to make specific app policies:

`access.hcl`
```
path "secret/*" { #some path in secrets
    capabilities = ["read"]
}```

```
vault policy write backend access.hcl
```

in `nomad task definition`:
```
      vault {
        policies = ["backend"] # policy given above

        change_mode   = "signal"
        change_signal = "SIGUSR1"
      }
```