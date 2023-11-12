# Nagios-Plugin exporter

The nagios-plugin exporter provides Prometheus-compatible metrics using Nagios/Monitoring plugins.


## Running this software

### From binaries

[Build](#Building the software) the exporter

Then:

    ./prometheus-nagios-plugin-exporter <flags>


### Checking the results

Visiting [http://localhost:9665/probe?http_host=google.com&module=http](http://localhost:96655/probe?http_host=google.com&module=http)
will return metrics for a nagios-plugin probe named *htt* against google.com.
The `nagios_plugin_probe_success` metric indicates if the probe succeeded.

Metrics concerning the operation of the exporter itself are available at the
endpoint <http://localhost:9665/metrics>.

### Debugging probe requests

Adding a `debug=true` parameter to the probe endpoint will return debug information for that probe,
assuming, the feature has been enabled. By default it not active for security reasons, but can
be enabled using the `--web.debug` commandline argument.

### TLS and basic authentication

The Nagios-Plugin Exporter supports TLS and basic authentication. This enables better
control of the various HTTP endpoints.

To use TLS and/or basic authentication, you need to pass a configuration file
using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

Note that the TLS and basic authentication settings affect all HTTP endpoints:
/metrics for scraping, /probe for probing, and the web UI.

## Building the software

### Local Build

    make


## [Configuration](docs/CONFIGURATION.md)

Nagios-Plugin exporter is configured via a [configuration file](docs/CONFIGURATION.md) and command-line flags
(such as what configuration file to load, what port to listen on, and the logging format and level).

Nagios-Plugin exporter can reload its configuration file at runtime. If the new configuration is not well-formed, the changes will not be applied.
A configuration reload is triggered by sending a `SIGHUP` to the Nagios-Plugin exporter process or by sending a HTTP POST request to the `/-/reload` endpoint.

To view all available command-line flags, run `./prometheus-nagios-plugin-exporter -h`.

To specify which [configuration file](docs/CONFIGURATION.md) to load, use the `--config.file` flag.

Additionally, an [example configuration](example.yml) is also available.

The exporter itself does not perform any monitoring, but instead calls executables
following the [Nagios Plugin conventions](https://nagios-plugins.org/doc/guidelines.html).
The monitoring scope of this exporte is therefor limited to the available plugin executables.

The timeout of each probe is automatically determined from the `scrape_timeout` in the [Prometheus config](https://prometheus.io/docs/operating/configuration/#configuration-file),
slightly reduced to allow for network delays. This can be further limited by the `timeout` in the Nagios-Plugin exporter config file. If neither is specified, it defaults to 120 seconds.

## Prometheus Configuration

Nagios-plugin exporter implements the multi-target exporter pattern, so we advice
to read the guide [Understanding and using the multi-target exporter pattern
](https://prometheus.io/docs/guides/multi-target-exporter/) to get the general
idea about the configuration.

Most Nagios plugins require some kind of target information, which can be
done with relabelling.

Example config:
```yml
scrape_configs:
  - job_name: 'nagios_plugin'
    metrics_path: /probe
    params:
      module: [http]  # Using the check_http plugin
    static_configs:
      - targets:
        - http://prometheus.io    # Target to probe with http.
        - https://prometheus.io   # Target to probe with https.
        - http://example.com:8080 # Target to probe with http on port 8080.
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_http_host
      - source_labels: [__param_http_host]
        target_label: instance
      - target_label: __address__
        replacement: 127.0.0.1:9665  # The nagios_plugin exporter's real hostname:port.
  - job_name: 'nagios_plugin_exporter'  # collect nagios_plugin exporter's operational metrics.
    static_configs:
      - targets: ['127.0.0.1:9665']
```

## Permissions

Given that the Nagios-plugin exporter simple executes scripts, special
adjustments and preparations must be on that front (e.g. ensuring the
`net_raw` capability is set on the `check_icmp` nagios plugin)

If a Nagios plugin requires elevated privileges, the only way to ensure
proper execution, is to run the exporter with elevated privileges as well.

