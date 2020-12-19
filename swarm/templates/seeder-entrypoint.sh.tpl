#!/bin/bash

set -e

ezs -dbpath={{.seederDbPath}} -seedaddr={{.seedAddr}} -trackeraddr={{.trackerAddr}} &

{{range .files}}
ezl add /ez/files/{{.name}}
{{end}}
