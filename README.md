# Venue Cluster
Sets up Consul & Nomad Servers & Clients given an inventory.

## Instructions for repo
Your machine/operator node will need the following pre-installed:
* `nomad`
* `consul`
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

## Instructions on setup on existing servers

```
 ansible-playbook setup.yml -i datacenters/contabo/inventory -u root -e @secrets/secrets.yml
```

On AWS: 

```
 ansible-playbook setup.yml -i datacenters/aws/inventory -u ubuntu -e @secrets/secrets.yml --private-key venue-dev.pem
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
- [ ] Observability
    - [ ] Server health
        - [ ] CPU monitor
        - [ ] RAM usage monitor
        - [ ] HD usage monitor
    - [ ] Log aggregation of jobs
    - [ ] Nomad metrics
    - [ ] Prometheus metrics
    - [ ] Job tracing
- [ ] Networking
    - [x] Understand service mesh/ingress etc from consul
    - [ ] Ingress to outside world with http/https
    - [ ] Pull private docker images
- [ ] Vault setup
    - [ ] setup cluster secrets
    - [ ] template configs
    - [ ] Systemctl script & startup
    - [ ] Auto-unlock with script/ansible/terraform
    - [ ] Integrate with Nomad


# Kill orphaned nomad mounts

```
export NOMAD_DATA_ROOT=«Path to your Nomad data_dir»

for ALLOC in `ls -d $NOMAD_DATA_ROOT/alloc/*`; do for JOB in `ls ${ALLOC}| grep -v alloc`; do umount ${ALLOC}/${JOB}/secrets; umount ${ALLOC}/${JOB}/dev; umount ${ALLOC}/${JOB}/proc; umount ${ALLOC}/${JOB}/alloc; done; done
```
# Cert for client access with browser.
```
openssl pkcs12 -inkey consul-agent-ca-key.pem -in consul-agent-ca.pem -export -out consul.pfx
```