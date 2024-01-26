#!/bin/bash

# This is a simple shortcut used to create a loopback SSH tunnel from localhost:3001
# to localhost:3000 through a remote server to naturally reproduce the network latency
# that would be experienced by people playing the game over the internet.

SSH_HOST=$1

if [ -z "$SSH_HOST" ]; then
  echo "Usage: $0 <ssh-host>"
  exit 1
fi

ssh -N -L localhost:3001:localhost:4000 -R localhost:4000:localhost:3000 "$SSH_HOST"
