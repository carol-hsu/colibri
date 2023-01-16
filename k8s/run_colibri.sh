#!/bin/bash

## Copyright 2022 Carol Hsu
## 
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
## 
##     http://www.apache.org/licenses/LICENSE-2.0
## 
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.

APP_YAML=
HOSTNAME=
USER=
CMD_KEYWORD=

kubectl create -f $APP_YAML

# mimic a random time to trigger colibri
sleep 10

pid=$(ssh $USER@$HOSTNAME ps aux | grep $CMD_KEYWORD | grep "root" | awk '{print $2}')

cp colibri.yml $CMD_KEYWORD.yml

sed -i "s/APP_PID/$pid/g" $CMD_KEYWORD.yml

kubectl create -f $CMD_KEYWORD.yml

# wait a while for metrics colleciton, change to any closer time to your querying configurations
sleep 20

colibri_pod=$(kubectl get pod | grep "colibri-job" | awk '{print $1}')
while [ "$(kubectl get pod | grep "colibri-job" | awk '{print $3}')" = "Running" ]; do
    sleep 1
    echo "Waiting"
done

kubectl logs $colibri_pod
