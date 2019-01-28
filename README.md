# Docker Image Cleaner

This container is for standalone docker or docker swarm clusters for cleaning images as a maintenance utility that relays inside as a container.
It works with time interval that defaults to 1h and gives a filter list to exclude specific images from being deleted.

To run in a swarm cluster:

```bash
sudo docker stack deploy -c deploy/swarm/docker-image-cleaner.yml cleaner
```

To run in a standalone docker host:

```bash
sudo docker-compose -f deploy/standalone/docker-image-cleaner.yml up -d
```

To run it ad-hoc:

```bash
sudo docker run \
  -e TIME_INTERVAL=5s \
  -e FILTER=alpine:latest \
  -v /var/run/docker.sock:/var/run/docker.sock -d omerha/docker-image-cleaner:latest
```

Configuration environment variables:

- TIME_INTERVAL='120m' (can be s,m,h)
- FILTER='alpine:latest,your-image:tag' (Filter your images you wish to keep on your docker hosts)

Optional environment variables: DOCKER_API_VERSION that now defaults to 1.39