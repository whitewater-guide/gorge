#! /bin/bash

set -ex
LIBPROJ_DIR=${LIBPROJ_DIR:-libproj}
LIBPROJ_VERSION=8.2.1

if [ ! -d "${LIBPROJ_DIR}/src" ]; then 
  echo "downloading libproj sources"
  mkdir -p ${LIBPROJ_DIR}
  curl https://download.osgeo.org/proj/proj-${LIBPROJ_VERSION}.tar.gz | tar -xz --strip-components=1 --directory ${LIBPROJ_DIR}
fi

# unlike cmake, this wull build both shared (for tests and dev) and static (for release) libraries
cd ${LIBPROJ_DIR}
./configure --disable-tiff --without-mutex --without-curl
make
make install
