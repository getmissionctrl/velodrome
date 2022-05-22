# Venue Cluster
Sets up Consul & Nomad Servers & Clients given an inventory.


## Instructions
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
ansible-playbook setup.yml -i inventory -u root
```

## TODO
- [ ] Harden servers
    - [x] Add SSH Key login
    - [x] Setup UFW firewall rules
    - [ ] Template to allow hosts in cluster access to all ports
    - [x] Restart firewall
    - [x] Disable password login
    - [x] Run firewall script
- [x] Install all required software
- [ ] Consul setup
    - [ ] Setup cluster secrets
    - [ ] Template configs
    - [ ] Add configs to cluster
    - [ ] Systemctl script & startup
- [ ] Nomad setup
    - [ ] Setup cluster secrets
    - [ ] Template configs
    - [ ] Add configs to cluster
    - [ ] Systemctl scripts and startup