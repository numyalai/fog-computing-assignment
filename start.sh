#!/bin/bash
trap 'kill $(jobs -p)' SIGINT
./services/router &
./services/client &
./services/watcher &
wait