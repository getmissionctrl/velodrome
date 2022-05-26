[{% for result in hostvars[inventory_hostname]["groups"]["consul_servers"] %}
  node "{{result}}" {
    policy = "write"
  }
{% endfor %}]
[{% for result in hostvars[inventory_hostname]["groups"]["clients"] %}
  node "{{result}}" {
    policy = "write"
  }
{% endfor %}]
