# Seamless Developer Experience

This is a prototype of a command line tool called `kube-sdx` (seamless DX) which enables you to automatically switch between different Kubernetes clusters and continue your work uninterrupted. 

![screen shot of kube-sdx](img/kube-sdx-screen-shot.png)

- [Prerequisits](#prerequisits)
- [Install](#install)
- [Use](#use)
- [Platform-specific notes](#platform-specific-notes)
- [How it works](#how-it-works)

## Prerequisits

We assume you have `kubectl` (or OpenShift's `oc`) command line tool installed and configured as well as that you have a local cluster (Minikube, Minishift, Docker for Desktop) set up and at least one remote cluster configured. The tests have been carried out with the following configuration:

```bash
$ kubectl version --short
Client Version: v1.11.3
Server Version: v1.10.0

$ minikube version
minikube version: v0.28.2
```

## Install

We support Linux, macOS, and Windows and you can download the binaries here:

- Linux: [kube-sdx-linux]()
- macOS: [kube-sdx-macos]()
- Windows: [kube-sdx-windows]()

Download the binary, rename it to `kube-sdx` and put it on your path, and finally make it executable (in *nix: `chmod +x kube-sdx`).

## Use

### Basics
Once downloaded and set up, you can launch `kube-sdx` like so:

```bash
$ kube-sdx --remote=$WORK_CONTEXT
```

Note that `--remote` is the only parameter you must supply, otherwise `kube-sdx` doesn't know what to track (and snapshot) and hence can't function properly. But what happens if you leave it out? Simply this:

```bash
$ kube-sdx
I'm sorry Dave, I'm afraid I can't do that.
I need to know which remote context you want, pick one from below and provide it via the --remote parameter:

CURRENT   NAME                                                      CLUSTER                               AUTHINFO                                       NAMESPACE
          default/192-168-99-100:8443/developer                     192-168-99-100:8443                   developer/192-168-99-100:8443                  default
          default/192-168-99-100:8443/system:admin                  192-168-99-100:8443                   system:admin/192-168-99-100:8443               default
          docker-for-desktop                                        docker-for-desktop-cluster            docker-for-desktop
          dok/api-pro-us-east-1-openshift-com:443/mhausenb          api-pro-us-east-1-openshift-com:443   mhausenb/api-pro-us-east-1-openshift-com:443   dok
          mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb   api-pro-us-east-1-openshift-com:443   mhausenb/api-pro-us-east-1-openshift-com:443   mh9sandbox
*         minikube                                                  minikube                              minikube

```

So no worries, `kube-sdx` will gently remind you to set `--remote` in any case ;)

NOTE: If you want to see `kube-sdx` in action, why don't you check out the [walkthrough](walkthrough.md) of some typical sessions?

### Policies

With `--policy` you set the initial context and what kind of resources to capture and consequently restore, after a switch from either `ONLINE` to `OFFLINE` or the other way round. The logic is as follows:

POLICY     | ONLINE | OFFLINE
---        | ---    | ---
`local:*`  | use `local` context, but no capture | use `local` context and capture it
`remote:*` | use `remote` context and capture it | doesn't make sense (NOP)

For example, if you'd use `--policy=local:deployments` and you're currently offline (that is, no connection to remote cluster) then `kube-sdx` would capture local resources of type `deployments` and once you're online, it would mirror these to the remote.

## Platform-specific notes

Under Windows, you *must* specify the `SDX_KUBECTL_BIN` environment variable, since auto-discovery doesn't work there. Also, if you are using an OpenShift remote cluster, you typically want to set `SDX_KUBECTL_BIN` to `$(which oc)`.

## How it works

In a nutshell, `kube-sdx` keeps an eye on the API server of the configured remote cluster (via a simple HTTP `GET`) and if the connection detection fails, assumes you're offline. Once in offline mode, `kube-sdx` switches over to the local cluster and you can continue your work there. In the background, `kube-sdx` takes regular snapshots of certain key resources such as deployments or services and stores them in YAML docs that get applied to the respective environment, when a switch occurs.

**Local development.** If you want to play around with `kube-sdx` or extend it, here's what's needed:

```bash
$ go version
go version go1.10 darwin/amd64

$ go build -o kube-sdx && \
             ./kube-sdx \
             --namespace=mh9sandbox \
             --remote=mh9sandbox/api-pro-us-east-1-openshift-com:443/mhausenb
```

Above command will build the latest version, creating a binary called `kube-sdx` and execute it with the no `local` env defined (hence going with the default `minikube`), keeping the namespace `mh9sandbox` alive (default: `default`) and uses as the `remote` a context that specifies a project in OpenShift Online.