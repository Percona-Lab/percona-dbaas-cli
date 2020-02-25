#!/bin/bash

echo Build percona dbaas
cd ../cmd
go build -o ../bin/percona-dbaas 

echo Run tests
cd ../integtests
go run ./ ../bin/percona-dbaas
