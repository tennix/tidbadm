package k8s

import (
	"github.com/pingcap/tidbadm/runtime"
)

var (
	defaultCount   int32 = 1
	defaultVersion       = "v3.0.5"
)

type k8s struct {
	kubeconfig string
}

func New() k8s {
	return k8s{}
}

func (r k8s) Init() (*runtime.TidbCluster, error) {
	var deploy = runtime.TidbCluster{
		APIVersion: "pingcap.com/v1",
		Kind:       "TidbCluster",
		Metadata: runtime.Metadata{
			Name: "my-cluster",
		},
		Spec: runtime.Spec{
			Version: &defaultVersion,
			Runtime: runtime.K8sRuntime,
			PD: runtime.PDSpec{
				Count: &defaultCount,
			},
			TiKV: runtime.TiKVSpec{
				Count: &defaultCount,
			},
			TiDB: &runtime.TiDBSpec{
				Count: &defaultCount,
			},
		},
	}
	return &deploy, nil
}

func (r k8s) Plan(tc runtime.TidbCluster) error {
	return nil
}

func (r k8s) Apply(tc runtime.TidbCluster) error {
	return nil
}

func (r k8s) Destroy(tc runtime.TidbCluster) error {
	return nil
}
