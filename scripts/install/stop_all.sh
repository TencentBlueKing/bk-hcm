#!/usr/bin/env bash

cd ../data-service && sh data-service.sh stop
cd ../api-server && sh api-server.sh stop
cd ../auth-server && sh auth-server.sh stop
cd ../cloud-server && sh cloud-server.sh stop
cd ../account-server && sh account-server.sh stop
