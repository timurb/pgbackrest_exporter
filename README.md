# pgBackRest Exporter

Prometheus exporter for [pgBackRest](https://pgbackrest.org/).

The metrics are collected by parsing result of `pgbackrest info --output json` command. The metrics are collected for all stanzas received by command. You need to run exporter in the same host where pgBackRest was installed.

Tested for pgBackRest 2.33.

## Collected metrics

The metrics provided by the client.

* `pgbackrest_exporter_stanza_status` - current stanza status.
* `pgbackrest_exporter_repo_status` - current repo status by stanza.

    Values description for stanza and repo statuses:
    - 0: ok,
    - 1: missing stanza path,
    - 2: no valid backups,
    - 3: missing stanza data,
    - 4: different across repos,
    - 5: database mismatch across repos,
    - 6: requested backup not found,
    - 99: other.

* `pgbackrest_exporter_backup_info` - backup info by stanza and backup type.
    
    Values description:
     - 1 - info about backup is exist.

* `pgbackrest_exporter_backup_duration` - backup duration in seconds by stanza and backup type.
* `pgbackrest_exporter_backup_size` - backup size by stanza and backup type.
* `pgbackrest_exporter_backup_database_size` - database size in backup by stanza and backup type.
* `pgbackrest_exporter_backup_repo_backup_set_size` - repo set size in backup by stanza and backup type.
* `pgbackrest_exporter_backup_repo_backup_size` - repo size in backup by stanza and backup type.
* `pgbackrest_exporter_wal_archive_status` - WAL archive status by stanza.

    Values description:
    - 0 - any one of WALMin and WALMax have empty value, there is no correct information about WAL archiving,
    - 1 - both WALMin and WALMax have no empty values, there is correct information about WAL archiving.

    When flag `--verbose.info` is specified - WALMin and WALMax are added as metric labels.
    This creates new different time series on each WAL archiving.

## Getting Started
### Building and running

```bash
git clone https://github.com/woblerr/pgbackrest_exporter.git
cd pgbackrest_exporter
make build
./pgbackrest_exporter <flags>
```

Available configuration flags:

```bash
./pgbackrest_exporter --help
usage: pgbackrest_exporter [<flags>]

Flags:
  --help                      Show context-sensitive help (also try --help-long and --help-man).
  --prom.port="9854"          Port for prometheus metrics to listen on.
  --prom.endpoint="/metrics"  Endpoint used for metrics.
  --collect.interval=600      Collecting metrics interval in seconds.
  --verbose.info              Enable additional metrics labels.
```

### Building and running docker
By default, pgBackRest version is `2.33`. Another version can be specified via arguments.
For base image used [docker-pgbackrest](https://github.com/woblerr/docker-pgbackrest) image.

```bash
make docker
```

or for specific pgBackRest version

```bash
docker build -f Dockerfile --build-arg BACKREST_VERSION=2.33 -t pgbackrest_exporter .
```

Environment variables supported by this image:
* all environment variables from [docker-pgbackrest](https://github.com/woblerr/docker-pgbackrest#docker-pgbackrest)  image;
* `EXPORTER_ENDPOINT` - metrics endpoint, default `/metrics`;
* `EXPORTER_PORT` - port for prometheus metrics to listen on, default `9854`;
* `COLLECT_INTERVAL` - collecting metrics interval in seconds, default `600`;

You will need to mount the necessary directories or files inside the container.

Simple run:

```bash
docker run -d \
    --name pgbackrest_exporter \
    -p 9854:9854 \
    -v  /etc/pgbackrest/pgbackrest.conf:/etc/pgbackrest/pgbackrest.conf \ 
    pgbackrest_exporter
```

With some enviroment variables:

```bash
docker run -d \
    --name pgbackrest_exporter \
    -e BACKREST_USER=postgres \
    -e BACKREST_UID=1001 \
    -e BACKREST_GROUP=postgres \
    -e BACKREST_GID=1001 \
    -e TZ=America/Chicago \
    -e COLLECT_INTERVAL=60 \
    -p 9854:9854 \
    -v  /etc/pgbackrest/pgbackrest.conf:/etc/pgbackrest/pgbackrest.conf \
    pgbackrest_exporter
```

### Running as systemd service

* Register `pgbackrest_exporter` (already builded, if not - exec `make build` before) as a systemd service:

```bash
 make make prepare-service
```

Validate prepared file `pgbackrest_exporter.service` and run:

```bash
sudo make install-service
```

* View service logs:

```bash
journalctl -u pgbackrest_exporter.service
```

* Delete systemd service:

```bash
sudo make remove-service
```

---
Manual register systemd service:

```bash
cp pgbackrest_exporter.service.template pgbackrest_exporter.service
```

In file `pgbackrest_exporter.service` replace `{PATH_TO_FILE}` to full path to `pgbackrest_exporter`.

```bash
sudo cp pgbackrest_exporter.service /etc/systemd/system/pgbackrest_exporter.service
sudo systemctl daemon-reload
sudo systemctl enable pgbackrest_exporter.service
sudo systemctl restart pgbackrest_exporter.service
systemctl -l status pgbackrest_exporter.service
```
