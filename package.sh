#!/bin/sh
set -e

# compile for version
make
if [ $? -ne 0 ]; then
    echo "make error"
    exit 1
fi

# Get version
raw_version=`./bin/mefrps --version`
frp_version=`echo $raw_version | sed 's/MEFrp_//g'`
echo "build version: $frp_version"

# cross_compiles
make -f ./Makefile.cross-compiles

rm -rf ./release/packages
mkdir -p ./release/packages

os_all='linux windows darwin freebsd android'
arch_all='386 amd64 arm arm64 mips64 mips64le mips mipsle riscv64 loong64 s390x'
extra_all='_ hf'

cd ./release

for os in $os_all; do
    for arch in $arch_all; do
        for extra in $extra_all; do
            suffix="${os}_${arch}"
            if [ "x${extra}" != x"_" ]; then
                suffix="${os}_${arch}_${extra}"
            fi
            
            if [ "x${os}" = x"windows" ]; then
                if [ -f "./mefrpc_${suffix}.exe" ]; then
                    mv "./mefrpc_${suffix}.exe" "./packages/mefrpc_${suffix}_${frp_version}.exe"
                fi
                if [ -f "./mefrps_${suffix}.exe" ]; then
                    mv "./mefrps_${suffix}.exe" "./packages/mefrps_${suffix}_${frp_version}.exe"
                fi
            else
                if [ -f "./mefrpc_${suffix}" ]; then
                    mv "./mefrpc_${suffix}" "./packages/mefrpc_${suffix}_${frp_version}"
                fi
                if [ -f "./mefrps_${suffix}" ]; then
                    mv "./mefrps_${suffix}" "./packages/mefrps_${suffix}_${frp_version}"
                fi
            fi
        done
    done
done

cd -
