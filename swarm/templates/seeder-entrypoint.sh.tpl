#!/bin/bash

set -e

ezs -dbpath={{.seederDbPath}} -seedaddr={{.seedAddr}} -trackeraddr={{.trackerAddr}} &

sleep 5

{{range .files}}
ezl add /ez/files/{{.name}}
{{end}}

sleep infinity
