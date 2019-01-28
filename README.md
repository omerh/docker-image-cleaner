# Docker Image Cleaner

This container is for standalone docker or docker swarm clusters for cleaning images as a maintancen utility that relays inside as a container.

To run in a swarm cluster:

```bash
sudo docker stack deploy -c deploy/swarm/docker-image-cleaner.yml cleaner
```

To run in a standalone docker host:

```bash
sudo docker-compose -f deploy/standalone/docker-image-cleaner.yml up -d
```

Configuration environment variables:

- TIME_INTERVAL='120m' (can be s,m,h)
- FILTER='alpine:latest,your-image:tag' (Filter your images you wish to keep on your docker hosts)