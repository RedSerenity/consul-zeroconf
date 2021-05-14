package consul

import consulApi "github.com/hashicorp/consul/api"

type ConsulClient struct {
	Client     *consulApi.Client
	Token      string
	Datacenter string
	Namespace  string
}
