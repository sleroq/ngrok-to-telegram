#!/bin/bash

set -e

source scripts/env.sh

LDFLAGS=(
  "-X 'nkgrok-to-telegram.BOT_TOKEN=${BOT_TOKEN}'"
  "-X 'nkgrok-to-telegram.USERNAME=${USERNAME}'"
)

go build -ldflags="${LDFLAGS[*]}" -o out/startngrok
