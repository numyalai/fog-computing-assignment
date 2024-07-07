#!/bin/bash

while [ true ]; do
    curl -X POST 34.65.17.116:6001 --data "TEST"
    sleep 3
done