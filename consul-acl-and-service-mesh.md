[instructions here](https://learn.hashicorp.com/tutorials/consul/access-control-setup-production)

[nomad here](https://learn.hashicorp.com/tutorials/nomad/consul-service-mesh) - create policy and token

## In short
Can this be automated by running it locally, pointing at remote Consul?

### Consul
* Bootstrap ACL: `consul acl bootstrap`, make note of `SecretID`
* `export CONSUL_HTTP_TOKEN=[SECRETID]`
* `consul acl policy create -name consul-policies -rules @policies/consul-policies.hcl` (consul-policies should be hostnames of each server)
* Foreach client: `consul acl token create -description "agent token"  -policy-name consul-policies`
* Add token to `client.j2`:

```
acl {
  tokens {
    agent  = "<agent token here>"
  }
}
```
* Restart

### Nomad
* `consul acl policy create -name nomad-client -rules @policies/nomad-client-policy.hcl`
* `consul acl policy create -name nomad-server -rules @policies/nomad-server-policy.hcl`
* `consul acl token create -description "agent token"  -policy-name nomad-client -policy-name nomad-server`
* make note of `SecretID`
* Add to nomad config:

```
consul{
  token = "{{NOMAD_CONSUL_TOKEN}}"
}
```
* Restart


AccessorID:       68d7f0cf-a726-9291-dce5-b4c3f2facace
SecretID:         7ce796c3-93c3-860e-9424-80551eab1765
Description:      Bootstrap Token (Global Management)
Local:            false
Create Time:      2022-05-25 20:57:12.789910922 +0000 UTC
Policies:
   00000000-0000-0000-0000-000000000001 - global-management



 root@ubuntu-4gb-nbg1-1:~# consul acl token create -description "10.0.0.5 token" -policy-name consul-policies
AccessorID:       2bd9108b-12e6-a90c-7df2-54ca1800d40d
SecretID:         92756e87-4fb6-a6e7-0097-10501640214f
Description:      10.0.0.5 token
Local:            false
Create Time:      2022-05-25 21:08:24.56135747 +0000 UTC
Policies:
   b6f63fe6-48d6-cd6d-d466-299c9ccf5b41 - consul-policies




 root@ubuntu-4gb-nbg1-1:~# consul acl token create -description "10.0.0.4 token" -policy-name consul-policies
AccessorID:       c6893f9a-cf2b-7409-b92d-7d877159fbbb
SecretID:         52748b72-12f1-8b1b-3d6b-184dd4a70547
Description:      10.0.0.4 token
Local:            false
Create Time:      2022-05-25 21:08:49.303779609 +0000 UTC
Policies:
   b6f63fe6-48d6-cd6d-d466-299c9ccf5b41 - consul-policies



AccessorID:       04ad07ac-b68d-79bc-4881-5594e62382c6
SecretID:         9ff7868d-739a-f421-24c9-755cdbabb0ab
Description:      Nomad Demo Agent Token
Local:            false
Create Time:      2022-05-25 21:30:09.49065111 +0000 UTC
Policies:
   24deea85-7800-68f3-d85a-3446b002a338 - nomad-server
   a281cd0f-d528-3cf6-9355-0543f502d1bc - nomad-client