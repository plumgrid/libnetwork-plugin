# PLUMgrid libnetwork-plugin

The PLUMgrid Libnetwork Plugin enables seamless way to connect Docker containers to PLUMgrid.
It enables rich networking functions, secure multi-tenancy, multi-host networking,
high availability, distributed scale-out performance, and container-based micro-segmentation
for Docker clouds.


## Pre-Requisites
- PLUMgrid Standalone
- golang1.4 >=
- Install Docker v1.9+


## Using PLUMgrid Libnetwork Plugin

Before launching PLUMgrid libnetwork plugin, ensure that Docker is [installed](https://docs.docker.com/engine/installation/) on your host(s)

Build the binary using the instructions below and launch it on your host. PLUMgrid plugin must be launched
once per host.

Create a network using PLUMgrid driver
```
$ docker network create --subnet=10.0.0.0/24 -d plumgrid net1
```

Create a network on specific domain
```
$ docker network create --subnet=10.0.0.0/24 -d plumgrid -o domain=DomainID net1
```

You can choose to link the network to a specific router in PLUMgrid virtual domain as well,
this is entirely optional and depending on your use case, you may use it.
```
$ docker network create --subnet=10.0.0.0/24 -d plumgrid -o router=router-1 net1
```

Run your containers using the Docker CLI
```
$ docker run -itd --name=cont1 --net=net1 busybox
```
Run several containers, and verify they can ping each other


## Developing PLUMgrid Libnetwork Plugin

Install Docker (if not already)
```
curl -sSL https://get.docker.com/ | sh
```

Start the docker daemon
```
$ docker daemon
```

Create a [go workspace](https://golang.org/doc/code.html#Workspaces)
```
$ mkdir $HOME/work
```

Export GOPATH environment variable
```
$ export GOPATH=$HOME/work
```

Add workspace’s bin subdirectory to your path
```
$ export PATH=$PATH:$GOPATH/bin
```

Clone the PLUMgrid libnetwork plugin repo
```
$ git clone https://github.com/plumgrid/libnetwork-plugin.git $GOPATH/src/github.com/libnetwork-plugin/
```

Move to plugin directory and make binaries
```
$ cd $GOPATH/src/github.com/libnetwork-plugin/
$ make
```

Move config.ini under /opt/pg/libnetwork/

Update this /opt/pg/libnetwork/config.ini to contain the PLUMgrid Virtual IP

Now run the PLUMgrid libnetwork plugin
```
$ sudo mkdir -p /run/docker/plugins
$ ./plugin/plumgrid

