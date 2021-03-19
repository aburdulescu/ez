#!/bin/bash

set -e

ezs -dbpath=/tmp/seeder_db -seedaddr=$(hostname -I) -trackeraddr=tracker

# sleep 5

# {{range .files}}
# ezl add /ez/files/{{.name}}
# {{end}}

# sleep infinity
