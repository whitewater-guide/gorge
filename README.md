# Gorge

![Release](https://github.com/whitewater-guide/gorge/workflows/Release/badge.svg?branch=master&event=push)

Gorge is a service which harvests hydrological data (river's discharge and water level) on schedule.
Harvested data is stored in database and can be queried later.

## Table of contents

- [Why should I use it?](#why-should-i-use-it-)
- [Data sources](#data-sources)
- [Usage](#usage)
  - [Launching](#launching)
  - [Working with API](#working-with-api)
  - [Available scripts](#available-scripts)
  - [Other](#other)
- [Development](#development)
  - [Inside container](#inside-container)
  - [On host machine](#on-host-machine)
  - [Building and running](#building-and-running)
  - [Writing scripts](#writing-scripts)
- [TODO](#todo)
- [License](#license)

## Why should I use it?

This project is mainly intended for whitewater enthusiasts. Currently, there are several projects that harvest and/or publish hydrological data for kayakers and other river folks. There's certain level of duplication, because these projects harvest data from the same sources. So, if you have a project and want to add new data source(s) to it, you have 3 choices:

1. Write parser/harvester yourself and harvest data yourself
2. Reuse parser/harvester from another project, but harvest data yourself
3. Cooperate with another project to reduce load on the original data source

So how can gorge/whitewater.guide help you? Currently, you can harvest data from whitewater.guide (which uses gorge internally to publish it). It's available via our [GRAPHQL endpoint](https://whitewater.guide/graphql). Please respect the [original data licenses](scripts/README.md). This is option 3.

If you prefer option 2, you can run gorge server in docker container and use our scripts to harvest data, so you don't have to write them yourself.

Gorge was designed with 2 more features in mind. These features are not implemented yet, but they should not take long for us to implement in case someone would like to use them:

- standalone distribution. Gorge can be distributed as standalone linux/mac/windows program, so you can execute it from cli and get harvested results in your stdout. In case you don't want docker and gorge server.
- pushing data downstream. Instead of pulling data from gorge, we can make gorge push data to your project.

## Data sources

You can find the list of our data sources and their statuses [here](scripts/README.md)

## Usage

Gorge is distributed as a ~50Mb [docker image](https://github.com/whitewater-guide/gorge/packages/113546) with two binary files:

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
--http-proxy string        HTTP client proxy (for example, you can use mitm for local development)
--http-timeout int         Request timeout in seconds (default 60)
--http-user-agent string   User agent for requests sent from scripts. Leave empty to use fake browser agent (default "whitewater.guide robot")
--http-without-tls         Disable TLS for some gauges
--log-format string        Set this to 'json' to output log in json (default "json")
--log-level string         Log level. Leave empty to discard logs (default "warn")
--pg-db string             Postgres database (default "postgres")
--pg-host string           Postgres host (default "db")
--pg-password string       Postgres password
--pg-user string           Postgres user (default "postgres")
--pg-without-timescale     During initialization, measurements table will not be transformed into TimescaleDB hypertable
--port string              Port (default "7080")
--redis-host string        Redis host (default "redis")
--redis-port string        Redis port (default "6379")
```

Postgres and redis can also be configured using following environment variables:

- POSTGRES_HOST
- POSTGRES_DB
- POSTGRES_USER
- POSTGRES_PASSWORD
- REDIS_HOST
- REDIS_PORT

Environment variables have lower priority than cli flags.

Gorge uses database to store harvested measurements and scheduled jobs. It comes with postgres and sqlite drivers. Postgres with timescaledb extension is recommended for production. Gorge will initialize all the required tables. Check out sql migration file if you're curious about db schema.

Gorge uses cache to store safe-to-lose data: latest measurement from each gauge and harvest statuses. It comes with redis (recommended) and embedded redis drivers.

Gorge server is supposed to be running in private network. It doesn't support HTTPS. If you want to expose it to public, use reverse proxy.

### Working with API

Below is the list of endpoints exposed by gorge server. You can use `request.http` files in project root and script directories to play with running server.

- `GET /version`

  Returns running server version:

  ```json
  {
    "version": "1.0.0"
  }
  ```

- `GET /scripts`

  Returns array of available scripts with their harvest modes:

  ```json
  [
    {
      "name": "sepa",
      "mode": "oneByOne"
    },
    {
      "name": "switzerland",
      "mode": "allAtOnce"
    }
  ]
  ```

- `POST /upstream/{script}/gauges`

  Lists gauges available for harvest in an upstream source.

  URL parameters:

  - `script` - script name for upstream source

  POST body contains JSON that contains script-specific parameters. For example, it can contain authentication credentials for protected sources. Another example is `all_at_once` test script, which accepts `gauges` JSON parameter to specify number of gauges to return.

  Returns JSON array of gauges. For example:

  ```json
  [
    {
      "script": "tirol", // script name
      "code": "201012", // gauge code in upstream source
      "name": "Lech / Steeg", // gauge name
      "url": "https://apps.tirol.gv.at/hydro/#/Wasserstand/?station=201012", // upstream gauge webpage for humans
      "levelUnit": "cm", // units of water level measurement, if gauge provides water level
      "flowUnit": "cm", // units of water discharge measurement, if gauge provides discharge
      "location": {
        // gauge location in EPSG4326 coordinate system, if provided
        "latitude": 47.24192,
        "longitude": 10.2935,
        "altitude": 1109
      }
    }
  ]
  ```

- `POST /upstream/{script}/measurements?codes=[codes]&since=[since]`

  Harvests measurements directly from upstream source without saving them.

  URL parameters:

  - `script` - script name for upstream source
  - `codes` - comma-separated list of gauge codes to return. This parameter is required for one-by-one scripts. For all-at-once scripts it's optional, and without it all gauges will be returned.
  - `since` - optional unix timestamp indicating start of the period you want to get measurements from. This is passed directly to upstream, if it support such parameter (very few actually do)

  POST body contains JSON that contains script-specific parameters. For example, it can contain authentication credentials for protected sources. Another example is `all_at_once` test script, which accepts `min`, `max` and `value` JSON parameters to control produced values.

  Returns JSON array of measurements. For example:

  ```json
  [
    {
      "script": "tirol", // script name
      "code": "201178", // gauge code
      "timestamp": "2020-02-25T17:15:00Z", // timestamp in RFC3339
      "level": 212.3, // water level value, if provided, otherwise null
      "flow": null // water discharge value, if provided, otherwise null
    }
  ]
  ```

- `GET /jobs`

  Returns array of running jobs:

  ```json
  [
    {
      "id": "3382456e-4242-11e8-aa0e-134a9bf0be3b", // unique job id
      "script": "norway", // job script
      "gauges": {
        // array of gauges that this job harvests
        "100.1": null,
        "103.1": {
          // it's possible to set script-specific parameter for each individual gauge
          "version": 2
        }
      },
      "cron": "38 * * * *", // job's cron schedule, for all-at-once jobs
      "options": {
        // script-specific parameters
        "csv": true
      },
      "status": {
        // information about running job
        "success": true, // whether latest execution was successful
        "timestamp": "2020-02-25T17:44:00Z", // latest execution timestamp
        "count": 10, // number of measurements harvested during latest execution
        "next": "2020-02-25T17:46:00Z", // next execution timestamp
        "error": "somethin went wrong" // latest execution error, omitted when success = true
      }
    }
  ]
  ```

- `GET /jobs/{jobId}`

  URL parameters:

  - `jobId` - harvest job id

  Returns the job description. It's same as item in `/jobs` array, but without `status`

  ```json
  {
    "id": "3382456e-4242-11e8-aa0e-134a9bf0be3b",
    "script": "norway",
    "gauges": {
      "100.1": null,
      "103.1": {
        "version": 2
      }
    },
    "cron": "38 * * * *",
    "options": null
  }
  ```

- `GET /jobs/{jobId}/gauges`

  URL parameters:

  - `jobId` - harvest job

  Returns map object with gauge statuses, where keys are gauge codes and values are statuses:

  ```json
  [
    {
    "010802": {
      "success": false,
      "timestamp": "2020-02-24T18:00:00Z",
      "count": 0,
      "next": "2020-02-25T18:00:00Z"
    }
  ]
  ```

- `POST /jobs`

  Adds new job.

  POST body must contain JSON job description. For example:

  ```json
  {
    "id": "78a9e166-2a73-4be2-a3fb-71d254eb7868", // unique id, must be set by client
    "script": "one_by_one", // script for this job
    "gauges": {
      // list of gauges
      "g000": null, // set to null if gauge has no script-specific options
      "g001": { "version": 2 } // or pass script-specific options
    },
    "options": {
      // optional, common script-specific options
      "auth": "some_token"
    },
    "cron": "10 * * * *" // cron schedule required for all-at-once scripts
  }
  ```

  Returns same object in case of success, error object otherwise

- `DELETE /jobs/{jobId}`

  URL parameters:

  - `jobId` - harvest job id

  Stop the job and deletes it from schedule

- `GET /measurements/{script}/{code}?from=[from]&to=[to]`

  URL parameters:

  - `script` - script name
  - `code` - optional, gauge code
  - `from` - optional unix timestamp indicating start of the period you want to get measurements from. Default to 30 days from now.
  - `to` - optional unix timestamp indicating end of the period you want to get measurements from. Defaults to now.

  Returns array of measurements that were harvested and stored in gorge database for given script (and gauge). Resulting JSON is same as in `/upstream/{script}/measurements`

- `GET /measurements/{script}/{code}/latest`

  URL parameters:

  - `script` - script name, required
  - `code` - gauge code, optional

  Returns array of measurements for given script or gauge. For each gauge, only latest measurement will be returned. Resulting JSON is same as in `/upstream/{script}/measurements`

- `GET /measurements/{script}/{code}/nearest?to=[to]`

  URL parameters:

  - `script` - script name, required
  - `code` - gauge code, optional
  - `to` - required unix timstamp indicating

  For given script and code, returns one measurement that is nearest to timestamp provided via `to` query string. If no measurements +- 1 hour of given timestamps are found, returns null

- `GET /measurements/latest?scripts=[scripts]`

  URL parameters:

  - `scripts` - comma-separated list of script names, required

  Same as `GET /measurements/{script}/{code}/latest` but allows to return latest measurements from multiple scripts at once.

### Available scripts

List of available scripts is [here](scripts/README.md)

### Other

There're Typescript type definitions for the API available on [NPM](https://www.npmjs.com/package/@whitewater-guide/gorge)

## Development

### Inside container

Preferred way of development is to develop inside docker container. I do this in [VS Code](https://code.visualstudio.com/docs/remote/containers). The repo already contains `.devcontainer` configuration.

If you use `docker-compose.yml` you need `.env.development` file where you can put env variables with secrets for scripts. The app will work without those variables, but docker-compose requires `.env.development` file to be present. If you use VS Code, `.devcontainer` takes care of this.

Some tests require postgres. You cannot run them inside docker container (unless you want to mess with docker-inside-docker). They're excluded from main test set, I run them using `make test-nodocker` from host machine or CI environment.

Docker-compose stack comes with [mitmproxy](https://mitmproxy.org/). You can monitor your development server requests at `http://localhost:6081` on host machine.

### On host machine

If you want to develop on host machine, you'll need following libraries installed on it (they're installed in docker image, see Dockerfile for more info):

- [libproj](https://proj.org/) shared library, to convert coordinate systems. Currently version 5.2.0 is required. (to match version from debian buster). On MacOS this can be installed via brew:

```
brew tap-new $USER/local-tap
brew extract --version='5.2.0' proj $USER/local-tap
brew install proj@5.2.0
```

Also you'll need following go tools:

- [modd](https://github.com/cortesi/modd) - live reloading tool, not really required, but some might prefer such workflow
- [golangci-lint](github.com/golangci/golangci-lint) - not a requirement, but this is the linter of choice and CI uses it

These tools are installed locally (see `tools.go`), but you should make sure that binaries are in your `PATH`

### Building and running

Take a look at [Makefile](Makefile). Here are the highlights:

- `make run` builds and launches server and cli, provides live reloading and tests
- `make build` builds `gorge-server` and `gorge-cli` binaries in `/go/bin` directory
- `make test` runs all tests expect postgres tests
- `make test-nodocker` runs all test including postgres tests
- `make lint` runs linter

### Writing scripts

Here are some recommendations for writing scripts for new sources

- Examine `testscripts` package, which contains simplest test scripts. Then examine existing scripts. Some of them process JSON and CSV sources, other parse raw HTML pages.
- Write tests, but when testing, **do not use** calls to real URLs, because unit tests can flood upstream with requests
- Round locations to 5 digits precision [link](https://en.wikipedia.org/wiki/Decimal_degrees), round levels and flows to what seems reasonable
- When converting coordinates, use `core.ToEPSG4326` utility function. It uses [PROJ](https://proj.org/) internally
- Use `core.Client` http client, which sets timeout, user-agent and has various helpers
- Do not bother with sorting results - this is done by script consumers
- Do not filter by `codes` and `since` inside worker. They are meant to be passed to upstream. Empty `codes` for all-at-once script must return all available measurements.
- Return null value (`nulltype.NullFloat64{}`) for level/flow when it's not provided
- Pay extra attention to time zones!
- Pass variables like access keys via script options, but provide environment variable fallbacks
- Provide sample http requests (see `requests.http` files)
- Be forgiving when handling errors: only exit harvest function on real stoppers. If a single JSON object/CSV line causes error - log it then process next entry.

## TODO

- Notify when some script seem to be broken
- Build this using github actions _without_ docker. Problem: ubuntu 18.04 has old version of libproj-dev
- Virtual gauges
  - Statuses
  - What happens when one component is broken?
- Authorization
- Pushing data downstream to peer projects (webhooks)
- Subscriptions
- Advanced scheduling with time affinity
- Scripts as Go plugins
- Send logs to sentry
- Per-script binaries for third-party consumption
- Add description attribute to scripts (cause nzbop is ugly)
- GRAPHQL api with sources and local gauges
- (DX) add git pre-push hooks to test and lint

## License

[MIT](LICENSE)
