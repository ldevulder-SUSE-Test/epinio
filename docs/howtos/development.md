# Development Guidelines

## Local development environment

### Get a cluster

There are many options on how to get a local cluster for development. Here are a few:

- [k3d](https://k3d.io/)
- [k3s](https://github.com/k3s-io/k3s)
- [kind](https://github.com/kubernetes-sigs/kind)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)

Assuming you have `k3d` installed, you can create a cluster with this command:

```
k3d cluster create epinio
```

This command should automatically update your default kubeconfig to point to
the new cluster but if you need to save your kubeconfig manually you can do it with:

```
k3d kubeconfig get epinio > epinio_kubeconfig
```

### Build Epinio

You can build Epinio with the following make target:

```
make build
```

This is building Epinio for linux on amd64 architecture. If you are on a
different OS or architecture you can use one of other the available `build-*` targets.
Look at the Makefile at the root of the project to see what is available.

### Installing Epinio

While we have a [dedicated document](https://docs.epinio.io/installation/installation.html) for cluster
specific instructions, there are some differences for dev environments.
These differences are explained in the section [Behind the curtains](#curtain) at the end of this document.

Since we use `k3d` in our CI tests we have created the make target `prepare_environment_k3d` to prepare
such an environment. That script uses the value of the "EPINIO_SYSTEM_DOMAIN" environment variable
for the `system-domain` installation argument. If the variable is not set, it will try to use a "magic" domain
in the form of "1.2.3.4.omg.howdoi.website" where `1.2.3.4` is the IP address of your k3d cluster and
`omg.howdoi.website` is a mirror-dns service which resolves to the IP address in front of it (similar to nip.io, xip.io etc).

For all other environments the following commands should be sufficient:

```
EPINIO_DONT_WAIT_FOR_DEPLOYMENT=1 ./dist/epinio-linux-amd64 install --system-domain=your-system-domain-here.org
make patch-epinio-deployment
```

You have to use a `system-domain` that points to the Traefik Service on your cluster.
Depending on the type of cluster you use, this IP address may be the one of your
cluster's container (e.g. in k3d, kind etc) or one provided by a load balancer
(e.g. in public cloud providers). In any case, if you don't know what that IP address
is before the Epinio installation, it will be printed for you at the end of the
installation. You can then use it to set up your DNS after the fact.

After making changes to the binary simply invoking `make patch-epinio-deployment` again
will upload the changes into the running cluster.

Another thing `epinio install` does after deploying all components is
the creation and targeting of a standard namespace, `workspace`.

Another post-deployment action performed by `install` is the automatic `config update`
saving API credentials and certs into the client configuration file. As that
command talks directly to the cluster and not the epinio API the
failing server component does not matter.

If the cluster is not running on linux-amd64 it will be necessary to set
`EPINIO_BINARY_PATH` to the correct binary to place into the epinio server
([See here](https://github.com/epinio/epinio/blob/a4b679af88d58177cecf4a5717c8c96f382058ed/scripts/patch-epinio-deployment.sh#L19)).

If the client operation is performed outside of a git checkout it will be
necessary to set `EPINIO_BINARY_TAG` to the correct tag
([See here](https://github.com/epinio/epinio/blob/a4b679af88d58177cecf4a5717c8c96f382058ed/scripts/patch-epinio-deployment.sh#L20)).

The make target `tag` can be used in the checkout the binary came from to
determine this value.

Also, the default `make build` target builds a dynamically linked
binary. This can cause issues if for example the glibc library in the
base image doesn't match the one on the build system. To get past that
issue it is necessary to build a statically linked binary with a
command like:

```
GOARCH="amd64" GOOS="linux" CGO_ENABLED=0 go build -o dist/epinio-linux-amd64
```

#### Mixed Windows/Linux Scenario

A concrete example of the above would be the installation of Epinio from a
Windows host without a checkout, to a Linux-based cluster.

In that scenario the Windows host has to have both windows-amd64 and linux-amd64
binaries. The first to perform the installation, the second for
`EPINIO_BINARY_PATH` to be put into the server.

Furthermore, as the Windows host is without a checkout, the tag has to be
determined in the actual checkout and set into `EPINIO_BINARY_PATH`.

Lastly, do not forget to set up a proper domain so that the client can talk to
the server after installation is done. While during installation only a suitable
`KUBECONFIG` is required after the client will go and use the information from
the ingress, and that then has to properly resolve in the DNS.

<a id='curtain'>
#### Behind the curtains

Setting up a dev cluster takes quite a bit more than the plain

```
epinio install
```

found in the quick install intructions.

Let's look at what is actually done:

When building Epinio, the generated binary assumes that there is a
container image for the Epinio server components, with a tag that
matches the commit you built from.  For example, when calling `make
build` on commit `7bfb700`, the version reported by Epinio is
something like `v0.0.5-75-g7bfb700` and an image `epinio/server:v0.0.5-75-g7bfb700`
is expected to be found.

This works fine for released versions, because the release pipeline ensures
that such an image is built and published.

However when building locally building and publishing an image for
every little change is ... inconvenient.

As described above we set
```
export EPINIO_DONT_WAIT_FOR_DEPLOYMENT=1
```

before calling the `epinio` binary that was created during `make build`.

This tells the `epinio install` command to not wait for the Epinio server
deployment. Since that will be failing without the image. Inspecting
the cluster with

```
kubectl get pod -n epinio --selector=app.kubernetes.io/name=epinio-server
```

will confirm this.

Running `make patch-epinio-deployment` compensates for this issue.
This make target patches the failing Epinio server deployment to use an
existing image from some release and then copies the locally built
`dist/epinio-linux-amd64` binary into it, ensuring that it runs the
same binary as the client.

__Note__ When building for another OS or architecture the
`dist/epinio-linux-amd64` binary will not exist. In this case the path
has to be specified by the environment variable `EPINIO_BINARY_PATH`
as described above in this document.
