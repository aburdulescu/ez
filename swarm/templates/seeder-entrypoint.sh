#!/bin/bash

set -e

cd /go-ez/config
{{range .files}}
/go-ez/bin/ezl add /go-ez/files/{{.name}}
{{end}}

/go-ez/bin/ezs --config /go-ez/config/ezs.json
