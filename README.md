# Minimal MongoDB Golang REST microservice

This is a minimal key/value microservice allowing to store and retrieve a key/value pairs in MongoDB

## Run the demo
```sh
# clone the repository
git clone https://github.com/telemac/mongo-minimal-microservice.git
cd mongo-minimal-microservice

# Launch a MongoDB demo database in a docker container
docker-compose up -d
```

## Compile the binary for your native os with a go environment installed
```sh
go build -o mongo-minimal-microservice cmd/main.go
```

## Compile the binary for linux without a go environment installed
```sh
docker run --rm --name golang -t -v "$PWD:/src" golang:1.13.3-buster sh -c "cd /src && go build -o mongo-minimal-microservice cmd/main.go"
```

## Run the binary
```sh
# get the command line options
./mongo-minimal-microservice -h

# run the demo microservice with default options
./mongo-minimal-microservice
```

## Request all key/value pairs

Open another terminal and insert and request records

```sh
# insert a key/value pair
curl -X POST -d'{"k":"name","v":"Alexandre"}' localhost:9090/kv

# get all key/value pairs
curl localhost:9090/kv
```
