#!/bin/bash
echo "启动程序中...."
ps -ef|grep ./main|awk '{print $2}'|xargs kill -9
sleep 2s
nohup ./main >/dev/null &
sleep 1s
echo "程序启动完成！"