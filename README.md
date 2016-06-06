# PLUMgrid libnetwork-plugin
The PLUMgrid Libnetwork Plugin enables seamless way to connect Docker containers to PLUMgrid. It provides
distributed and scalable fast networking, with built in HA capabilities and access to feature rich
networking platform as well as advanced networking capabilities like tunneling, multi-host networking,
DNS, security etc.

## Pre-Requisites
- PLUMgrid ONS
- golang1.4 >=
- Install Docker v1.9+
```
curl -sSL https://get.docker.com/ | sh
```
## QuickStart Instructions
The quick start instructions will guide in creating a docker network with PLUMgrid libnetwork plugin and then launch docker containers onto that network.

* Start the docker daemon
```
$ docker daemon
```
* create a [go workspace](https://golang.org/doc/code.html#Workspaces)
```
$ mkdir $HOME/work
```
* export GOPATH environment variable
```
$ export GOPATH=$HOME/work
```
add workspace’s bin subdirectory to your path
```
$ export PATH=$PATH:$GOPATH/bin
```
* clone the PLUMgrid libnetwork plugin repo
```
$ git clone https://github.com/plumgrid/libnetwork-plugin.git $GOPATH/src/github.com/libnetwork-plugin/
```
* move to plugin directory and make binaries
```
$ cd $GOPATH/src/github.com/libnetwork-plugin/
$ make
```
* move config.ini under /opt/pg/libnetwork/

* update this /opt/pg/libnetwork/config.ini to contain the PLUMgrid Virtual IP

* now run the PLUMgrid libnetwork plugin
```
$ sudo mkdir -p /run/docker/plugins
$ ./plugin/plumgrid
```
* the plugin run above will take up the current terminal session, open a new one for the commands given below

* create a network using PLUMgrid driver
```
$ docker network create --subnet=10.0.0.0/24 -d plumgrid net1
```
* create a network on specific domain
```
$ docker network create --subnet=10.0.0.0/24 -d plumgrid -o domain=domainID net1
```
* create a docker container onto that network
```
$ docker run -itd --name=cont1 --net=net1 busybox
```
run several containers, and verify they can ping each other
