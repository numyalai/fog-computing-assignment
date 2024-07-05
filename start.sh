#!/bin/bash
trap 'kill $(jobs -p)' SIGINT
./services/router &
./services/client localhost localhost:5001 &
./services/watcher &
wait