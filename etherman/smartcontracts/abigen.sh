#!/bin/sh

set -e

gen() {
    local package=$1
    local forkName=$2

    abigen --bin ${forkName}/bin/${package}.bin --abi ${forkName}/abi/${package}.abi --pkg=${package} --out=${forkName}/${package}/${package}.go
}

gen polygonvalidium etrog
gen polygondatacommittee etrog

gen polygonvalidium elderberry
gen polygondatacommittee elderberry
