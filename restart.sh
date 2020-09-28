#!/usr/bin/env bash

cat ./pid.txt | xargs kill -10
rm -f ./log/*.log
nohup ./tcpproxy &
