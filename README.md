# Gorge

![Build and test Go](https://github.com/whitewater-guide/gorge/workflows/Build%20and%20test%20Go/badge.svg)

Gorge is a service which harvests hydrological data (river's discharge and water level) on schedule.
Harvested data is stored in database and can be queried later.

## Usage

Gorge is distributed as docker image with two binary files:

- `gorge-server` (_entrypoint_) - web server with REST API
- `gorge-cli` - command-line client for this server. Since image is distroless, use `docker exec gorge gorge-cli` to call it

### Launching

`gorge-server` accepts configuration via cli arguments (use `gorge-server --help`). You can pass them via docker-compose command field, like this:

```yaml
command:
  [
    "--pg-db",
    "gorge",
    "--debug",
    "--log-format",
    "plain",
    "--db-chunk-size",
    "1000",
  ]
```

Here is the list of available flags:

```
--cache string             Either 'inmemory' or 'redis' (default "redis")
--db string                Either 'inmemory' or 'postgres' (default "postgres")
--db-chunk-size int        Measurements will be saved to db in chunks of this size. When set to 0, they will be saved in one chunk, which can cause errors
--debug                    Enables debug mode, sets log level to debug
--endpoint string          Endpoint path (default "/")
--http-timeout int         Request timeout in seconds (default 60)
--http-user-agent string   User agent for requests sent from scripts. Leave empty to use fake browser agent (default "whitewater.guide robot")
--log-format string        Set this to 'json' to output log in json (default "json")
--log-level string         Log level. Leave empty to discard logs (default "warn")
--pg-db string             Postgres database (default "postgres")
--pg-host string           Postgres host (default "db")
--pg-password string       Postgres password
--pg-user string           Postgres user (default "postgres")
--port string              Port (default "7080")
--redis-host string        Redis host (default "redis")
--redis-port string        Redis port (default "6379")
```

Postgres and redis can also be configured using folowing environment variables:

- POSTGRES_HOST
- POSTGRES_DB
- POSTGRES_USER
- POSTGRES_PASSWORD
- REDIS_HOST
- REDIS_PORT

Environment variables have lower priority than cli flags.

Gorge uses database to store harvested measurements and scheduled jobs. It comes with postgres and sqlite drivers. Postgres with timescaledb extension is recommended for production. Gorge will initialize all the required tables. Check out sql migration file if you're curious about db schema.

Gorge uses cache to store safe-to-lose data: latest measurement each gauge and harvest statuses. It comes with redis (recommended) and embedded redis drivers.

## Development

Preferred way of development is to develop inside docker container. I do this in [VS Code](https://code.visualstudio.com/docs/remote/containers). There's a compose file for this purpose.

There's a [modd](https://github.com/cortesi/modd) tool installed in dev image, which enables liver reloading and tests. Start it using `make run`.

If you want to develop on host machine, you'll need following tools installed on it (they're installed in docker image, see Dockerfile for more info):

- [libproj](https://proj.org/) shared library, to convert coordinate systems
- [go-bindata](https://github.com/go-bindata/go-bindata) to embed sql scripts
- [modd](https://github.com/cortesi/modd) it's actually optional

Some tests require postgres. You cannot run them inside docker container (unless you want to mess with docker-inside-docker). They're excluded from main test set, I run them using `make test-nodocker` from host machine or CI environment.

### Writing scripts

Here are some recommendations for writing scripts for new sources

- Write tests, but when testing, **do not use** calls to real URLs, because unit tests can flood upstream with requests
- Round locations to 5 digits precision [link](https://en.wikipedia.org/wiki/Decimal_degrees), round levels and flows to what seems reasonable
- When converting coordinates, use core.ToEPSG4326 utility function. It uses [PROJ](https://proj.org/) internally
- Use `core.Client` http client, which sets timeout, user-agent and has various helpers
- Do not bother with sorting results - this is done by script consumers
- Do not filter by `codes` and `since` inside worker. They are meant to be passed to upstream. Empty `codes` for all-at-once script must return all available measurements.
- Return null value (`nulltype.NullFloat64{}`) for level/flow when it's not provided
- Pay extra attention to time zones!
- Pass variables like access keys via script options
- Provide sample http requests (see `requests.http` files)

## Env variables

Container makes use of following env variables. Env variables have lesser priority than config values.

| Name              | Default value | Desription                                  |
| ----------------- | ------------- | ------------------------------------------- |
| POSTGRES_HOST     |               | Postgres connection details - host          |
| POSTGRES_DB       |               | Postgres connection details - database name |
| POSTGRES_USER     |               | Postgres connection details - user          |
| POSTGRES_PASSWORD |               | Postgres connection details - password      |
| REDIS_HOST        | redis         | Redis connection details - host             |
| REDIS_PORT        | 6379          | Redis connection details - port             |

## TODO

- Build this using github actions _without_ docker. Problem: ubuntu 18.04 has old version of libproj-dev
- Virtual gauges
  - Statuses
  - What happens when one component is broken?
- Authorization
- Pushing
- Subscriptions
- Advanced scheduling, new harvest mode: batched
- Scripts as Go plugins
- Send logs to sentry
- Per-script binaries for third-party consumption
- Autogenerate typescript definitions
- Add "upstream" JSON field to allow upstream methods pass arbitrary fields
