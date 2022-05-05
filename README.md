# Demo of a key/value API implemented on top of SingleStore

> 👋 Hello! I'm [@carlsverre][gh-carlsverre] ([twitter][tw-carlsverre]), and I'll be helping you out today while we talk about building a key/value API on top of SingleStore. If you are playing with this project on your own and get stuck, feel free to ask for help in the [SingleStore forums][s2-forums] or via a [Github Issue][gh-issue]

This project implements a key/value style API using a combination of stored procedures and user defined functions on top of [SingleStore][s2]. This project is **not production ready** and should only be used as an example of what you can do with SingleStore.

For fun, this project uses [go-redisproto][redisproto] to simulate the Redis protocol for easy integration with Redis tools and libraries. By no means does this project claim to support the entirety of Redis' semantics or is trying to be a replacement for Redis in its current form. Certain workloads will run best on Redis while other workloads will benefit more from using SingleStore. Both technologies bring a different set of tradeoffs to the table.

Some things to note about the key/value (KV) implementation in this project:

**Values are stored using Universal Storage.** This takes advantage of both row and column oriented storage under the hood. Since it is a hybrid storage engine it is able to provide reasonable OLTP performance while letting us run analytics over the data. As an example of this, I added a non-Redis command called `SWITHMEMBER` which is able to retrieve all of the sets containing the provided value. Doing this in Redis either requires running `SISMEMBER` on every set or [manually maintaining a reverse index that maps values to keys][so-redis-swithmember]. SingleStore is able to do this in a single query and take advantage of automatic indexes as well as vectorized execution.

**We use a table for each key kind.** This allows us to take advantage of columnar storage and relational algebra to accelerating queries like set intersection. You could easily extend this to supporting other native types such as integers, JSON, or timestamps.

**The schema is sharded by key.** SingleStore is a distributed system, and as such we need to decide how to shard the data. Similar to Redis Cluster, s2kv shards the data by key. This allows us to scale up to extremely high numbers of keys while gaining the many benefits of a distributed system such as concurrency, availability, and durability. 

# Run s2kv yourself!

This project is easy to run and I encourage you to play with it a bit. Once you get it setup you can easily add additional commands or play with different ways to optimize the queries or schema. Let's get started!

## Dependencies

This git repo includes a [VS Code development container][vscode-devcontainer] configuration. This means that if you open this repo using VS Code the entire development environment can be automatically setup for you.

If you want to run s2kv without using a dev container you will need to provision the following dependencies:

 * [golang][golang] 1.17 or above
 * [mockgen][mockgen]
 * [redis-cli][redis-cli]
 * [redis-benchmark][redis-benchmark]

## Setting up SingleStore

Since s2kv is implemented on top of SingleStore, we need a SingleStore cluster to connect to. We recommend either running it locally in a docker image or using the SingleStore Managed Service. See the guides below for details:

### Using the SingleStore Managed Service

1. [Sign up][try-free] for $500 in free managed service credits.
2. Create a S-00 sized cluster in [the portal][portal]
3. Copy `config.example.toml` to a new file and edit it to match:

```toml
[database]
host = "THE CONNECTION ENDPOINT"
port = "3306"
username = "admin"
password = "THE ADMIN PASSWORD"
database = "kv"
```

### Using the SingleStore cluster-in-a-box Docker image

**This will not work on a Mac M1 or ARM hardware**

1. [Sign up][try-free] for a free SingleStore license. This allows you to run up to 4 nodes up to 32 gigs each for free. Grab your license key from [SingleStore portal][portal] and set it as an environment variable.

   ```bash
   export SINGLESTORE_LICENSE="singlestore license"
   ```

2. Start a SingleStore [cluster-in-a-box][ciab] using Docker:

   ```bash
   docker run -it \
       --name ciab \
       -e LICENSE_KEY=${SINGLESTORE_LICENSE} \
       -e ROOT_PASSWORD=test \
       -p 3306:3306 -p 9000:9000 -p 8080:8080 \
       singlestore/cluster-in-a-box
   docker start ciab
   ```

3. Get the private ip address of the docker container you just started

```bash
docker inspect -f '{{ .NetworkSettings.IPAddress }}' ciab
```

4. Copy `config.example.toml` to a new file and edit it to match:

```toml
[database]
host = "THE PRIVATE IP OF THE DOCKER CONTAINER"
port = "3306"
username = "root"
password = "test"
database = "kv"
```

## Initialize the schema

Using the SQL editor (in the [portal][portal]) or via the mysql CLI run the contents of [schema.sql](schema.sql) and [procedures.sql](procedures.sql) against the database. Here is how I would do this using the mysql CLI:

```bash
mysql -u root -h 172.17.0.4 -ptest <schema.sql <procedures.sql
```

## Run tests

To make sure everything is working you can run tests like so:

```bash
go test -config PATH_TO_YOUR_CONFIG_FILE
```

## Run s2kv

In a terminal you can start s2kv like so:

```bash
go build s2kv/cmd/s2kv
./s2kv -config PATH_TO_YOUR_CONFIG_FILE
```

## Connect with redis-cli

While s2kv is running you can simply run `redis-cli` to connect:

```
$ redis-cli
127.0.0.1:6379> set foo bar
OK
127.0.0.1:6379> get foo
"bar"
127.0.0.1:6379> set i 1
OK
127.0.0.1:6379> incrby i 2
(integer) 3
127.0.0.1:6379> get i
"3"
127.0.0.1:6379> sadd set 1
OK
127.0.0.1:6379> sadd set 2
OK
127.0.0.1:6379> sadd set 3
OK
127.0.0.1:6379> scard set
(integer) 3
127.0.0.1:6379> sadd bar 2
OK
127.0.0.1:6379> sadd bar 3
OK
127.0.0.1:6379> sadd bar 4
OK
127.0.0.1:6379> sinter set bar
1) "3"
2) "2"
127.0.0.1:6379> quit
```

## Use `redis-benchmark` to run many commands quickly

Not all of Redis's API is implemented so take errors output by redis-benchmark with a grain of salt (most can be ignored).

**Note on performance:** SingleStore is distributed so should only be compared to Redis Cluster if at all. In general, these are two very different systems and thus **this tool should not be used to compare Redis and SingleStore**. If you have a workload which would benefit from SingleStore's semantics - please reach out to us so we can make sure you get the performance you need.

```bash
$ redis-benchmark -t set,get -n 100000 -q
ERROR: command not supported
ERROR: failed to fetch CONFIG from 127.0.0.1:6379
WARN: could not fetch server CONFIG
SET: 5505.70 requests per second
GET: 42158.52 requests per second

$ redis-benchmark -P 16 -n 10000000 -r 10000000 -q sadd bar __rand_int__
ERROR: command not supported
ERROR: failed to fetch CONFIG from 127.0.0.1:6379
WARN: could not fetch server CONFIG
sadd bar __rand_int__: 26155.21 requests per second
```

## What's next?

If you got this far I encourage you to dive into the code and start figuring out how it works. Most of the actual key/value logic is contained in [schema.sql](schema.sql) and [procedures.sql](procedures.sql). The server which hosts the Redis API can be found in the various go files. I suggest looking at [commands.go](commands.go) which contains the logic for each of the supported Redis commands. If you want an example of adding a new command, check out [implement_decrby.md](implement_decrby.md). Have fun!

<!-- link index -->

[s2]: https://www.singlestore.com
[redisproto]: https://github.com/secmask/go-redisproto
[vscode-devcontainer]: https://code.visualstudio.com/docs/remote/containers
[mockgen]: https://github.com/golang/mock
[golang]: https://go.dev
[redis-cli]: https://redis.io/docs/manual/cli/
[redis-benchmark]: https://redis.io/docs/reference/optimization/benchmarks/
[try-free]: https://www.singlestore.com/try-free/
[ciab]: https://github.com/memsql/deployment-docker
[portal]: https://portal.singlestore.com/
[so-redis-swithmember]: https://stackoverflow.com/a/59377541/65872
[gh-carlsverre]: https://www.github.com/carlsverre
[tw-carlsverre]: https://www.twitter.com/carlsverre
[s2-forums]: https://www.singlestore.com/forum/
[gh-issue]: issues