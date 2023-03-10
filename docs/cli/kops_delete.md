
<!--- This file is automatically generated by make gen-cli-docs; changes should be made in the go CLI command code (under cmd/kops) -->

## kops delete

Delete clusters, instancegroups, instances, and secrets.

```
kops delete {-f FILENAME}... [flags]
```

### Examples

```
  # Delete a cluster using a manifest file
  kops delete -f my-cluster.yaml
  
  # Delete a cluster using a pasted manifest file from stdin.
  pbpaste | kops delete -f -
```

### Options

```
  -f, --filename strings   Filename to use to delete the resource
  -h, --help               help for delete
  -y, --yes                Specify --yes to immediately delete the resource
```

### Options inherited from parent commands

```
      --config string   yaml config file (default is $HOME/.kops.yaml)
      --name string     Name of cluster. Overrides KOPS_CLUSTER_NAME environment variable
      --state string    Location of state storage (kops 'config' file). Overrides KOPS_STATE_STORE environment variable
  -v, --v Level         number for the log level verbosity
```

### SEE ALSO

* [kops](kops.md)	 - kOps is Kubernetes Operations.
* [kops delete cluster](kops_delete_cluster.md)	 - Delete a cluster.
* [kops delete instance](kops_delete_instance.md)	 - Delete an instance.
* [kops delete instancegroup](kops_delete_instancegroup.md)	 - Delete instance group.
* [kops delete secret](kops_delete_secret.md)	 - Delete one or more secrets.
* [kops delete sshpublickey](kops_delete_sshpublickey.md)	 - Delete an SSH public key.

