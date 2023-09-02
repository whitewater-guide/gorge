#! /bin/bash

set -ex

LIBPROJ_DIR=libproj
LIBPROJ_VERSION=8.2.1

if [[ $(ldconfig -p) =~ "libproj.so" ]]; then
  echo "libproj found"
  exit 0
fi

if [ ! -d "${LIBPROJ_DIR}/src" ]; then 
  echo "downloading libproj sources"
  mkdir -p ${LIBPROJ_DIR}
  curl https://download.osgeo.org/proj/proj-${LIBPROJ_VERSION}.tar.gz | tar -xz --strip-components=1 --directory ${LIBPROJ_DIR}
fi

mkdir -p ${LIBPROJ_DIR}/build
cd ${LIBPROJ_DIR}/build

# Having shared libraries makes life easire during development
BUILD_SHARED_LIBS=ON
if [ -n "${CI}" ]; then
  echo "disabled building shared libs"
  BUILD_SHARED_LIBS=OFF
fi

cmake -DBUILD_APPS=OFF \
    -DBUILD_SHARED_LIBS=${BUILD_SHARED_LIBS} \
    -DBUILD_TESTING=OFF \
    -DCMAKE_BUILD_TYPE=Release \
    -DENABLE_CURL=OFF \
    -DENABLE_TIFF=OFF \
    ..
cmake --build .
cmake --build . --target install
