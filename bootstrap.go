package main

import (
	"log"

	consulApi "github.com/hashicorp/consul/api"
	"redserenity.com/consul-bootstrap/bootstrap"
	"redserenity.com/consul-bootstrap/consul"
)

func BootstrapCommon(config *consulApi.Config, retries, delay int) (*consul.ConsulClient, *consulApi.ACLToken) {
	consulClient := ConnectConsulServer(config, retries, delay)

	var bootstrapAclToken *consulApi.ACLToken

	// If a token is provided, then we will skip bootstrapping and continue on...
	if *bootstrapToken != "" {
		log.Printf("==> Token provided. Skipping ACL Bootstrap...")
		consulClient.Token = *bootstrapToken
	} else {
		bootstrapAclToken = bootstrap.Bootstrap(consulClient, retries, delay)
		// If it's nil, then bootstrap has already happened.
		if bootstrapAclToken != nil {
			consulClient.Token = bootstrapAclToken.SecretID
		}
	}

	return consulClient, bootstrapAclToken
}

func BootstrapServer(client *consul.ConsulClient, bootstrapAclToken *consulApi.ACLToken) {
	if bootstrapAclToken != nil {
		bootstrap.SaveBootstrapKey(client, "self", bootstrapAclToken)
	}

	bootstrap.SetupAnonPolicies(client)

	nodeToken := bootstrap.SetupNodePolicy(client, *consulNodeName, *consulNodePrefix)
	bootstrap.UpdateAclConfig(nodeToken, *consulConfigDir, "acl.hcl")

	regToken := bootstrap.SetupRegisterToken(client)
	bootstrap.SaveRegisterToken(regToken, "", *zeroConfDir, "zeroconf.json")
	log.Printf("==> (Sensitive) Service Registration Token = %s", regToken.SecretID)

	bootstrap.SetupClusterKV(client)
	bootstrap.LockDownNodeJoining(*consulConfigDir, "gossip.hcl")
	log.Printf("==> Bootstrapping complete! A restart may be required for all ACL configurations to work.")
}

func BootstrapCluster(client *consul.ConsulClient, bootstrapAclToken *consulApi.ACLToken, retries, delay int) {
	if bootstrapAclToken != nil {
		zeroConfConsul := ConnectZeroConfServer(*zeroConfAddress, *zeroConfToken, retries, delay)
		bootstrap.SaveBootstrapKey(zeroConfConsul, "cluster", bootstrapAclToken)
	}

	bootstrap.SetupAnonPolicies(client)
	log.Printf("==> Bootstrapping complete! A restart may be required for all ACL configurations to work.")
}

func RegisterZeroConfNode(config *consulApi.Config, retries, delay int) {
	if *zeroConfAddress == "" || *zeroConfToken == "" {
		log.Fatal("-zeroconf-address and -zeroconf-token are required.")
	}

	consulClient := ConnectConsulServer(config, retries, delay)
	zeroConfClient := ConnectZeroConfServer(*zeroConfAddress, *zeroConfToken, retries, delay)

	if !consul.PolicyExistsByName(consulClient, *consulNodePrefix+*consulNodeName) {
		log.Printf("==> Registering Node (%s) with ZeroConf Server...", *consulNodeName)
		nodeToken := bootstrap.SetupNodePolicy(consulClient, *consulNodeName, *consulNodePrefix)
		err := consul.SaveKV(zeroConfClient, "cluster/nodes/"+*consulNodeName+"/token", nodeToken.SecretID)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("==> Registered node %s with ZeroConf Server", *consulNodeName)
	}

	service := &consulApi.AgentServiceRegistration{
		ID:   bootstrap.SanitizeNodeName(*consulNodeName),
		Name: "consul-cluster",
		Port: 8500,
	}

	agentClient := zeroConfClient.Client.Agent()

	if err := agentClient.ServiceRegister(service); err != nil {
		log.Fatal(err)
	}

	log.Printf("==> Registered service %s with ZeroConf Server", bootstrap.SanitizeNodeName(*consulNodeName))
}

func DeregisterZeroConfNode(config *consulApi.Config, retries, delay int) {
	if *zeroConfAddress == "" || *zeroConfToken == "" {
		log.Fatal("-zeroconf-address and -zeroconf-token are required.")
	}

	zeroConfClient := ConnectZeroConfServer(*zeroConfAddress, *zeroConfToken, retries, delay)
	agentClient := zeroConfClient.Client.Agent()

	if err := agentClient.ServiceDeregister(bootstrap.SanitizeNodeName(*consulNodeName)); err != nil {
		log.Fatal(err)
	}

	log.Printf("==> Node %s deregistered with ZeroConf Server", *consulNodeName)
}

func ConnectConsulServer(config *consulApi.Config, retries, delay int) *consul.ConsulClient {
	client, err := consul.ConnectConsulWithRetry(config, retries, delay)
	if err != nil {
		log.Fatalf("==> Unable to connect to Consul server %s after %d tries. Giving up.", config.Address, retries)
	}

	consulClient := &consul.ConsulClient{
		Client:     client,
		Namespace:  config.Namespace,
		Datacenter: config.Datacenter,
		Token:      config.Token,
	}

	return consulClient
}

func ConnectZeroConfServer(address, token string, retries, delay int) *consul.ConsulClient {
	config := consulApi.DefaultConfig()
	config.Address = address
	config.Token = token

	client, err := consul.ConnectConsulWithRetry(config, retries, delay)
	if err != nil {
		log.Fatalf("==> Unable to connect to Consul server %s after %d tries. Giving up.", config.Address, retries)
	}

	consulClient := &consul.ConsulClient{
		Client:     client,
		Namespace:  config.Namespace,
		Datacenter: config.Datacenter,
		Token:      config.Token,
	}

	return consulClient
}
