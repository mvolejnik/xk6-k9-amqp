[![GitHub release](https://img.shields.io/github/v/release/mvolejnik/xk6-k9-amqp?style=for-the-badge)](https://github.com/mvolejnik/xk6-k9-amqp/releases)
[![GitHub CI Workflow](https://img.shields.io/github/actions/workflow/status/mvolejnik/xk6-k9-amqp/build.yaml?branch=master&style=for-the-badge)](https://github.com/mvolejnik/xk6-k9-amqp/actions/workflows/build.yaml)
[![GitHub License](https://img.shields.io/github/license/mvolejnik/xk6-k9-amqp?style=for-the-badge)](https://github.com/mvolejnik/xk6-k9-amqp/blob/master/LICENSE)

# xk6-k9-amqp

```
   / \__
  (    0\____
  /         Y
 /   (______/
/     /    U
```

## K6 K9-AMQP extension

[K6](https://k6.io/) extension for AMQP 0-9-1 / [RabbitMQ](https://www.rabbitmq.com/) uses RabbitMQ [GoLang client library](https://github.com/rabbitmq/amqp091-go).

"K6 K9 AMQP" - it's a pun. K6 for Grafana K6 testing tool, K-9 for Canine/(police) dog. K6 and K-9, both make world a better place to live in. Even not noticing what happen when a dog see a rabbit...

## Why another AMQP extension?

There is [Grafana xk6-amqp extensions](https://github.com/grafana/xk6-amqp), but...

- it's deprecated
- causes [High Channel Churn](https://www.rabbitmq.com/docs/channels#monitoring) rate due to opening new channel for every message being published
- no connection pool support (connection per VU is possible, but even so it might be too many in complex scenarios)

## Links

How to send metrics to Prometheus see [K6 K9-AMQP extension prometheus remote write](Prometheus.md)

## K6 version compatibility

Extension [build tested](https://github.com/mvolejnik/xk6-k9-amqp/actions?query=branch%3Amaster) with versions (but might build with other versions with no issue)

- v2.1.0
- v2.0.0

Plugin version v0.0.7 k6 support:

- v1.5.0
- v1.4.2
- v1.1.0
- v1.0.0
- v0.59.0


## K6 Extension Registry Requirements

- ✅ a [LICENSE](./LICENSE) file using an allowed license
- ✅ a [README](./README.md) that explains what the extension does and how to use it
- ✅ a valid [go.mod](./go.mod)
- ✅ runnable [examples](./examples/) in an examples folder
- ✅ at least one versioned [release](https://github.com/mvolejnik/xk6-k9-amqp/releases)
- ✅ a GitHub workflow with golangci-lint running on every pull request and every merge on the main branch
- ✅ the xk6 topic set

## Simple Sample

```javascript
import { sleep } from 'k6';
import k9amqp from 'k6/x/k9amqp';
import queue from 'k6/x/k9amqp/queue';
import exchange from 'k6/x/k9amqp/exchange';

export const options = {
  vus: 10,
  duration: '30s',
}

const amqpOptions = {
  host : __ENV.AMQP_HOST || "localhost",
  port : __ENV.AMQP_PORT || 5672,
  vhost : __ENV.AMQP_VHOST || "/",
  username : __ENV.AMQP_USERNAME || "guest",
  password : __ENV.AMQP_PASSWORD || "guest"
}

const poolOptions = {
  channels_per_conn : __ENV.AMQP_CHANNELS_PER_CONN || 2,
  channels_cache_size : __ENV.AMQP_CACHE_SIZE || 10,
}

// Inits K9 AMQP Client, Client is singleton object initialized only once for all VUs.
const client = new k9amqp.Client(amqpOptions, poolOptions)

export function setup() {
  // Gets already inited K9 AMQP Client
  const client = new k9amqp.Client(amqpOptions, poolOptions)
  // Creates 'test.ex' exchange, 'test.q' queue and binds the queue to the exchange using 'test' routing key
  exchange.declare(client, {name: "test.ex", kind: "topic", durable: true})
  queue.declare(client, {name: "test.q", durable: true, args: {"x-message-ttl": 300000}})
  queue.bind(client, {name: "test.q", key: "test", exchange: "test.ex"})
}

// Cleans up
export function teardown(data) {
  const client = new k9amqp.Client()
  queue.delete(client, {name: "test.q"})
  exchange.delete(client, {name: 'test.ex'})
  client.teardown()
}

export default function() {
  // Publishes simple JSON message to 'test.ex' exchange usnit 'test' routing key.
  client.publish(
  { exchange: "test.ex", key: "test"},
  { content_type: "application/json",
    priority: 0,
    app_id: "k6",
    body: "{\"test\":\"fe95QtmPWcZ1e6UA4SAzFE1lSNbO60hfkpr6D8RSfT7JEfAK7MC1rWX0noKq6XcLxuGnS6s95RJnqxakFAKxUkXIyS77GnXSygIQQmjrTvIiIXkIVGnXkCgbyrDUSM6V8FFuQGEswmh7si9qHQa3q6NauHmtgOfFOdvv6qj2nGs6UNSOLLlpOhjysF2sHVL5pitHRvZaP80a5Cj6R3nopkmh2ZCgoovWBpFZC1yGL2IBqxE40tzLMVORTapTm23PT1gwCvLdL7ykvHNLnhOH5hS5Bu1SuU0lcn4BjY1sRd6nDgzt1I9ys6pkXJf1K4SGZsd7UYGPjdqDC2cLEmSH6KTk0W2Na4vIxr1nkXJyUpvv9yXZLnuPa82SpjbVeJ6aNhlJuSJQReSUOeQGU3c7s0dlQnZ4miePKX3TXMNCDu1eOMKcypAD9aIFpguV32egOcJLX8HxCQ21Q41m8wcMumw0xxWorLHxMd6eZJIcrmsOJip8H0Lf\"}"
  }
 )
 sleep(.05)
 // Consumes message from 'test.q'
 let msg = client.get({queue: "test.q", auto_ack: true})
 if (msg) {
  // do smth.
  //console.log(msg)
 }
}
```

## Build K6 with K9 AMQP extension

```sh
xk6 build latest --with github.com/mvolejnik/xk6-k9-amqp@v0.1.3
```

```sh
3:54PM INF Building new k6 binary (native)
3:54PM INF Initializing Go module
3:54PM INF Creating k6 main
3:54PM INF adding dependency go.k6.io/k6/v2@v2.1.0
3:54PM INF importing extensions
3:54PM INF adding dependency github.com/mvolejnik/xk6-k9-amqp@v0.1.3
3:54PM INF Building k6
3:54PM INF Build complete
3:54PM INF Cleaning up work directory /tmp/k6foundry768369310
3:54PM INF Successful build platform=linux/amd64
3:54PM INF added module=go.k6.io/k6/v2 version=v2.1.0
3:54PM INF added module=github.com/mvolejnik/xk6-k9-amqp version=v0.1.3
3:54PM INF A new binary has been built based on k6 version=v2.1.0

xk6 has now produced a new k6 binary which may be different than the command on your system path!
Be sure to run './k6 run <SCRIPT_NAME>' from the '/home/mvolejnik/Git/xk6-k9-amqp' directory.
```

Verify:
```sh
./k6 version
```

```sh
k6 v2.1.0 (go1.26.3, linux/amd64)
Extensions:
  xk6-k9-amqp (devel), k6/x/k9amqp [js]
  xk6-k9-amqp (devel), k6/x/k9amqp/exchange [js]
  xk6-k9-amqp (devel), k6/x/k9amqp/queue [js]
```


### Build K6 with K9 AMQP extension locally
`go build .` used in development as fast compilation for syntax errors as it's way faster than `xk6 build...`.

```sh
$ cat build.sh 
#!/usr/bin/env bash

go build . && xk6 build latest --with xk6-k9-amqp=.
```

```sh
$ ./build.sh 
```

```sh
11:21AM INF Building new k6 binary (native)
11:21AM INF Initializing Go module
11:21AM INF Creating k6 main
11:21AM INF adding dependency go.k6.io/k6/v2@v2.1.0
11:21AM INF importing extensions
11:21AM INF adding dependency xk6-k9-amqp => .
11:21AM INF Building k6
11:21AM INF Build complete
11:21AM INF Cleaning up work directory /tmp/k6foundry1002671912
11:21AM INF Successful build platform=linux/amd64
11:21AM INF added module=go.k6.io/k6/v2 version=v2.1.0
11:21AM INF added module=xk6-k9-amqp version=v0.0.0-00010101000000-000000000000
11:21AM INF A new binary has been built based on k6 version=v2.1.0

xk6 has now produced a new k6 binary which may be different than the command on your system path!
Be sure to run './k6 run <SCRIPT_NAME>' from the '/home/mvolejnik/Git/xk6-k9-amqp' directory.
```

### Run RqbbitMQ

```sh
sudo docker run -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine
```

```sh
Unable to find image 'rabbitmq:3-management-alpine' locally
3-management-alpine: Pulling from library/rabbitmq
2d35ebdb57d9: Pull complete 
9ab46c560811: Pull complete 
7f4fe0a74327: Pull complete 
fe0979840307: Pull complete 
5541f4a1e9a3: Pull complete 
f7900ae6286e: Pull complete 
64a815d19254: Pull complete 
5a8ebb2679b2: Pull complete 
3fdf1da6c087: Pull complete 
3c428bdb4834: Pull complete 
63aa2505d58f: Download complete 
1fd50e0483ed: Download complete 
Digest: sha256:606d8c0d6b3c18d1da9afc53bc7cdb2a8d5486df91b5a9830e9e07626c9ae281
Status: Downloaded newer image for rabbitmq:3-management-alpine
2026-07-12 09:24:39.652310+00:00 [notice] <0.44.0> Application syslog exited with reason: stopped
2026-07-12 09:24:39.655537+00:00 [notice] <0.254.0> Logging: switching to configured handler(s); following messages may not be visible in this log output
2026-07-12 09:24:39.656090+00:00 [notice] <0.254.0> Logging: configured log handlers are now ACTIVE
2026-07-12 09:24:39.660940+00:00 [info] <0.254.0> ra: starting system quorum_queues
2026-07-12 09:24:39.661022+00:00 [info] <0.254.0> starting Ra system: quorum_queues in directory: /var/lib/rabbitmq/mnesia/rabbit@6acf2ba0b401/quorum/rabbit@6acf2ba0b401
2026-07-12 09:24:39.686031+00:00 [info] <0.268.0> ra system 'quorum_queues' running pre init for 0 registered servers
2026-07-12 09:24:39.690044+00:00 [info] <0.269.0> ra: meta data store initialised for system quorum_queues. 0 record(s) recovered
2026-07-12 09:24:39.696235+00:00 [notice] <0.274.0> WAL: ra_log_wal init, open tbls: ra_log_open_mem_tables, closed tbls: ra_log_closed_mem_tables
2026-07-12 09:24:39.707416+00:00 [info] <0.254.0> ra: starting system coordination
2026-07-12 09:24:39.707466+00:00 [info] <0.254.0> starting Ra system: coordination in directory: /var/lib/rabbitmq/mnesia/rabbit@6acf2ba0b401/coordination/rabbit@6acf2ba0b401
2026-07-12 09:24:39.708058+00:00 [info] <0.282.0> ra system 'coordination' running pre init for 0 registered servers
2026-07-12 09:24:39.708326+00:00 [info] <0.283.0> ra: meta data store initialised for system coordination. 0 record(s) recovered
2026-07-12 09:24:39.708416+00:00 [notice] <0.288.0> WAL: ra_coordination_log_wal init, open tbls: ra_coordination_log_open_mem_tables, closed tbls: ra_coordination_log_closed_mem_tables
2026-07-12 09:24:39.709690+00:00 [info] <0.254.0> ra: starting system coordination
2026-07-12 09:24:39.709713+00:00 [info] <0.254.0> starting Ra system: coordination in directory: /var/lib/rabbitmq/mnesia/rabbit@6acf2ba0b401/coordination/rabbit@6acf2ba0b401
2026-07-12 09:24:39.752807+00:00 [info] <0.254.0> Waiting for Khepri leader for 30000 ms, 9 retries left
2026-07-12 09:24:39.756983+00:00 [notice] <0.292.0> RabbitMQ metadata store: candidate -> leader in term: 1 machine version: 1
2026-07-12 09:24:39.760325+00:00 [info] <0.254.0> Khepri leader elected
2026-07-12 09:24:39.760385+00:00 [info] <0.254.0> Waiting for Khepri projections for 30000 ms, 9 retries left
2026-07-12 09:24:39.897774+00:00 [info] <0.254.0> 
2026-07-12 09:24:39.897774+00:00 [info] <0.254.0>  Starting RabbitMQ 3.13.7 on Erlang 26.2.5.16 [jit]
2026-07-12 09:24:39.897774+00:00 [info] <0.254.0>  Copyright (c) 2007-2024 Broadcom Inc and/or its subsidiaries
2026-07-12 09:24:39.897774+00:00 [info] <0.254.0>  Licensed under the MPL 2.0. Website: https://rabbitmq.com

  ##  ##      RabbitMQ 3.13.7
  ##  ##
  ##########  Copyright (c) 2007-2024 Broadcom Inc and/or its subsidiaries
  ######  ##
  ##########  Licensed under the MPL 2.0. Website: https://rabbitmq.com

  Erlang:      26.2.5.16 [jit]
  TLS Library: OpenSSL - OpenSSL 3.1.8 11 Feb 2025
  Release series support status: see https://www.rabbitmq.com/release-information

  Doc guides:  https://www.rabbitmq.com/docs
  Support:     https://www.rabbitmq.com/docs/contact
  Tutorials:   https://www.rabbitmq.com/tutorials
  Monitoring:  https://www.rabbitmq.com/docs/monitoring
  Upgrading:   https://www.rabbitmq.com/docs/upgrade

  Logs: <stdout>

  Config file(s): /etc/rabbitmq/conf.d/10-defaults.conf

  Starting broker...2026-07-12 09:24:39.898319+00:00 [info] <0.254.0> 
2026-07-12 09:24:39.898319+00:00 [info] <0.254.0>  node           : rabbit@6acf2ba0b401
2026-07-12 09:24:39.898319+00:00 [info] <0.254.0>  home dir       : /var/lib/rabbitmq
2026-07-12 09:24:39.898319+00:00 [info] <0.254.0>  config file(s) : /etc/rabbitmq/conf.d/10-defaults.conf
2026-07-12 09:24:39.898319+00:00 [info] <0.254.0>  cookie hash    : Msc4Kn/v2NH8H+MCUO2R0w==
2026-07-12 09:24:39.898319+00:00 [info] <0.254.0>  log(s)         : <stdout>
2026-07-12 09:24:39.898319+00:00 [info] <0.254.0>  data dir       : /var/lib/rabbitmq/mnesia/rabbit@6acf2ba0b401
2026-07-12 09:24:40.109907+00:00 [info] <0.254.0> Running boot step pre_boot defined by app rabbit
2026-07-12 09:24:40.109963+00:00 [info] <0.254.0> Running boot step rabbit_global_counters defined by app rabbit
2026-07-12 09:24:40.110191+00:00 [info] <0.254.0> Running boot step rabbit_osiris_metrics defined by app rabbit
2026-07-12 09:24:40.110296+00:00 [info] <0.254.0> Running boot step rabbit_core_metrics defined by app rabbit
2026-07-12 09:24:40.110881+00:00 [info] <0.254.0> Running boot step rabbit_alarm defined by app rabbit
2026-07-12 09:24:40.116698+00:00 [info] <0.329.0> Memory high watermark set to 12343 MiB (12942662041 bytes) of 30857 MiB (32356655104 bytes) total
2026-07-12 09:24:40.118275+00:00 [info] <0.331.0> Enabling free disk space monitoring (disk free space: 830262673408, total memory: 32356655104)
2026-07-12 09:24:40.118303+00:00 [info] <0.331.0> Disk free limit set to 50MB
2026-07-12 09:24:40.118874+00:00 [info] <0.254.0> Running boot step code_server_cache defined by app rabbit
2026-07-12 09:24:40.118918+00:00 [info] <0.254.0> Running boot step file_handle_cache defined by app rabbit
2026-07-12 09:24:40.122481+00:00 [info] <0.334.0> Limiting to approx 927 file handles (832 sockets)
2026-07-12 09:24:40.122586+00:00 [info] <0.335.0> FHC read buffering: OFF
2026-07-12 09:24:40.122615+00:00 [info] <0.335.0> FHC write buffering: ON
2026-07-12 09:24:40.122744+00:00 [info] <0.254.0> Running boot step worker_pool defined by app rabbit
2026-07-12 09:24:40.122769+00:00 [info] <0.315.0> Will use 16 processes for default worker pool
2026-07-12 09:24:40.122788+00:00 [info] <0.315.0> Starting worker pool 'worker_pool' with 16 processes in it
2026-07-12 09:24:40.123174+00:00 [info] <0.254.0> Running boot step database defined by app rabbit
2026-07-12 09:24:40.123305+00:00 [info] <0.254.0> Peer discovery: configured backend: rabbit_peer_discovery_classic_config
.....
2026-07-12 09:24:41.644366+00:00 [info] <0.770.0> Management plugin: HTTP (non-TLS) listener started on port 15672
2026-07-12 09:24:41.644476+00:00 [info] <0.800.0> Statistics database started.
2026-07-12 09:24:41.644523+00:00 [info] <0.799.0> Starting worker pool 'management_worker_pool' with 3 processes in it
2026-07-12 09:24:41.650297+00:00 [info] <0.818.0> Prometheus metrics: HTTP (non-TLS) listener started on port 15692
2026-07-12 09:24:41.650389+00:00 [info] <0.704.0> Ready to start client connection listeners
2026-07-12 09:24:41.651291+00:00 [info] <0.862.0> started TCP listener on [::]:5672
 completed with 5 plugins.
2026-07-12 09:24:41.679545+00:00 [info] <0.704.0> Server startup complete; 5 plugins started.
2026-07-12 09:24:41.679545+00:00 [info] <0.704.0>  * rabbitmq_prometheus
2026-07-12 09:24:41.679545+00:00 [info] <0.704.0>  * rabbitmq_federation
2026-07-12 09:24:41.679545+00:00 [info] <0.704.0>  * rabbitmq_management
2026-07-12 09:24:41.679545+00:00 [info] <0.704.0>  * rabbitmq_management_agent
2026-07-12 09:24:41.679545+00:00 [info] <0.704.0>  * rabbitmq_web_dispatch
2026-07-12 09:24:41.875025+00:00 [info] <0.9.0> Time to start RabbitMQ: 3559 ms
```

#### Run Simple Sample


```sh
./k6 run examples/simple.js
```

```sh
INFO[0000] 2026/07/12 11:27:37 INFO init amqp client with pool {ChannelsPerConn:2 ChannelsCacheSize:10}

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  / 
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: examples/simple.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 10 looping VUs for 30s (gracefulStop: 30s)

INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO exchange created name=test.ex
INFO[0000] 2026/07/12 11:27:37 INFO queue created name=test.q
INFO[0000] 2026/07/12 11:27:37 INFO queue binded name=test.q key=test
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:27:37 INFO no available channel in pool, creating new one
INFO[0030] 2026/07/12 11:28:07 INFO queue deleted name=test.q
INFO[0030] 2026/07/12 11:28:07 INFO exchange deleted name=test.ex
INFO[0030] 2026/07/12 11:28:07 INFO Teardown AMQP Client


  █ TOTAL RESULTS

    CUSTOM
    amqp_pub_failed........: 0    0/s
    amqp_pub_sent..........: 5750 191.067308/s
    amqp_sub_failed........: 0    0/s
    amqp_sub_no_delivery...: 5750 191.067308/s
    amqp_sub_received......: 5750 191.067308/s

    EXECUTION
    iteration_duration.....: avg=52.21ms min=50.41ms med=52.17ms max=56.13ms p(90)=52.71ms p(95)=52.86ms
    iterations.............: 5750 191.067308/s
    vus....................: 10   min=10       max=10
    vus_max................: 10   min=10       max=10

    NETWORK
    data_received..........: 0 B  0 B/s
    data_sent..............: 0 B  0 B/s




running (0m30.1s), 00/10 VUs, 5750 complete and 0 interrupted iterations
default ✓ [======================================] 10 VUs  30s
```

### Consumer Listener Sample

The sample uses `constant-vus` executor and listener to consume messages from the queue.

```javascript
import k9amqp from 'k6/x/k9amqp';
import queue from 'k6/x/k9amqp/queue';
import exchange from 'k6/x/k9amqp/exchange';
import { vu } from 'k6/execution';
import { fail, sleep } from 'k6';

export const options = {
  scenarios: {
    publish: {
      executor: 'constant-arrival-rate',
      duration: '1m',
      rate: 20000,
      timeUnit: '1m',
      preAllocatedVUs: 20,
      maxVUs: 40,
      exec: 'produce'
    },
    consume: {
      executor: 'constant-vus',
      vus: 2,
      duration: '1m10s',
      exec: 'consume'
    },
  },
}

const consumeDuration = options.scenarios.consume.duration

const amqpOptions = {
  host : __ENV.AMQP_HOST || "localhost",
  port : __ENV.AMQP_PORT || 5672,
  vhost : __ENV.AMQP_VHOST || "/",
  username : __ENV.AMQP_USERNAME || "guest",
  password : __ENV.AMQP_PASSWORD || "guest"
}

const poolOptions = {
  channels_per_conn : __ENV.AMQP_CHANNELS_PER_CONN || 1,
  channels_cache_size : __ENV.AMQP_CACHE_SIZE || 20,
}

// Inits K9 AMQP Client
const client = new k9amqp.Client(amqpOptions, poolOptions)

const unitMap = {
  h: 3600,
  m: 60,
  s: 1,
};

export function setup() {
  // Gets already inited K9 AMQP Client
  const client = new k9amqp.Client(amqpOptions, poolOptions)
  // Creates 'test.ex' exchange, 'test.q' queue and binds the queue to the exchange using 'test' routing key
  exchange.declare(client, {name: "test.ex", kind: "topic", durable: true})
  queue.declare(client, {name: "test.q", durable: true, args: {"x-message-ttl": 3600000}})
  queue.bind(client, {name: "test.q", key: "test", exchange: "test.ex"})
}

// Cleans up
export function teardown(data) {
  const client = new k9amqp.Client()
  //queue.delete(client, {name: "test.q"})
  //exchange.delete(client, {name: 'test.ex'})
  client.teardown()
}

export default function() {
  fail('no exec defined in scenario');
}

export function produce() {
  // Publishes simple JSON message to 'test.ex' exchange usnit 'test' routing key.
  client.publish(
  { exchange: "test.ex", key: "test"},
  { content_type: "application/json",
    priority: 0,
    app_id: "k6",
    body: "{\"test\":\"fe95QtmPWcZ1e6UA4SAzFE1lSNbO60hfkpr6D8RSfT7JEfAK7MC1rWX0noKq6XcLxuGnS6s95RJnqxakFAKxUkXIyS77GnXSygIQQmjrTvIiIXkIVGnXkCgbyrDUSM6V8FFuQGEswmh7si9qHQa3q6NauHmtgOfFOdvv6qj2nGs6UNSOLLlpOhjysF2sHVL5pitHRvZaP80a5Cj6R3nopkmh2ZCgoovWBpFZC1yGL2IBqxE40tzLMVORTapTm23PT1gwCvLdL7ykvHNLnhOH5hS5Bu1SuU0lcn4BjY1sRd6nDgzt1I9ys6pkXJf1K4SGZsd7UYGPjdqDC2cLEmSH6KTk0W2Na4vIxr1nkXJyUpvv9yXZLnuPa82SpjbVeJ6aNhlJuSJQReSUOeQGU3c7s0dlQnZ4miePKX3TXMNCDu1eOMKcypAD9aIFpguV32egOcJLX8HxCQ21Q41m8wcMumw0xxWorLHxMd6eZJIcrmsOJip8H0Lf\"}"
  }
 )
}

export function consume() {
  let c = 0;
  const listener = function(data) {
    c++;
    // do smth.
   }
  client.listen({queue: "test.q", auto_ack: true}, listener);
  sleep(toSeconds(consumeDuration));
  console.log(`Consumer [${vu.idInTest}] consumed ${c} messages`)
}

function toSeconds(duration) {
  if (!duration || typeof duration !== 'string') {
    return 0;
  }
  const regex = /(\d+)([hms])/g;
  let totalSeconds = 0;
  const matches = duration.matchAll(regex);
  for (const match of matches) {
    const value = parseInt(match[1], 10);
    const unit = match[2];
    totalSeconds += value * unitMap[unit];
  }
  return totalSeconds;
}
```

#### Run Consumer Listner Sample


```sh
./k6 run ./examples/produce-consume-listener.js
```

```sh
INFO[0000] 2026/07/12 11:29:16 INFO init amqp client with pool {ChannelsPerConn:1 ChannelsCacheSize:20} 

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /  
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: ./examples/produce-consume-listener.js
        output: -

     scenarios: (100.00%) 2 scenarios, 42 max VUs, 1m40s max duration (incl. graceful stop):
              * consume: 2 looping VUs for 1m10s (exec: consume, gracefulStop: 30s)
              * publish: 333.33 iterations/s for 1m0s (maxVUs: 20-40, exec: produce, gracefulStop: 30s)

INFO[0000] 2026/07/12 11:29:16 INFO no available channel in pool, creating new one
INFO[0000] 2026/07/12 11:29:17 INFO exchange created name=test.ex
INFO[0000] 2026/07/12 11:29:17 INFO queue created name=test.q
INFO[0000] 2026/07/12 11:29:17 INFO queue binded name=test.q key=test
INFO[0070] Consumer [4] consumed 9997 messages           source=console
INFO[0070] Consumer [12] consumed 10003 messages         source=console
INFO[0070] 2026/07/12 11:30:27 INFO Teardown AMQP Client


  █ TOTAL RESULTS

    CUSTOM
    amqp_pub_failed......: 0     0/s
    amqp_pub_sent........: 20001 285.386009/s

    EXECUTION
    iteration_duration...: avg=7.06ms min=15.14µs med=57.5µs max=1m10s p(90)=107.14µs p(95)=128.18µs
    iterations...........: 20003 285.414546/s
    vus..................: 2     min=2        max=3 
    vus_max..............: 22    min=22       max=22

    NETWORK
    data_received........: 0 B   0 B/s
    data_sent............: 0 B   0 B/s




running (1m10.1s), 00/22 VUs, 20003 complete and 0 interrupted iterations
consume ✓ [======================================] 2 VUs      1m10s
publish ✓ [======================================] 00/20 VUs  1m0s  333.33 iters/s

```
