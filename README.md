# RealtimeGo

## Start

To read code start from file `main.go`

### Before start

1. Copy file `.env.template` to `.env`
2. Change connection information and IP address.
3. Check command options `go run . -h`

### With External Message Broker
```bash
# Start external message broker
docker compose up pulsar kafka redis -d

# run golang servers in your host
go run . -p 6000 -t 0
go run . -p 6001 -t 0
# with docker (have to change option -t=n;when 0<n<3, in docker-compose.yml before run the command)
docker compose up --build realtime
docker compose up realtime2
docker compose up realtime3
```


### With Internal Message Transport

The current version does not implement a connection retry. Please start servers at nearly the same time.

```bash
go run . -p 6000 -b 16000 -t 3
go run . -p 6001 -b 16001 -t 3

# with docker (have to change option -t=3 in docker-compose.yml before run the command)
docker compose up --build realtime
docker compose up realtime2
docker compose up realtime3
```

