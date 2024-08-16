#! /bin/bash

rm -f main.zip bootstrap > /dev/null 2>&1 
CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o bootstrap
zip main.zip bootstrap
rm bootstrap
