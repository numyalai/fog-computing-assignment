#!/bin/bash
trap 'kill $(jobs -p)' SIGINT
./router &
./client &
wait