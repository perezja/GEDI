#!/bin/bash
set -e

service postgresql start
ssh-keygen -t rsa -N '' -f ~/.ssh/id_rsa

# Then exec the container's main process (what's set as CMD in the Dockerfile).
exec "$@" 
