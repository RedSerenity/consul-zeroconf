package templates

const ANON_POLICY = `node_prefix "" {
	policy = "read"
}
service_prefix "" {
	policy = "read"
}
key_prefix "" {
	policy = "deny"
}
`

const NODE_POLICY = `node "{{.Name}}" {
  policy = "write"
}
agent "{{.Name}}" {
  policy = "write"
}
service_prefix "" {
  policy = "read"
}
key_prefix "_rexec" {
  policy = "write"
}
`

const REGISTRATION_POLICY = `service "consul-cluster" {
	policy = "write"
}
key_prefix "cluster/nodes" {
  policy = "write"
}
service_prefix "" {
	policy = "read"
}
node_prefix "" {
	policy = "read"
}
`

const ACL_CONFIG = `acl {
  enabled                  = true
  default_policy           = "deny"
  enable_token_persistence = true

  tokens {
    agent = "{{.SecretID}}"
  }
}`
