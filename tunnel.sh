#!/bin/bash

# This is a simple shortcut used to create a loopback SSH tunnel from localhost:3001
# to localhost:3000 through a remote server to naturally reproduce the network latency
# that would be experienced by people playing the game over the internet.

SSH_HOST=$1

if [ -z "$SSH_HOST" ]; then
  echo "Usage: $0 <ssh-host>"
  exit 1
fi

TARGET_PORT=3000 # port that the host will listen on
SOURCE_PORT=3001 # port that the client will connect to
echo "tunneling localhost:$SOURCE_PORT -> localhost:$TARGET_PORT through $SSH_HOST"

ssh -N \
  -L localhost:$SOURCE_PORT:localhost:4000 \
  -R localhost:4000:localhost:$TARGET_PORT \
  "$SSH_HOST"
