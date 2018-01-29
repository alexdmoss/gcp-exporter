# Google Cloud Platform - Prometheus exporter

[![pipeline status](https://gitlab.com/gitlab-org/ci-cd/gcp-exporter/badges/master/pipeline.svg)](https://gitlab.com/gitlab-org/ci-cd/gcp-exporter/commits/master)
[![coverage report](https://gitlab.com/gitlab-org/ci-cd/gcp-exporter/badges/master/coverage.svg)](https://gitlab.com/gitlab-org/ci-cd/gcp-exporter/commits/master)

`gcp-exporter` is a Prometheus exporter for Google Cloud Platform resources.

This tool looks for defined GCP resources and exports metrics about them

## Usage

First, you need to create [a Service Account][gcp-service-account] for the exporter. Notice,
that regarding to chosen collectors, you will need to assign proper permissions to the
Service Account.

Having the Service Account, download the JSON file with its credentials.

### Command line options

The general syntax of command line parameters is as following:

```
gcp-exporter [global-options] command [command-options] [arguments]
```

#### General command line options

| Name                | Type | Required? | Description |
|---------------------|------|-----------|-------------|
| `--debug`           | bool | no        | Show debug runtime information |
| `--no-color`        | bool | no        | Disable logs coloring |
| `--help` or `-h`    | bool | no        | Show help and exit |
| `--version` or `-v` | bool | no        | Show version info and exit |

#### Commands

##### `help` command

Shows general help (similar to `--help` flag). It can be used also together with command
name. In that case it will show usage information about the selected command.

##### `start` command

Starts the exporter.

_command options_

| Name                           | Type    | Required? | Description |
|--------------------------------|---------|-----------|-------------|
| `--listen`                     | string  | yes       | Listen address for metrics and debug HTTP server (e.g. "0.0.0.0:1234") |
| `--interval`                   | integer | no        | Number of seconds between requesting data from GCP (default: `60`) |
| `--service-account-file`       | string  | no        | Path to GCP Service Account JSON file (default: `~/.google-service-account.json`) |
| `--instances-collector-enable` | bool    | no        | Enables instances collector |
| `--project`                    | string  | no        | Count instances that belong to selected project; may be used multiple times |
| `--zone`                       | string  | no        | Count instances that belong to selected zone; may be used multiple times |
| `--match-tag`                  | string  | no        | Count instances that are matching selected tag; may be used multiple times |

1. Instances collector will look for instances for all defined `project+zone` pairs.
1. If `match-tag` is used, then an instance will be counted if it matches any of specified tags.

**Example usage** 

```bash
$ /opt/prometheus/gcp-exporter/gcp-exporter start \
    --listen :9393 \
    --interval 15 \
    --service-account-file /opt/prometheus/gcp-exporter/service-account-file.json \
    --instances-collector-enable \
    --project project-id-1 \
    --project project-id-2 \
    --zone us-east1-c \
    --zone us-east1-d \
    --match-tag docker-machine
```

##### `get-token` command

Allows to get the oAuth2 Token, using specified Service Account JSON file. The token
may be next used to do a manual API calls (e.g. with `curl`).

_command options_

| Name                           | Type    | Required? | Description |
|--------------------------------|---------|-----------|-------------|
| `--service-account-file`       | string  | no        | Path to GCP Service Account JSON file (default: `~/.google-service-account.json`) |

**Example usage**

```bash
$ /opt/prometheus/gcp-exporter/gcp-exporter get-token \
    --service-account-file /opt/prometheus/gcp-exporter/service-account-file.json
```

## Using Docker container

Prepared Docker image is configured to run the `start` command. To make it working you should
remember about passing the Service Account JSON file to the container. It can be done with
volumes feature.

If you want to access the exporter from a remote Prometheus server, you should also remember
about exposing the port.


**Example**

```bash
$ docker run -d \
         --restart always \
         --name gcp_exporter \
         --log-driver=syslog \
         --log-opt tag=gcp_exporter \
         -e NO_COLOR=true \
         -v /opt/prometheus/gcp-exporter/service-account-file.json:/service-account-file.json \
         -p 9393:9393 \
         registry.gitlab.com/gitlab-org/ci-cd/gcp-exporter:0.1 \
            --listen :9393 \
            --interval 15 \
            --service-account-file /service-account-file.json \
            --instances-collector-enable \
            --project project-id-1 \
            --project project-id-2 \
            --zone us-east1-c \
            --zone us-east1-d \
            --match-tag docker-machine
```

## Author

Tomasz Maczukin, 2018, GitLab

## License

MIT

[gcp-service-account]: https://cloud.google.com/compute/docs/access/service-accounts