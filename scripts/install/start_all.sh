#!/usr/bin/env bash

cd ../data-service && sh data-service.sh start
cd ../api-server && sh api-server.sh start
cd ../auth-server && sh auth-server.sh start
cd ../cloud-server && sh cloud-server.sh start
cd ../account-server && sh account-server.sh start
