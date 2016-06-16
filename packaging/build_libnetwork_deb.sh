#!/bin/bash
#
# build_libnetwork_deb
#

WDIR=$(readlink -m `pwd`/..)
PKG_PREFIX=libnetwork
PKG_DEBVERSION="1.0-1"
PKG_DESCRIPTION="PLUMgrid Docker libnetwork plugin"
PKG_ARCH="all"
PKG_DEPS="docker-engine (>=1.11)"

BDIR_BASE=${WDIR}/${PKG_PREFIX}_${PKG_DEBVERSION}_${PKG_ARCH}


function gen_control_file(){

rm -rf ${BDIR_BASE}/DEBIAN/control
mkdir -p ${BDIR_BASE}/DEBIAN
cat > ${BDIR_BASE}/DEBIAN/control <<DELIM___
Package: $1
Version: $2
Section: custom
Priority: optional
Architecture: $3
Depends: $5
Maintainer: Javeria Khan <javeriak@plumgrid.com>
Description: $4
DELIM___
}


function build_package() {
  if [ $# -gt 0 ] ; then
    script_loc=$1
  else
    script_loc=${WDIR}/packaging/debian-control
  fi
  if [ -e ${script_loc}/${PKG_PREFIX}-postinst.sh ]; then
      cp -pf ${script_loc}/${PKG_PREFIX}-postinst.sh ${BDIR_BASE}/DEBIAN/postinst
      chmod 0755  ${BDIR_BASE}/DEBIAN/postinst
      echo "Found post-inst file ${script_loc}/${PKG_PREFIX}-postinst.sh"
  fi
  if [ -e ${script_loc}/${PKG_PREFIX}-prerm.sh ]; then
      cp -pf ${script_loc}/${PKG_PREFIX}-prerm.sh ${BDIR_BASE}/DEBIAN/prerm
      chmod 0755  ${BDIR_BASE}/DEBIAN/prerm
      echo "Found pre-rm file ${script_loc}/${PKG_PREFIX}-prerm.sh"
  fi
   
  fakeroot dpkg-deb --build ${BDIR_BASE}
  status=$?

  # exit immediately upon error, do not proceed further
  if [[ $status -ne 0 ]]; then
    return $status
  fi

  return $status
}

# Delete previously created files/dirs
rm -f ${WDIR}/${PKG_PREFIX}_*.deb || true
rm -rf ${BDIR_BASE} || true

# Create necessary target directories
mkdir -p "${BDIR_BASE}/DEBIAN"
mkdir -p "${BDIR_BASE}/opt/pg/${PKG_PREFIX}"
mkdir -p "${BDIR_BASE}/etc/init.d"
mkdir -p "${BDIR_BASE}/run/docker/plugins"

# Generate libnetwork binary
pushd "${WDIR}" > /dev/null
make
popd > /dev/null

cp ${WDIR}/config.ini ${BDIR_BASE}/opt/pg/${PKG_PREFIX}/
cp ${WDIR}/plugin/plumgrid ${BDIR_BASE}/opt/pg/${PKG_PREFIX}/libnetwork
cp init ${BDIR_BASE}/etc/init.d/libnetwork

chmod 0755 ${BDIR_BASE}/etc/init.d/libnetwork

gen_control_file  "${PKG_PREFIX}"  "${PKG_DEBVERSION}"  "${PKG_ARCH}" "${PKG_DESCRIPTION}" "${PKG_DEPS}" ""

build_package
status=$?

rm -rf ${BDIR_BASE}
echo "Package build exiting with status: ${status}"
exit $status
