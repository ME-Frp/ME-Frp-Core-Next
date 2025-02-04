#!/bin/sh
set -e

# compile for version
make
if [ $? -ne 0 ]; then
    echo "make error"
    exit 1
fi

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
                if [ -f "./mefrpc_${os}_${arch}.exe" ]; then
                    mv "./mefrpc_${os}_${arch}.exe" "./packages/mefrpc_${os}_${arch}.exe"
                fi
                if [ -f "./mefrps_${os}_${arch}.exe" ]; then
                    mv "./mefrps_${os}_${arch}.exe" "./packages/mefrps_${os}_${arch}.exe"
                fi
            else
                if [ -f "./mefrpc_${suffix}" ]; then
                    mv "./mefrpc_${suffix}" "./packages/mefrpc_${suffix}"
                fi
                if [ -f "./mefrps_${suffix}" ]; then
                    mv "./mefrps_${suffix}" "./packages/mefrps_${suffix}"
                fi
            fi
        done
    done
done

# Get version and rename files with version at the end
cd packages
raw_version=`./bin/mefrpc --version`
frp_version=`echo $raw_version | sed 's/MEFrp_//g'`
echo "build version: $frp_version"

for file in *; do
    if [ -f "$file" ]; then
        if [[ "$file" == *.exe ]]; then
            base_name="${file%.exe}"
            mv "$file" "${base_name}_${frp_version}.exe"
        else
            mv "$file" "${file}_${frp_version}"
        fi
    fi
done

cd -
