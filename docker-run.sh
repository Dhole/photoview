#!/bin/sh
set -ex

docker run -it --mount type=bind,source=${PWD}/api,target=/app --network host --entrypoint /bin/bash photoview
