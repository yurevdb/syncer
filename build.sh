#!/bin/bash

go build -o out/syncer cmd/syncer/syncer.go
sudo rm /usr/bin/syncer
sudo cp out/syncer /usr/bin/syncer

go build -o out/syncerd cmd/daemon/syncerd.go
sudo rm /usr/bin/syncerd
sudo cp out/syncerd /usr/bin/syncerd
