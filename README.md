# Redis API implemented on top of SingleStore

Start SingleStore somewhere (e.g. docker, managed service) and run `./schema.sql` and `./procedures.sql`. Modify `s2redis/cmd/s2redis/main.go` to have correct connection details.

Run s2redis like so:
```bash
go build s2redis/cmd/s2redis
./s2redis
```

Use `redis-benchmark` for perf testing:

```bash
redis-benchmark -t set,get -n 100000 -q
```

Use `redis-cli` for manual testing.