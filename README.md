# Consul Zero Configuration

![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/RedSerenity/consul-zeroconf?label=Version&style=for-the-badge)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/RedSerenity/consul-zeroconf/DockerBuildPush?label=Docker%20Build&style=for-the-badge)
![Docker Stars](https://img.shields.io/docker/stars/redserenity/consul-zeroconf?style=for-the-badge)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/redserenity/consul-zeroconf?sort=date&style=for-the-badge)

---

_* Work in Progress *_

<p>This application makes it easy to spin up a Consul cluster with almost zero manual commands.</p>
<p>Start up a single Consul server to coordinate bootstrapping an entire Consul cluster.</p>
<p align="center">
	<img src="https://github.com/redserenity/consul-zeroconf/blob/master/architecture.png?raw=true" />
</p>

<br/>

**Bootstrap ZeroConf Server**
```shell
consul-zeroconf -bootstrap-server -address=http://server.consul:8500 -config-dir="/consul/config"
```

**Bootstrap ZeroConf Cluster**
```shell
consul-zeroconf -bootstrap-cluster -address=http://node0.consul:8500 -config-dir="/consul/config" -zeroconf-address=http://server.consul:8500 -zeroconf-token=<token from previous command>
```

**Command Line Arguments**
```shell
  -address string
        Consul Address (e.g. http://localhost:8500) (default "http://localhost:8500")
  -bootstrap-cluster
        Bootstrap the ZeroConf Cluster
  -bootstrap-server
        Bootstrap the ZeroConf Server
  -bootstrap-token string
        Consul Bootstrap Token
  -config-dir string
        Consul config directory (default "/consul/config/")
  -connect-delay int
         (default 5)
  -connect-retries int
        Number of times to retry connecting to Consul. (default 10)
  -deregister-node
        Deregister the node from the ZeroConf Server.
  -node-name string
        Consul Node Name
  -node-prefix string
        Policy prefix for node name (default "Node-")
  -register-node
        Register the node with the ZeroConf Server.
  -version
        Display program version
  -zeroconf-address string
        ZeroConf Server address
  -zeroconf-dir string
        ZeroConf directory (default "/consul/zeroconf")
  -zeroconf-token string
        ZeroConf Server token used for Service Registration
```

