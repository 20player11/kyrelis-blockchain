#!/bin/sh
set -eu
cd "$(dirname "$0")/operator"
exec go run ./cmd/kyrelis-operator
