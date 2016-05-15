### Overview

csproxy is a proxy/bridge between the Graphite Carbon protocol and statsd and is
suited for real-time metrics that can tolerate loss of timestamp in
the translation. It optionally provides the ability to perform
regexp transformations on the metrics before they are emitted to statsd
or Carbon.  It is originally motivated by wanting to use collectd to send
metrics out to datadog, which provides a local statsd (dogstatsd) listener
without writing a collectd plugin that transformed metric names.

csproxy accepts metrics on the carbon protocol in the format:

```
# path ts val
collectd.host_example_net.interface-en2.if_octets.tx 0.000000 1463321399
```

and emits those metrics back out to statsd as a gauge and carbon:

```
# statsd over UDP (with a transform to remove the hostname)
# path:val|g
collectd.interface-en2.if_octets.tx:0.000000|g

# back out to carbon over TCP
# path ts val
collectd.host_example_net.interface-en2.if_octets.tx 0.000000 1463321399
```

Example from the log:

```
2016/05/15 10:10:49 Received metric: path: collectd.host_example_net.interface-bridge0.if_octets.tx value: 0.000000 timestamp: 2016-05-15T10:10:49-04:00
2016/05/15 10:10:49 Sending metrics to statsd: collectd.interface-bridge0.if_octets.tx:0.000000|g
2016/05/15 10:10:49 Sending metric to carbon: collectd.host_example_net.interface-bridge0.if_octets.tx 0.000000 1463321449
```

### Configuration

Create `csproxy.yaml` in the same dir as the binary.  It should look like this:

```yaml
---
# listeners:
#   carbon:
#     address: 127.0.0.1
#     port: 2003
#   http:
#     address: 127.0.0.1
#     port: 9080
writers:
  statsd:
    address: localhost
    port: 8125
  carbon:
    address: carbon.example.net
    port: 2003
transforms:
  statsd:
    # remove the hostname portion of collectd metrics with regexp replace
    - match: ^collectd\.(.*?)\.(.*)
      replace: collectd.$2
    - match: ^foo
      replace: bar
```

### Liveness checks

`csproxy` provides an HTTP listener for liveness checks at `/_ping`

```
$ http http://localhost:9080/_ping
HTTP/1.1 200 OK
Content-Length: 4
Content-Type: text/plain; charset=utf-8
Date: Sun, 15 May 2016 14:14:41 GMT

PONG
```

A non-200 or missing PONG indicates the service is unavailable.

### Building

```
# install deps
go get

# build it
go build

# run it (after writing your config!)
./csproxy
```
