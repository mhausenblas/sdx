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

Switching over to context mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb
--- STARTING SDX

I'm using the following configuration:
- local context: minikube
- remote context: mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb
- namespace to keep alive: mh9sandbox
---

Connection detection [ONLINE], probe https://api.pro-us-east-1.openshift.com:443 resulted in 200 OK
Connection detection [OFFLINE], probe resulted in Get https://api.pro-us-east-1.openshift.com:443: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
...
```
