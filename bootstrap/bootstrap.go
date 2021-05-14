package bootstrap

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"

	consulApi "github.com/hashicorp/consul/api"
	"redserenity.com/consul-bootstrap/config"
	"redserenity.com/consul-bootstrap/consul"
	"redserenity.com/consul-bootstrap/templates"
)

func Bootstrap(consulClient *consul.ConsulClient, retries, delay int) *consulApi.ACLToken {
	token, isBootstrapped, err := consul.BootstrapAcl(consulClient.Client, retries, delay)
	if err != nil {
		log.Fatal(err)
	}
	if isBootstrapped {
		log.Print("==> System is already Bootstrapped. Add -bootstrap-token argument to bypass bootstrapping and setup policies instead.")
		return nil
	}

	log.Printf("==> Consul ACL has been bootstrapped.")
	log.Printf("==> (Sensitive) Bootstrap Token = %s", token.SecretID)

	return token
}

func SaveBootstrapKey(client *consul.ConsulClient, key string, bootstrapToken *consulApi.ACLToken) {
	log.Printf("==> Saving Bootstrap token to KV store.")

	if err := consul.SaveKVStruct(client, "bootstrap/"+key+"/complete", bootstrapToken); err != nil {
		log.Fatal(err)
	}

	if err := consul.SaveKV(client, "bootstrap/"+key+"/token", bootstrapToken.SecretID); err != nil {
		log.Fatal(err)
	}
}

func SetupAnonPolicies(client *consul.ConsulClient) {
	log.Printf("==> Updating Anonymous token with sane defaults.")

	policy, err := consul.CreatePolicy(
		client,
		"anon-management",
		"Anonymous Management Policy that grants read-only access to Services & Nodes.",
		templates.ANON_POLICY)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := consul.UpdateTokenPolicies(client, config.ANON_TOKEN, []string{policy.Name}); err != nil {
		log.Fatal(err)
	}
}

func SetupNodePolicy(client *consul.ConsulClient, nodeName string, nodePrefix string) *consulApi.ACLToken {
	log.Printf("==> Creating Node policy for %s.", nodeName)

	safeNodeName := SanitizeNodeName(nodeName)
	template, err := config.GetTemplate("NodePolicy", templates.NODE_POLICY, struct{ Name string }{Name: safeNodeName})
	if err != nil {
		log.Fatal(err)
	}

	policy, err := consul.CreatePolicy(client, nodePrefix+safeNodeName, "Agent Policy for node "+nodeName, template)
	if err != nil {
		log.Fatal(err)
	}

	token, err := consul.CreatePolicyToken(client, "Agent Token for policy "+nodePrefix+safeNodeName, policy)
	if err != nil {
		log.Fatal(err)
	}

	return token
}

func SetupRegisterToken(client *consul.ConsulClient) *consulApi.ACLToken {
	log.Printf("==> Creating registration policy & token.")

	policy, err := consul.CreatePolicy(
		client,
		"cluster-registration",
		"Policy for cluster nodes to register with the ZeroConf server",
		templates.REGISTRATION_POLICY)
	if err != nil {
		log.Fatal(err)
	}

	token, err := consul.CreatePolicyToken(client, "Registration Token for policy cluster-registration", policy)
	if err != nil {
		log.Fatal(err)
	}

	return token
}

func SaveRegisterToken(token *consulApi.ACLToken, address, path, file string) {
	log.Printf("==> Saving registration token in %s%s.", path, file)

	content, err := json.MarshalIndent(&ZeroConf{Address: address, Token: token.SecretID}, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	if err := config.SaveConfig(path, file, string(content)); err != nil {
		log.Fatal(err)
	}
}

func UpdateAclConfig(nodeToken *consulApi.ACLToken, path, file string) {
	log.Printf("==> Updating acl config in %s%s.", path, file)

	template, err := config.GetTemplate("AclConfig", templates.ACL_CONFIG, nodeToken)
	if err != nil {
		log.Fatal(err)
	}

	if err := config.SaveConfig(path, file, template); err != nil {
		log.Fatal(err)
	}
}

func SetupClusterKV(client *consul.ConsulClient) {
	log.Printf("==> Setting up cluster structure in KV store.")

	if err := consul.SaveKV(client, "bootstrap/cluster/complete", "{}"); err != nil {
		log.Fatal(err)
	}

	if err := consul.SaveKV(client, "bootstrap/cluster/token", ""); err != nil {
		log.Fatal(err)
	}

	if err := consul.SaveKVStruct(client, "cluster/nodes/", ""); err != nil {
		log.Fatal(err)
	}

	if err := consul.SaveKV(client, "cluster/gossip_key", GenerateKey()); err != nil {
		log.Fatal(err)
	}
}

func LockDownNodeJoining(path, file string) {
	log.Printf("==> Locking down the node from (possible) rogue nodes.")

	if err := config.SaveConfig(path, file, "encrypt = \""+GenerateKey()+"\""); err != nil {
		log.Fatal(err)
	}
}

func SanitizeNodeName(nodeName string) string {
	return strings.Replace(nodeName, ".", "_", -1)
}

func GenerateKey() string {
	key := make([]byte, 32)
	n, err := rand.Reader.Read(key)
	if err != nil {
		log.Fatal(err)
	}
	if n != 32 {
		log.Fatal("Could not generate enough entropy for GenerateKey() function")
	}
	return base64.StdEncoding.EncodeToString(key)
}
