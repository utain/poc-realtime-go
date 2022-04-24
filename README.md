# RealtimeGo

## Start

```bash
# setup env
cp .env.template .env

# start docker
docker compose up pulsar kafka redis -d

# run golang servers in local
go run . -h
go run . -p 6000 -t 0
go run . -p 6001 -t 0
# or run with docker
docker compose up --build realtime
docker compose up realtime2
docker compose up realtime3
```