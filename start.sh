#!/bin/bash
trap 'kill $(jobs -p)' SIGINT
#./services/router &
./services/client 34.65.17.116:5001 &
./services/cpu_watcher &
./services/ram_watcher &
wait