# iota

A local lightweight serverless function as a sevice engine.

# Running

```sh
# build docker image from dockerfile
docker build -t iota ./worker

# run the api server
./iota

# navigate to http://localhost:8080 or send an api request
curl -X POST http://localhost:8080/run \
    -H "Content-Type: application/json" \
    -d '{"code": "print(2 + 2)"}'
```

## Features
- Warm start: Containers are pre-provisioned
- Concurrency: Handles multiple users via channels
- Self-Healing: Recovers automatically if container crashes
- Cached Queries: Optimized retriggers of same code

## todos
- add observability
- handle code injection