#!/bin/bash

binname="go-instrument"
srcpath="github.com/efritz/go-instrument/internal/e2e-tests"
genpath="./internal/e2e-tests/instrumented"

if [ ! -f "./${binname}" ]; then
    function finish {
        echo "Removing binary..."
        rm "./${binname}"
    }

    echo "Binary not found, building..."
    go build
    trap finish EXIT
fi

echo "Clearing old instrumented types..."
rm -f "${genpath}/*.go"

echo "Generating instrumented types..."
"./${binname}" -d "${genpath}" -f "${srcpath}" --metric-prefix '.*Do.*:action' --metric-prefix 'Next:calc'
