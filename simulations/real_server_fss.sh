#!/bin/bash
if [ $# -ne 0 ]; then
    echo "not enough arguments provided"
    exit 1
fi

id=$1
scheme=$2
expriment=$3

export GOGC=8000

# remove stats log and create new files
rm results/stats*

# build server
cd ../cmd/grpc/server
go build

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

# run servers
echo "##### running server $id with $scheme scheme #####"
# run server given the correct scheme 
cmd/grpc/server/server -id=$id -files=31 -experiment -scheme=$scheme | tee -a simulations/results/stats_server-${id}_${scheme}_${experiment}.log
wait $!
