#!/bin/bash

set -e

source scripts/env.sh

LDFLAGS=(
  "-X 'main.BOT_TOKEN=${BOT_TOKEN}'"
  "-X 'main.USERNAME=${USERNAME}'"
)

go build -ldflags="${LDFLAGS[*]}" -o out/startngrok
