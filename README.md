# tidbadm

tidbadm helps to deploy TiDB cluster.

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
