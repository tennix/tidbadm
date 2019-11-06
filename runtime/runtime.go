package runtime

type Metadata struct {
	Name            string `yaml:"name"`
	ResourceVersion uint64 `yaml:"resourceVersion"`
}

type Runtime string
type Member string

const (
	K8sRuntime           Runtime = "k8s"
	AnsibleRuntime       Runtime = "ansible"
	PDMember             Member  = "pd"
	TiKVMember           Member  = "tikv"
	TiDBMember           Member  = "tidb"
	DockerRegistryPrefix         = "pingcap"
	BinaryURL                    = "https://download.pingcap.org/tidb-%s-linux-amd64.tar.gz"
)

type Spec struct {
	bootstrapped bool
	Runtime      Runtime   `yaml:"runtime"`
	Version      *string   `yaml:"version,omitempty"`
	PD           PDSpec    `yaml:"pd"`
	TiKV         TiKVSpec  `yaml:"tikv"`
	TiDB         *TiDBSpec `yaml:"tidb,omitempty"`
}

type PDSpec struct {
	Count *int32 `yaml:"count,omitempty"`
	Nodes []Node `yaml:"nodes,omitempty"`
}

type TiKVSpec struct {
	Count *int32 `yaml:"count,omitempty"`
	Nodes []Node `yaml:"nodes,omitempty"`
}
type TiDBSpec struct {
	Count *int32 `yaml:"count,omitempty"`
	Nodes []Node `yaml:"nodes,omitempty"`
}

type Node struct {
	Host    string  `yaml:"host"`
	DataDir *string `yaml:"data-dir,omitempty"`
}

type TidbCluster struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
	Status     Status   `yaml:"status,omitempty"`
}

type Status struct {
	TiDBEndpoints   string `yaml:"tidb-endpoints"`
	GrafanaEndpoint string `yaml:"grafana-endpoint,omitempty"`
}

type Runner interface {
	Init() (*TidbCluster, error)
	Plan(TidbCluster) error
	Apply(TidbCluster) error
	Destroy(TidbCluster) error
}
