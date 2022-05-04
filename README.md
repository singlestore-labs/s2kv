# key/value API implemented on top of SingleStore

Start SingleStore somewhere (e.g. docker, managed service) and run `./schema.sql` and `./procedures.sql`. Modify `s2kv/cmd/s2kv/main.go` to have correct connection details.

Run s2kv like so:
```bash
go build s2kv/cmd/s2kv
./s2kv
```

Use `redis-benchmark` for perf testing:

```bash
redis-benchmark -t set,get -n 100000 -q
```

Use `redis-cli` for manual testing.