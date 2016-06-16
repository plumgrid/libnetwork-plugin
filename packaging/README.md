# PLUMgrid Libnetwork Plugin Packaging

Follow the steps below to build a libnetwork plugin debian package:

- Create a go workspace and clone the libnetwork plugin repo (follow main README steps).

- Navigate to the packaging directory and run the deb creation script:

```
$ cd $GOPATH/src/github.com/libnetwork-plugin/packaging
$ chmod +x build_libnetwork_deb.sh
$ ./build_libnetwork_deb.sh
```
This will make the source files and create a new debian package under *$GOPATH/src/github.com/libnetwork-plugin/*

- The debian can now be installed with:

```
$ dpkg -i libnetwork_<version>_all.deb
```

**Note:** this package has a dependency on docker-engine being pre-installed.
