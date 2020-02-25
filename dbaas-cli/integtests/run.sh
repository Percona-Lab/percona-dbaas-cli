#!/bin/bash

echo Build percona dbaas
cd ../cmd
go build -o percona-dbaas

echo Build integtests
cd ../integtests
go build

echo Run tests
./integtests