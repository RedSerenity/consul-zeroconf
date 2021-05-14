package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	consulApi "github.com/hashicorp/consul/api"
)

func (client *ConsulClient) WriteOpts() *consulApi.WriteOptions {
	return &consulApi.WriteOptions{
		Namespace:  client.Namespace,
		Datacenter: client.Datacenter,
		Token:      client.Token,
	}
}

func (client *ConsulClient) QueryOpts() *consulApi.QueryOptions {
	return &consulApi.QueryOptions{
		Namespace:  client.Namespace,
		Datacenter: client.Datacenter,
		Token:      client.Token,
	}
}

/* Bootstrap Functions */

func ConnectConsul(config *consulApi.Config) (*consulApi.Client, error) {
	log.Printf("==> Connecting to Consul Server running at %s", config.Address)
	client, err := consulApi.NewClient(config)
	if err != nil {
		return nil, err
	}

	statusClient := client.Status()
	_, err = statusClient.Leader()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ConnectConsulWithRetry(config *consulApi.Config, retries, delay int) (*consulApi.Client, error) {
	count := 0

	for {
		count++

		if count > retries {
			return nil, errors.New(fmt.Sprintf("Unable to connect to Consul server running at %s after %d attempts. Giving up.", config.Address, retries))
		}

		client, err := ConnectConsul(config)
		if err != nil {
			log.Printf("==> Unable to connect to Consul server running at %s. Pausing for %d seconds. Try %d of %d.", config.Address, delay, count, retries)
			time.Sleep(time.Duration(delay) * time.Second)
			continue
		}

		return client, nil
	}
}

// BootstrapAcl Returns ACL Token, AlreadyBootstrapped, Error
func BootstrapAcl(consulClient *consulApi.Client, retries, delay int) (*consulApi.ACLToken, bool, error) {
	aclClient := consulClient.ACL()

	count := 0
	for {
		count++

		if count > 10 {
			log.Printf("==> Unable to bootstrap server after %d attempts. Giving up.", count)
			return nil, false, errors.New("unable to bootstrap server")
		}

		token, _, err := aclClient.Bootstrap()
		if err != nil {
			if strings.Contains(err.Error(), "ACL bootstrap no longer allowed") {
				log.Printf("==> Server already bootstrapped.")
				return nil, true, err
			}

			if strings.Contains(err.Error(), "ACL support disabled") {
				log.Printf("==> Server ACL not enabled. Add 'acl { enabled = true }' to your config file and try again.")
				return nil, false, err
			}

			if strings.Contains(err.Error(), "ACL system is currently in legacy mode") {
				log.Printf("==> Server ACL not ready. Waiting %d seconds. Try %d of %d.", delay, count, retries)
				time.Sleep(time.Duration(delay) * time.Second)
				continue
			}
		}

		return token, false, nil
	}
}

/* Policy Functions */

func PolicyExistsByName(client *ConsulClient, policyName string) bool {
	policy, err := GetPolicyByName(client, policyName)
	if policy == nil || err != nil {
		return false
	}

	return true
}

func PolicyExistsById(client *ConsulClient, policyId string) bool {
	policy, err := GetPolicyById(client, policyId)
	if policy == nil || err != nil {
		return false
	}

	return true
}

func GetPolicyById(client *ConsulClient, policyId string) (*consulApi.ACLPolicy, error) {
	aclClient := client.Client.ACL()

	policy, _, err := aclClient.PolicyRead(policyId, client.QueryOpts())
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func GetPolicyByName(client *ConsulClient, policyName string) (*consulApi.ACLPolicy, error) {
	aclClient := client.Client.ACL()

	policy, _, err := aclClient.PolicyReadByName(policyName, client.QueryOpts())
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func CreatePolicy(client *ConsulClient, name, description, rules string) (*consulApi.ACLPolicy, error) {
	aclClient := client.Client.ACL()

	policy := &consulApi.ACLPolicy{
		Name: name,
		Description: description,
		Rules: rules,
	}

	policy, _, err := aclClient.PolicyCreate(policy, client.WriteOpts())
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func GroupAclPolicies(policies []*consulApi.ACLPolicy) []*consulApi.ACLTokenPolicyLink {
	var aclPolicies []*consulApi.ACLTokenPolicyLink

	for _, policy := range policies {
		aclPolicies = append(aclPolicies, &consulApi.ACLTokenPolicyLink{Name: policy.Name})
	}

	return aclPolicies
}

/* Token Functions */

func GetToken(client *ConsulClient, tokenId string) (*consulApi.ACLToken, error) {
	aclClient := client.Client.ACL()

	token, _, err := aclClient.TokenRead(tokenId, client.QueryOpts())
	if err != nil {
		return nil, err
	}

	return token, nil
}

func UpdateTokenPolicies(client *ConsulClient, tokenId string, policyNames []string) (*consulApi.ACLToken, error) {
	aclClient := client.Client.ACL()

	token, err := GetToken(client,tokenId)
	if err != nil {
		return nil, err
	}

	var policies []*consulApi.ACLTokenPolicyLink
	for _, link := range token.Policies {
		policies = append(policies, link)
	}
	for _, policy := range policyNames {
		policies = append(policies, &consulApi.ACLTokenPolicyLink{Name: policy})
	}

	token.Policies = policies

	updatedToken, _, err := aclClient.TokenUpdate(token, client.WriteOpts())
	if err != nil {
		return nil, err
	}

	return updatedToken, nil
}

func CreateToken(client *ConsulClient, description string, policies []*consulApi.ACLTokenPolicyLink) (*consulApi.ACLToken, error) {
	aclClient := client.Client.ACL()

	aclToken := &consulApi.ACLToken{
		Description: description,
		Policies:    policies,
	}

	token, _, err := aclClient.TokenCreate(aclToken, client.WriteOpts())
	if err != nil {
		return nil, err
	}

	return token, nil
}

func CreatePolicyToken(client *ConsulClient, description string, policy *consulApi.ACLPolicy) (*consulApi.ACLToken, error) {
	aclClient := client.Client.ACL()

	var policies []*consulApi.ACLTokenPolicyLink
	policies = append(policies, &consulApi.ACLTokenPolicyLink{
		Name: policy.Name,
	})

	aclToken := &consulApi.ACLToken{
		Description: description,
		Policies:    policies,
	}

	token, _, err := aclClient.TokenCreate(aclToken, client.WriteOpts())
	if err != nil {
		return nil, err
	}

	return token, nil
}

func GetSecret(token *consulApi.ACLToken) string {
	return token.SecretID
}

/* KV Functions */

func GetKV(client *ConsulClient, key string) (string, error) {
	kvClient := client.Client.KV()

	pair, _, err := kvClient.Get(key, client.QueryOpts())
	if err != nil {
		return "", err
	}

	return string(pair.Value), nil
}

func SaveKV(client *ConsulClient, key string, value string) error {
	kvClient := client.Client.KV()

	kvPair := &consulApi.KVPair{Key: key, Value: []byte(value)}
	_, err := kvClient.Put(kvPair, client.WriteOpts())
	if err != nil {
		return err
	}

	return nil
}

func SaveKVStruct(client *ConsulClient, key string, value interface{}) error {
	kvClient := client.Client.KV()

	valueSerialized, err := json.MarshalIndent(value, "", "\t")
	if err != nil {
		return err
	}

	kvPair := &consulApi.KVPair{Key: key, Value: valueSerialized}
	_, err = kvClient.Put(kvPair, client.WriteOpts())
	if err != nil {
		return err
	}

	return nil
}
