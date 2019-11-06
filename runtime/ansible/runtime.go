package ansible

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pingcap/tidbadm/runtime"
)

const (
	leftDelim  = "<%"
	rightDelim = "%>"
)

var (
	defaultPDDir    = "/data/pd"
	defaultTiKVDir  = "/data/tikv"
	defaultVersion  = "v3.0.5"
	deployUser      = "tidb"
	deployDirPrefix = "/opt/tidb-clusters/"
)

type ansible struct {
	user string
}

func New() ansible {
	return ansible{}
}

func (r ansible) Init() (*runtime.TidbCluster, error) {
	var deploy = runtime.TidbCluster{
		APIVersion: "pingcap.com/v1",
		Kind:       "TidbCluster",
		Metadata: runtime.Metadata{
			Name: "my-cluster",
		},
		Spec: runtime.Spec{
			Version: &defaultVersion,
			Runtime: runtime.AnsibleRuntime,
			PD: runtime.PDSpec{
				Nodes: []runtime.Node{
					{Host: "127.0.0.1", DataDir: &defaultPDDir},
				},
			},
			TiKV: runtime.TiKVSpec{
				Nodes: []runtime.Node{
					{Host: "127.0.0.1", DataDir: &defaultTiKVDir},
				},
			},
			TiDB: &runtime.TiDBSpec{
				Nodes: []runtime.Node{
					{Host: "127.0.0.1"},
				},
			},
		},
	}
	return &deploy, nil
}

func (r ansible) Plan(tc runtime.TidbCluster) error {
	return nil
}

func (r ansible) Apply(tc runtime.TidbCluster) error {
	if r.isBootstrapped(tc) {
		return nil
	}
	downloadURL := fmt.Sprintf(runtime.BinaryURL, *tc.Spec.Version)
	initTmpl := template.Must(template.New("ansibleInit").Delims(leftDelim, rightDelim).Parse(ansibleInitTmpl))
	prepareBuf := bytes.NewBuffer([]byte{})
	if err := initTmpl.Execute(prepareBuf, ansibleInit{downloadURL}); err != nil {
		return err
	}
	fmt.Printf("%s\n", prepareBuf)

	deployDir := deployDirPrefix + tc.Metadata.Name

	for _, node := range tc.Spec.PD.Nodes {
		systemdTmpl := template.Must(template.New("systemdSvcTmpl").Delims(leftDelim, rightDelim).Parse(systemdSvcTmpl))
		systemdBuf := bytes.NewBuffer([]byte{})
		systemd := systemdSvc{
			Name:       "pd",
			DeployUser: "tidb",
			DeployDir:  deployDir,
		}
		if err := systemdTmpl.Execute(systemdBuf, systemd); err != nil {
			return err
		}

		scriptTmpl := template.Must(template.New("runPdScriptTmpl").Delims(leftDelim, rightDelim).Parse(runPdScriptTmpl))
		scriptBuf := bytes.NewBuffer([]byte{})
		name := "pd"
		port := 2380
		script := runPdScript{
			Name:       name,
			Port:       port,
			StatusPort: 2379,
			SelfIP:     node.Host,
			DeployDir:  deployDir,
			DataDir:    *node.DataDir,
			Initials:   fmt.Sprintf("%s=%s:%d", name, node.Host, port),
		}
		if err := scriptTmpl.Execute(scriptBuf, script); err != nil {
			return err
		}

		installTmpl := template.Must(template.New("ansibleInstall").Delims(leftDelim, rightDelim).Parse(ansibleInstallTmpl))
		installBuf := bytes.NewBuffer([]byte{})

		install := ansibleInstall{
			Host:      node.Host,
			Component: "pd",
			DataDir:   *node.DataDir,
			DeployDir: deployDir,
			LogDir:    deployDir,
			Script:    scriptBuf.String(),
			Systemd:   systemdBuf.String(),
		}
		if err := installTmpl.Execute(installBuf, install); err != nil {
			return err
		}
		fmt.Printf("%s\n", installBuf)
	}

	return nil
}

func (r ansible) Destroy(tc runtime.TidbCluster) error {
	return nil
}

func (r ansible) isBootstrapped(tc runtime.TidbCluster) bool {
	return false
}
