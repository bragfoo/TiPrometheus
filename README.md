# TiPrometheus

A adapter/gateway for [TiKV](https://github.com/tikv/tikv).

Supports remote write and read.

Transport security is optional and can be enabled both between
Prometheus and the adapter and for communication with the
TiKV cluster.

## Quick Start

Build:
```
go build -o tiprometheus cmd/tiprometheus/app.go
```

Rename the example configuration `conf.example.toml` to
`conf.toml` and adapt the parameters to your needs.

Run with:
```
./tiprometheus -conf conf.toml
```

## Configuration

TiPrometheus is configured through a TOML file.
See `conf.example.toml` for an example.

### Options

* **PDHost**
  The PD cluster host to connect to.
  TiPrometheus does not connect directly to the TiKV cluster,
  but uses the PD cluster to obtain a node address.
* **TimeInterval**
  (Unknown)
* **AdapterListen**
  Which IP address and port the adapter should listen on,
  separated by `:`.
  To listen on 0.0.0.0, only write `:` followed by the port.
* **AdapterEnableTLS**
  If true, TLS will be enabled on the Prometheus connection.
  AdapterServerCertificate and AdapterServerKey must
  be specified as well.
* **AdapterCACertificate**
  The CA certificate used to validate connections from Prometheus.
  If not specified, all clients can connect without authentication.
* **AdapterServerCertificate**
  The server certificate to use for the listener.
* **AdapterServerKey**
  The key for AdapterServerCertificate
* **TiKVEnableTLS**
  If true, TLS will be enabled on the PD/TiKV connection.
  TiKVCACertificate, TiKVClientCertificate and TiKVClientKey
  must be specified as well.
* **TiKVCACertificate**
  The CA certificate used to validate the server certificate
  sent by PD/TiKV.
* **TiKVClientCertificate**
  The client certificate to use for authentication on PD/TiKV.
* **TiKVClientKey**
  The key for TiKVClientCertificate.

### Sections

The configuration file can contain multiple sections to
allow quick switching between development/production
environments.

Set the environment variable `RunTime` to the section
you would like to enable:

Run with:
```
RunTime=dev ./tiprometheus -conf conf.toml
```

