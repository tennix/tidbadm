# tidbadm

tidbadm helps to deploy and maintain TiDB cluster.

# Usage

1. Init

``` shell
tidbadm init --runtime ansible > tidb-cluster.yaml
```

Create an initial configuration for the specified runtime.

The configuration can then be further edited before apply.

2. Plan

``` shell
tidbadm plan -config tidb-cluster.yaml
```

This step is used to generate an execution plan. Users can use this command to view what will be changed to the cluster.

3. Apply
``` shell
tidbadm apply -config tidb-cluster.yaml
```

If users are satisfied with the plan execution, they can apply the execution plan.

After the apply, the cluster metadata is written into PD. So next time, users can retrieve the cluster spec from PD.

4. Status

``` shell
tidbadm status --name my-cluster
```

# Design

This command line tool is for unifying TiDB cluster management in different platforms.

To avoid adding *another* TiDB command line tool, we should merge pd-ctl/tikv-ctl/tidb-ctl into this tool as subcommands. So we can unify the interface for different components.

*Note:* The environment preparation is not included in this tool. However, we may provide a subcommand for preflight check to improve the deployment success ratio.

The cluster configuration is defined in `runtime.TidbCluster`, it contains intuitive configuration for a TiDB cluster.

A runtime needs to implement the `Runner` interface. Currently, the implementation is builtin the tidbadm. In the future, we may leverage the Go plugin mechanism or external program to communicate with the main `tidbadm` program.

``` go
type Runner interface {
	Init() (*TidbCluster, error)
	Plan(TidbCluster) error
	Apply(TidbCluster) error
	Destroy(TidbCluster) error
}
```

`tidbadm init` will generate an initial configuration for the specified runtime.

This initial configuration can then be edited further.

During the edition, users can run `tidbadm plan` to see what will be changed to the infrastructure.

After users are satisfied with the configuration, they can run `tidbadm apply` to apply the configuration.

When the cluster is successfully created, the tidb cluster configuration and status are written to PD with a resource version. And also adding a record into `$HOME/.tidbadm/config`.

``` yaml
clusters:
- name: my-cluster
  pd-endpoints:
  - 172.16.4.1:2379
  - 172.16.4.2:2379
  - 172.16.4.3:2379
```

So next time, users can use `tidbadm list` to show the managed TiDB clusters, and `tidbadm status --name my-cluster` to show the status of the specific TiDB cluster.

Each time users do with the cluster, the resource version of the `TidbCluster` is updated, when users try to update the cluster with a smaller resource version configuration, the apply is refused.

When the cluster is bootstrapped, the latter maintenance can be done by `tidbadm edit --name my-cluster`.
