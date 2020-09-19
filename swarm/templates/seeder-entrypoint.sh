#!/bin/bash

set -e

{{range .files}}
addFile.sh {{.name}} {{.size}}
{{end}}

/go-ez/bin/ezs --config /go-ez/config/ezs.json
