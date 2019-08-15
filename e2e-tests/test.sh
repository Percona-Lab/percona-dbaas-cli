#!/bin/bash

function destroy {
    local prefix=$(echo $1 | cut -d- -f1)
    if [[ ${prefix} == "pxc" ]]; then
        echo pxxxx$prefix
    fi
    if [[ ${prefix} == "psmdb" ]]; then
        echo pssss$prefix
    fi
    


}

destroy pxc-54555
destroy psmdb-5555