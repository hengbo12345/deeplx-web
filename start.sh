#!/bin/bash

set -e

echo > backend/logs/app.log
rm -fr backend/uploads/*

cd frontend
npm run build
cp -r dist/* ~/DockerUse/nginx/html/deeplx-html/

cd ../backend
SERVER_PORT=8449 LOG_LEVEL=debug go run cmd/server/main.go