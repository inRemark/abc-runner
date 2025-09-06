# redis-runner

[English](README.md) | [中文](README_zh.md)

## About

a benchmark tools for redis.

## Features

- support redis cluster mode use `--cluster`
- support redis sentinel and standalone mode default
- global self increasing key when r=0, random key when r>0
- operation set_get_random, set, get, pub, sub use `-t <option>`
- support ttl use `-ttl <option>`
- support read percentage use `-R <option>`

## Usage

```bash
./redis-runner-macos-arm64 redis --cluster -h localhost -p 6371 -a pwd@redis -n 100000 -c 10 -d 64 -r 0 -R 50 -ttl 120 -t set_get_random
```

```bash
Redis Info: mode:cluster, host:localhost, port:6371, password:pwd@redis, total:100000, parallel:10, dataSize:64, random:0, readPercent:50, ttl:120
Progress: 100000 / 100000, 100.00%

All 100000 request have completed. 
Parameters: 
Total: 100000, Parallel: 10, ReadPercent: 50 DataSize: 64, TTL: 120
Statistics: 
Read count: 49687, Write count: 50313, Generated keys: 50149, repeat Keys:0
Summary: 
rps: 65490, avg: 0.121ms, min: 0.015ms, p90: 0.200ms, p95: 0.250ms, p99: 0.495ms, max: 18.672ms
Completed.
```

## Options

```bash
./redis-runner-macos-arm64 --help
Usage: redis-runner [--help] [--version] <command> [options]
Sub Commands:

  redis      Redis client benchmark 
  http       HTTP client benchmark

Use tool help <command> for more information about a command.

Usage: redis-runner redis [options]
  --cluster     redis cluster mode (default false)
  -h            Server hostname (default 127.0.0.1) 
  -p            Server port (default 6379)
  -a            Password for Redis Auth
  -n            Total number of requests (default 100000)
  -c            Number of parallel connections (default 50)
  -d            Data size of SET/GET value in bytes (default 3)
  -r            Read operation percent (default 100%)
  -ttl          TTL in seconds (default 300)

Usage: redis-runner http [arguments]
  -url          Server url (default http://localhost:8080)
  -m            Method get/post (default get)
  -n            Total number of requests (default 10000)
  -c            Number of parallel connections (default 50)
```

## License
[MIT](LICENSE)