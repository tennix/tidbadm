package ansible

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	deployDirPrefix = "/opt/tidb-clusters"
)

type ansible struct {
	user      string
	deployDir string
}

func New() ansible {
	deployDir := deployDirPrefix // + tc.Metadata.Name
	return ansible{
		user:      "tidb",
		deployDir: deployDir,
	}
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

func (r ansible) genSystemdSvc(name string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	systemdTmpl := template.Must(template.New("systemdSvcTmpl").Delims(leftDelim, rightDelim).Parse(systemdSvcTmpl))
	systemd := systemdSvc{
		Name:       name,
		DeployUser: r.user,
		DeployDir:  r.deployDir,
	}
	if err := systemdTmpl.Execute(buf, systemd); err != nil {
		return nil, err
	}
	return buf, nil
}

func (r ansible) genTikvScript(pdAddrs []string, node runtime.Node) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	scriptTmpl := template.Must(template.New("runTikvScriptTmpl").Delims(leftDelim, rightDelim).Parse(runTikvScriptTmpl))
	script := runTikvScript{
		Port:       20160,
		StatusPort: 20180,
		SelfIP:     node.Host,
		PdAddrs:    strings.Join(pdAddrs, ","),
		DeployDir:  r.deployDir,
		DataDir:    *node.DataDir,
		LogDir:     r.deployDir + "/logs",
	}
	if err := scriptTmpl.Execute(buf, script); err != nil {
		return nil, err
	}
	return buf, nil
}

func (r ansible) genPdScript(name string, node runtime.Node, port int, initials []string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	scriptTmpl := template.Must(template.New("runPdScriptTmpl").Delims(leftDelim, rightDelim).Parse(runPdScriptTmpl))
	script := runPdScript{
		Name:       name,
		Port:       port,
		StatusPort: 2379,
		SelfIP:     node.Host,
		DeployDir:  r.deployDir,
		DataDir:    *node.DataDir,
		LogDir:     r.deployDir + "/logs",
		Initials:   strings.Join(initials, ","),
	}
	if err := scriptTmpl.Execute(buf, script); err != nil {
		return nil, err
	}
	return buf, nil
}

func (r ansible) genInstall(name, version string, script, systemd bytes.Buffer, node runtime.Node, binaries []string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	installTmpl := template.Must(template.New("ansibleInstall").Delims(leftDelim, rightDelim).Parse(ansibleInstallTmpl))
	install := ansibleInstall{
		Host:      node.Host,
		Version:   version,
		Component: name,
		DataDir:   *node.DataDir,
		DeployDir: r.deployDir,
		LogDir:    r.deployDir,
		Binaries:  binaries,
		Script:    script.String(),
		Systemd:   systemd.String(),
	}
	if err := installTmpl.Execute(buf, install); err != nil {
		return buf, err
	}
	return buf, nil
}

func (r ansible) genInit(version string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	initTmpl := template.Must(template.New("ansibleInit").Delims(leftDelim, rightDelim).Parse(ansibleInitTmpl))
	init := ansibleInit{
		Version: version,
	}
	if err := initTmpl.Execute(buf, init); err != nil {
		return nil, err
	}
	return buf, nil
}

func (r ansible) Apply(tc runtime.TidbCluster) error {
	if r.isBootstrapped(tc) {
		return nil
	}

	prepareBuf, err := r.genInit(*tc.Spec.Version)
	if err != nil {
		return err
	}
	prepareFile, err := ioutil.TempFile("", "playbook")
	if err != nil {
		return err
	}
	defer os.Remove(prepareFile.Name())

	fmt.Printf("Creating playbook file %s\n", prepareFile.Name())
	if _, err := prepareFile.Write(prepareBuf.Bytes()); err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	r.run(wg, prepareFile.Name())

	wg.Wait()

	initials := []string{}
	pdAddrs := []string{}
	for _, node := range tc.Spec.PD.Nodes {
		name := strings.ReplaceAll(node.Host, ".", "-")
		initials = append(initials, fmt.Sprintf("%s=%s:%d", name, node.Host, 2380))
		pdAddrs = append(pdAddrs, fmt.Sprintf("%s:2379", node.Host))
	}

	for _, node := range tc.Spec.PD.Nodes {
		name := string(runtime.PDMember)
		port := 2380
		systemdBuf, err := r.genSystemdSvc(name)
		if err != nil {
			return err
		}

		scriptBuf, err := r.genPdScript(strings.ReplaceAll(node.Host, ".", "-"), node, port, initials)
		if err != nil {
			return err
		}

		binaries := []string{"pd-server", "pd-ctl"}
		installBuf, err := r.genInstall(name, *tc.Spec.Version, *scriptBuf, *systemdBuf, node, binaries)
		if err != nil {
			return err
		}

		installFile, err := ioutil.TempFile("", "playbook")
		if err != nil {
			return err
		}
		defer os.Remove(installFile.Name())

		fmt.Printf("Creating playbook file %s\n", installFile.Name())
		if _, err := installFile.Write(installBuf.Bytes()); err != nil {
			return err
		}
		wg.Add(1)
		go r.run(wg, installFile.Name())
	}

	for _, node := range tc.Spec.TiKV.Nodes {
		name := string(runtime.TiKVMember)
		systemdBuf, err := r.genSystemdSvc(name)
		if err != nil {
			return err
		}

		scriptBuf, err := r.genTikvScript(pdAddrs, node)
		if err != nil {
			return err
		}

		binaries := []string{"tikv-server", "tikv-ctl"}
		installBuf, err := r.genInstall(name, *tc.Spec.Version, *scriptBuf, *systemdBuf, node, binaries)
		if err != nil {
			return err
		}

		installFile, err := ioutil.TempFile("", "playbook")
		if err != nil {
			return err
		}
		defer os.Remove(installFile.Name())

		fmt.Printf("Creating playbook file %s\n", installFile.Name())
		if _, err := installFile.Write(installBuf.Bytes()); err != nil {
			return err
		}

		wg.Add(1)
		go r.run(wg, installFile.Name())
	}

	wg.Wait()

	return nil
}

func (r ansible) run(wg sync.WaitGroup, playbook string) error {
	defer wg.Done()
	stderr := &bytes.Buffer{}
	cmd := exec.Command("ansible-playbook", playbook)
	cmd.Stderr = stderr
	err := cmd.Start()
	return err
}

func (r ansible) Destroy(tc runtime.TidbCluster) error {
	return nil
}

func (r ansible) isBootstrapped(tc runtime.TidbCluster) bool {
	return false
}
