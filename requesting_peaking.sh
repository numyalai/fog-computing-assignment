#!/bin/bash

while [ true ]; do
    curl -X POST localhost:6001 --data "TEST"
    sleep 3
done