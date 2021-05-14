package main

import (
	"flag"
	"log"
	"os"
	"strings"

	consulApi "github.com/hashicorp/consul/api"
)

var (
	// Versioning
	appVersion, appCommit, appBuiltAt, appBuiltBy, appBuiltOn string

	// Common
	version          = flag.Bool("version", false, "Display program version")
	consulAddress    = flag.String("address", "http://localhost:8500", "Consul Address (e.g. http://localhost:8500)")
	consulNodeName   = flag.String("node-name", "", "Consul Node Name")
	consulNodePrefix = flag.String("node-prefix", "Node-", "Policy prefix for node name")
	consulConfigDir  = flag.String("config-dir", "/consul/config/", "Consul config directory")

	// ZeroConf Server
	bootstrapServer = flag.Bool("bootstrap-server", false, "Bootstrap the ZeroConf Server")

	// ZeroConf Cluster
	bootstrapCluster = flag.Bool("bootstrap-cluster", false, "Bootstrap the ZeroConf Cluster")
	zeroConfAddress  = flag.String("zeroconf-address", "", "ZeroConf Server address")
	zeroConfToken    = flag.String("zeroconf-token", "", "ZeroConf Server token used for Service Registration")

	// ZeroConf Common
	bootstrapToken = flag.String("bootstrap-token", "", "Consul Bootstrap Token")
	zeroConfDir    = flag.String("zeroconf-dir", "/consul/zeroconf", "ZeroConf directory")

	registerNode   = flag.Bool("register-node", false, "Register the node with the ZeroConf Server.")
	deregisterNode = flag.Bool("deregister-node", false, "Deregister the node from the ZeroConf Server.")

	connectRetries = flag.Int("connect-retries", 10, "Number of times to retry connecting to Consul.")
	connectDelay   = flag.Int("connect-delay", 5, "")
)

func init() {
	flag.Parse()
	HandleVersion()
	HandleEnvVars()
	ErrorCheckParams()
}

func main() {
	consulConfig := consulApi.DefaultConfig()
	consulConfig.Address = *consulAddress

	if *bootstrapServer {
		client, bootstrapAclToken := BootstrapCommon(consulConfig, *connectRetries, *connectDelay)
		if bootstrapAclToken != nil {
			BootstrapServer(client, bootstrapAclToken)
			log.Printf("==> ZeroConf Server bootstrap finished.")
		}
	}

	if *bootstrapCluster {
		client, bootstrapAclToken := BootstrapCommon(consulConfig, *connectRetries, *connectDelay)
		if bootstrapAclToken != nil {
			BootstrapCluster(client, bootstrapAclToken, *connectRetries, *connectDelay)
			log.Printf("==> ZeroConf Cluster bootstrap finished.")
		}
	}

	if *registerNode {
		RegisterZeroConfNode(consulConfig, *connectRetries, *connectDelay)
	}

	if *deregisterNode {
		DeregisterZeroConfNode(consulConfig, *connectRetries, *connectDelay)
	}
}

func HandleVersion() {
	if *version {
		log.Printf("Consul ZeroConf\n  Version: %s\n  Commit: %s\n  Built: %s by %s on %s\n", appVersion, appCommit, appBuiltAt, appBuiltBy, appBuiltOn)
		os.Exit(0)
	}
}

func HandleEnvVars() {
	envConsulAddress := os.Getenv("CONSUL_HTTP_ADDRESS")
	envNodeName := os.Getenv("CONSUL_NODE_NAME")
	envNodePrefix := os.Getenv("CONSUL_NODE_PREFIX")
	envConfigDir := os.Getenv("CONSUL_CONFIG_DIR")

	envZeroConfAddress := os.Getenv("CONSUL_ZEROCONF_ADDRESS")
	envZeroConfToken := os.Getenv("CONSUL_ZEROCONF_TOKEN")

	if envConsulAddress != "" {
		*consulAddress = envConsulAddress
	}

	if envNodeName != "" {
		*consulNodeName = envNodeName
	}

	if envNodePrefix != "" {
		*consulNodePrefix = envNodePrefix
	}

	if envConfigDir != "" {
		*consulConfigDir = envConfigDir
	}

	if envZeroConfAddress != "" {
		*zeroConfAddress = envZeroConfAddress
	}

	if envZeroConfToken != "" {
		*zeroConfToken = envZeroConfToken
	}
}

func ErrorCheckParams() {
	if len(os.Args) == 1 {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !strings.HasSuffix(*consulConfigDir, "/") {
		*consulConfigDir = *consulConfigDir + "/"
	}

	if !strings.HasSuffix(*zeroConfDir, "/") {
		*zeroConfDir = *zeroConfDir + "/"
	}

	if *bootstrapServer && *registerNode {
		log.Fatal("==> Cannot specify both -bootstrap-server and -register-node")
	}

	if *bootstrapServer && *deregisterNode {
		log.Fatal("==> Cannot specify both -bootstrap-server and -deregister-node")
	}

	if *bootstrapServer && *bootstrapCluster {
		log.Fatal("==> Cannot specify both -bootstrap-server and -bootstrap-cluster")
	}

	if *bootstrapCluster && (*zeroConfAddress == "" || *zeroConfToken == "") {
		log.Fatal("==> -zeroconf-address and -zeroconf-token are required when using -bootstrap-cluster. One or both are missing.")
	}

	if *registerNode && (*zeroConfAddress == "" || *zeroConfToken == "") {
		log.Fatal("==> -zeroconf-address and -zeroconf-token are required when using -register-node. One or both are missing.")
	}

	if *deregisterNode && (*zeroConfAddress == "" || *zeroConfToken == "") {
		log.Fatal("==> -zeroconf-address and -zeroconf-token are required when using -deregister-node. One or both are missing.")
	}

	if *consulNodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}

		*consulNodeName = hostname
		log.Printf("==> Defaulting node name to hostname (%s)", hostname)
	} else {
		log.Printf("==> Node name set to %s", *consulNodeName)
	}
}