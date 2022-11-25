#!/usr/bin/env bash

cd ../data-service && sh data-service.sh start
cd ../api-server && sh api-server.sh start
cd ../auth-server && sh auth-server.sh start
