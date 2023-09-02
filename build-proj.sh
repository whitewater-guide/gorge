#! /bin/bash

if [[ $(ldconfig -p) =~ "libproj.so" ]]; then
  echo "libproj found"
  exit 0
fi

mkdir -p proj
curl https://download.osgeo.org/proj/proj-8.2.1.tar.gz | tar -xz --strip-components=1 --directory proj
mkdir -p proj/build
cd proj/build

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
