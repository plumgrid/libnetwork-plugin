# PLUMgrid libnetwork-plugin
PLUMgrid libnetwork plugin is a Docker network plugin that provides networking between docker containers using PLUMgrid.

## Features
- TODO

## Pre-Requisites
- PLUMgrid Director
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
* now run the PLUMgrid libnetwork plugin
```
$ sudo mkdir -p /run/docker/plugins
$ ./plugins/plumgrid
```
* create a network with docker
```
$ docker network create --subnet=10.0.0.0/24 -d plumgrid net1
```
* create a docker container onto that network
```
$ docker run -itd --name=cont1 --net=net1 busybox
```
run several containers, and verify they can ping each other
