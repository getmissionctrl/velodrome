# Venue Cluster
Sets up Consul & Nomad Servers & Clients given an inventory.

## Instructions for repo
You need git secret

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
- [ ] Consul setup
    - [x] Setup cluster secrets
    - [x] Template configs
    - [x] Add configs to cluster
    - [x] Systemctl script & startup
    - [ ] Verify cluster setup
- [ ] Nomad setup
    - [ ] Setup cluster secrets
    - [ ] Template configs
    - [ ] Add configs to cluster
    - [ ] Systemctl scripts and startup