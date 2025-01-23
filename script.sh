#!/bin/bash

# Build the Docker image
docker image build -f Dockerfile -t forumimg .

# Run the Docker container
docker container run -p 8080:8080 --detach --name forumcontainer forumimg

# Prune removes unused Docker objs
# ideally, remove unused within last 24h: docker system prune --filter "until=24h"
# -f forces prune without confirmation

# Remove -a all images without at least one container associated
docker image prune -a -f

# Removes stopped containers, unused networks, dangling imgs, unused build cache
docker system prune -f 

echo "Docker image built, container running, and system pruned."