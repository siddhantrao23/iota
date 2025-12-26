# iota

A local lightweight serverless function as a sevice engine.

# Running

```sh
# build docker image from dockerfile
docker build -t iota ./worker

# run the api server
./iota

# send an api request
curl -X POST http://localhost:8080/run \
    -H "Content-Type: application/json" \
    -d '{"code": "print(2 + 2)"}'
```

## Features
- Warm start: containers are pre-provisioned
- Concurrency: Handles multiple users via channels
- Self-Healing: Recovers automatically if container crashes
- Graceful Shutdown: Cleans up containers on exit

## todos
- add a cache that hashes input string to improve perf
- add observability
- handle code injection
- add a frontend