# Walkthrough

Let's do a hands-on walkthrough now.

First, we have a look at the available command line parameters: 

```bash
$ kube-sdx -h
Usage of kube-sdx:
  -local string
        the local context you want me to use (default "minikube")
  -namespace string
        the namespace you want me to keep alive (default "default")
  -remote string
        the remote context you want me to use
  -verbose
        if set to true, I'll show you all the nitty gritty details
```

Now let's launch it with some sensible values set:

```bash
$ kube-sdx \
  --namespace=mh9sandbox \
  --remote=mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb


```

Note that messages in yellow are from the connection detection module, running concurrently to the backup and restore module.

Especially when you use `--verbose` you'll also see messages in blue, coming from the low-level shelling out module, effectively showing which `kubectl` commands have been issued.