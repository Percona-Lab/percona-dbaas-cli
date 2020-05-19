#!/bin/bash

echo Build percona dbaas
cd ./dbaas-cli/cmd
go build -o ../../percona-dbaas 

echo Run tests
cd ../integtests
go build -o ../../integtests
