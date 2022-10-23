# Velodrome
Sets up Consul, Nomad & Vault Servers & Clients given an Cloud specific config, with full observability via Grafana, Prometheus, Loki & Tempo + [Fabio LB](https://fabiolb.net) as an ingress. 
Currently supports [Hetzner](https://www.hetzner.com), AWS, Azure & GCP coming soon.

Networking is setup so your bastion host (`allowed_ips`) have full access to the cluster, while Cloudflare IPs have access to ports 80 and 443.

# Pre-requisites
## Pre-requisite software
Your machine/operator node will need the following pre-installed (velodrome will check for presence before execution):
* `nomad`
* `consul`
* `vault`
* `ansible`
* `cfssl` & `cfssljson`

You probably want to also use [git secret](https://git-secret.io) to protect your `[base_dir]/secrets` directory in the generated files.
Additionally, [direnv](https://direnv.net) will make life easier, as `velodrome genenv --config.file [config]` will generate a direnv compatible `.envrc` file for you.

To ensure other env variables are preserved with `velodrome genenv`, just add this line into an existing `.envrc` file:

```
### GENERATED CONFIG BELOW THIS LINE, DO NOT EDIT!
```

## Other requirements
* An SSL certificate with cert, key and ca-file. Can easily be generated with for instance `Cloudflare` (network setup has been tested primarily with Cloudflare)
* An SSH key and project already setup in Hetzner (when using Hetzner).
* The following 4 environment variables set in your environment (S3 settings can be any S3 compatible store, including Cloudflare R2, this is used for Observability stack long-term storage):
    * `S3_ENDPOINT`
    * `S3_ACCESS_KEY`
    * `S3_SECRET_KEY`
    * `HETZNER_TOKEN` (generated from your Hetzner account)
* a `config.yaml` file. Please review the file with similar name in the root of this directory for options. Ensure that the IP of your machine/bastion host is in the `allowed_ips` section.

## Setup
Once all of the above steps are setup, just run `velodrome sync --config.file [config file]`. If no cluster exists, it will be setup for you. If one exists, it will be synced with your config, setting up the entire cluster.

# Non-automated steps requiring manual steps

## Add data sources in Grafana

```
Loki: http://loki-http.service.consul:3100

Prometheus: http://prometheus.service.consul:9000

Tempo: http://tempo.service.consul:3200
```

## Add Nomad & Node dashboards
* Nomad: add dashboard ID 15764
* Nodes: add dashboard ID 12486

## Link Loki to Tempo traces

Add derived fields:

```
Name: trace_id
Regex: "trace_id":"([A-Za-z0-9]+)" // this is for json format
Query: ${__value.raw}
Url label: Trace
Internal link: Tempo

```

## Load Balancing
Load balancing has been tested with Cloudflare Load Balancer. Simply put the public IPs of your client nodes (found in `config/inventory`) into a Cloudflare LB. This can be automated with Terraform, or simply done manually through the Cloudflare interface.

To make DNS work, you will need your root-domain setup, as well as CNAMEs for any additional domains or subdomains.

By default, the cluster will try to setup the following:
* `grafana.[your management_domain]` (you still need to setup DNS with Cloudflare)
    * Once public, please change the default password immediately!
* `consul.[your management_domain]` (you still need to setup DNS with Cloudflare)
    * Username/password will be `consul` and `CONSUL_HTTP_TOKEN` you get from `velodrome genenv`
## Nomad jobs, using consul & vault
There are examples in the `examples/` folder of this repo.
### Use vault from nomad job (example)
to make specific app policies:

`access.hcl`
```
path "secret/*" { #some path in secrets
    capabilities = ["read"]
}
```

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

## Kill orphaned nomad mounts if killing a client node

```
export NOMAD_DATA_ROOT=«Path to your Nomad data_dir»

for ALLOC in `ls -d $NOMAD_DATA_ROOT/alloc/*`; do for JOB in `ls ${ALLOC}| grep -v alloc`; do umount ${ALLOC}/${JOB}/secrets; umount ${ALLOC}/${JOB}/dev; umount ${ALLOC}/${JOB}/proc; umount ${ALLOC}/${JOB}/alloc; done; done
```


# TODO
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