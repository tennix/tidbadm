package ansible

type ansibleInit struct {
	Version string
}

// The prepare task is done on localhost before any other tasks
var ansibleInitTmpl = `- name: Prepare
  hosts: localhost
  gather_facts: no
  tasks:
  - name: Download binary
    get_url:
      url: https://download.pingcap.org/tidb-<% .Version %>-linux-amd64.tar.gz
      dest: "{{ playbook_dir }}/tidb-<% .Version %>-linux-amd64.tar.gz"
  - name: Extract tarball
    unarchive:
      src: tidb-<% .Version %>-linux-amd64.tar.gz
      dest: "{{ playbook_dir }}/"
`

type ansibleInstall struct {
	Host      string
	Name      string
	Version   string
	Component string
	DataDir   string
	DeployDir string
	LogDir    string
	Binaries  []string
	Script    string
	Systemd   string
}

// The install tasks can be executed simultaneously
var ansibleInstallTmpl = `- name: Install
  hosts: <% .Host %>
  gather_facts: no
  tasks:
  - name: Ensure directory exist
    file: path={{ item }} state=directory mode=0755 recurse=yes
    with_items:
    - <% .DataDir %>
    - <% .DeployDir %>
    - <% .DeployDir %>/bin
    - <% .DeployDir %>/conf
    - <% .LogDir %>
  - name: Install binary
    copy:
      src: tidb-<% .Version %>-linux-amd64/bin/{{ item }}
      dest: <% .DeployDir %>/bin/{{ item }}
      mode: 0755
    with_items:
<%- range .Binaries %>
    - <% . %>
<%- end %>
  - name: Install script
    copy:
      content: |
        <% .Script %>
      dest: <% .DeployDir %>/run_<% .Component %>.sh
      mode: 0755
  - name: Install service
    copy:
      content: |
        <% .Systemd %>
      dest: /etc/systemd/system/<% .Component %>.service
`

// The start tasks are executed after the install task finished
// and in the order PD -> TiKV -> TiDB
var ansibleStartTmpl = `- name: Install
  hosts: <% .Host %>
  gather_facts: no
  tasks:
  - name: Ensure service started
    systemd:
      name: <% .Name %>
      state: started
      enabled: yes
`

type systemdSvc struct {
	Name       string
	DeployUser string
	DeployDir  string
}

var systemdSvcTmpl = `[Unit]
        Description=<% .Name %> service
        After=syslog.target network.target remote-fs.target nss-lookup.target

        [Service]
        LimitNOFILE=1000000
        #LimitCORE=infinity
        LimitSTACK=10485760
        User=<% .DeployUser %>
        ExecStart=<% .DeployDir %>/scripts/run_<% .Name %>.sh
        Restart=always
        RestartSec=15s

        [Install]
        WantedBy=multi-user.target
`

type runPdScript struct {
	Name       string
	Port       int
	StatusPort int
	SelfIP     string
	DeployDir  string
	DataDir    string
	LogDir     string
	Initials   string
	Joins      string
}

var runPdScriptTmpl = `#!/usr/bin/env bash
        set -e
        ulimit -n 1000000

        # WARNING: This file was auto-generated. Do not edit!
        #          All your edit might be overwritten!

        DEPLOY_DIR=<% .DeployDir %>

        cd "${DEPLOY_DIR}" || exit 1

        exec bin/pd-server \
            --name=<% .Name %> \
            --client-urls=http://0.0.0.0:<% .StatusPort %> \
            --advertise-client-urls=http://<% .SelfIP %>:<% .StatusPort %> \
            --peer-urls=http://0.0.0.0:<% .Port %> \
            --advertise-peer-urls=http://<% .SelfIP %>:<% .Port %> \
            --data-dir=<% .DataDir %> \
        <%- if .Initials %>
            --initial-cluster=<% .Initials %> \
        <%- else %>
            --join=<% .Joins %> \
        <%- end %>
            --config=confg/pd.toml \
            --log-file=<% .LogDir %>/pd.log 2>> <% .LogDir %>/pd_stderr.log
`

type runTikvScript struct {
	Port       int
	StatusPort int
	SelfIP     string
	PdAddrs    string
	DeployDir  string
	DataDir    string
	LogDir     string
}

var runTikvScriptTmpl = `#!/usr/bin/env bash
        set -e
        ulimit -n 1000000

        # WARNING: This file was auto-generated. Do not edit!
        #          All your edit might be overwritten!

        DEPLOY_DIR=<% .DeployDir %>

        cd "${DEPLOY_DIR}" || exit 1

        export RUST_BACKTRACE=1

        export TZ=${TZ:-/etc/localtime}

        exec bin/tikv-server \
            --addr=0.0.0.0:<% .Port %> \
            --advertise-addr=<% .SelfIP %>:<% .Port %> \
            --status-addr=0.0.0.0:<% .StatusPort %> \
            --pd-endpoints=<% .PdAddrs %> \
            --data-dir=<% .DataDir %> \
            --config=conf/tikv.toml \
            --log-file=<% .LogDir %>/tikv.log 2>> <% .LogDir %>/tikv_stderr.log
`

var runTidbScriptTmpl = `#!/usr/bin/env bash
        set -e
        ulimit -n 1000000

        # WARNING: This file was auto-generated. Do not edit!
        #          All your edit might be overwritten!

        DEPLOY_DIR=<% .DeployDir %>

        cd "${DEPLOY_DIR}" || exit 1

        export TZ=<% .TZ %>

        exec bin/tidb-server \
            -P <% .Port %> \
            --status=<% .StatusPort %> \
            --advertise-address=<% .SelfIP %> \
            --path=<% .PdAddrs %> \
            --config=conf/tidb.toml \
            --enable-binlog=<% .Binlog %> \
            --log-slow-query=<% .LogDir %>/tidb-slow.log \
            --log-file=<% .LogDir %>/tidb.log 2>> <% .LogDir %>/tidb_stderr.log
`
