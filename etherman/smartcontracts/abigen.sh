#!/bin/sh

set -e

gen() {
    local package=$1
    mkdir -p ${package}
    abigen --bin bin/${package}.bin --abi abi/${package}.abi --pkg=${package} --out=${package}/${package}.go
}

gen polygondatacommittee
gen polygonvalidiumetrog
