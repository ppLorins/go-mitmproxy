#!/bin/bash

bin=go-mitmproxy-linux
pkill ${bin}

pid=`ps -ef | grep ${bin} | grep -v grep | awk '{print $2}'`
kill -9 $pid

nohup ./${bin} -debug=0 -upstream=http://localhost:1081 -redis_conn_string="localhost:6379,password=" > stdout.txt 2>&1 &