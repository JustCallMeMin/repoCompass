#!/usr/bin/env bash

set -eu

echo "Preparing demo fixtures..."
mkdir -p /tmp/repocompass-demo

echo "Cloning kubernetes/kubernetes..."
if [ ! -d "/tmp/repocompass-demo/kubernetes" ]; then
    git clone --depth 1 https://github.com/kubernetes/kubernetes /tmp/repocompass-demo/kubernetes
fi

echo "Cloning expressjs/express..."
if [ ! -d "/tmp/repocompass-demo/express" ]; then
    git clone --depth 1 https://github.com/expressjs/express /tmp/repocompass-demo/express
fi

echo "Cloning pallets/flask..."
if [ ! -d "/tmp/repocompass-demo/flask" ]; then
    git clone --depth 1 https://github.com/pallets/flask /tmp/repocompass-demo/flask
fi

echo "Demo fixtures prepared at /tmp/repocompass-demo"
