#!/bin/bash

APP_YAML=
HOSTNAME=
USER=
CMD_KEYWORD=

kubectl create -f $APP_YAML

sleep 10

pid=$(ssh $USER@$HOSTNAME ps aux | grep $CMD_KEYWORD | grep "root" | awk '{print $2}')

cp grabber.yml $CMD_KEYWORD.yml

sed -i "s/1234/$pid/g" $CMD_KEYWORD.yml

kubectl create -f $CMD_KEYWORD.yml



