[![GitHub release](https://img.shields.io/github/v/release/mvolejnik/xk6-k9-amqp?style=for-the-badge)](https://github.com/mvolejnik/xk6-k9-amqp/releases)
[![GitHub CI Workflow](https://img.shields.io/github/actions/workflow/status/mvolejnik/xk6-k9-amqp/build.yaml?branch=master&style=for-the-badge)](https://github.com/mvolejnik/xk6-k9-amqp/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvolejnik/xk6-k9-amqp?style=for-the-badge)](https://goreportcard.com/report/github.com/mvolejnik/xk6-k9-amqp)
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
- causes [Hiigh Channel Churn](https://www.rabbitmq.com/docs/channels#monitoring) rate due to openning new channel for every message being published
- no connection pool support (connection per VU is possible, but even so it might be too many in complex scenarios)

## Links

How to send metrics to Prometheus see [K6 K9-AMQP extension prometheus remote write](Prometheus.md)

## K6 version compatibility

Extension [build tested](https://github.com/mvolejnik/xk6-k9-amqp/actions?query=branch%3Amaster) with versions (but might build with other versions with no issue)

- v1.1.0
- v1.0.0
- v0.59.0


## K6 Extension Registry Requirements

- ✅ a [README](./README.md) file - project description, build and usage instructions, as well as k6 version compatibility. The goal is to provide enough information to quickly and easily evaluate the extension.
- ❌ the xk6 topic set (not ready for this yet)
- ✅ a [non-restrictive](./LICENSE) license
- ✅ an [examples](./examples/) folder with examples
- ✅ at least one versioned [release](https://github.com/mvolejnik/xk6-k9-amqp/releases)

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

```
$ xk6 build latest --with github.com/mvolejnik/xk6-k9-amqp@v0.0.6

2025/06/22 10:52:49 INFO Building k6
2025/06/22 10:52:49 INFO Building new k6 binary (native)
2025/06/22 10:52:49 INFO Initializing Go module
go: creating new go.mod: module k6
2025/06/22 10:52:49 INFO Creating k6 main
2025/06/22 10:52:49 INFO adding dependency go.k6.io/k6@latest
go: finding module for package github.com/fsnotify/fsnotify
go: finding module for package gopkg.in/tomb.v1
go: found gopkg.in/tomb.v1 in gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
go: found github.com/fsnotify/fsnotify in github.com/fsnotify/fsnotify v1.9.0
2025/06/22 10:52:49 INFO importing extensions
2025/06/22 10:52:49 INFO adding dependency github.com/mvolejnik/xk6-k9-amqp@v0.0.6
go: downloading github.com/mvolejnik/xk6-k9-amqp v0.0.6
2025/06/22 10:53:08 INFO Building k6
2025/06/22 10:53:15 INFO Build complete
2025/06/22 10:53:15 INFO Cleaning up work directory /tmp/k6foundry1730656898

xk6 has now produced a new k6 binary which may be different than the command on your system path!
Be sure to run './k6 run <SCRIPT_NAME>' from the '/home/mvolejnik/tmp/k6' directory.
```

Verify:
```
$ ./k6 --version
k6 v1.0.0 (go1.24.3, linux/amd64)
Extensions:
  github.com/mvolejnik/xk6-k9-amqp v0.0.6, k6/x/k9amqp [js]
  github.com/mvolejnik/xk6-k9-amqp v0.0.6, k6/x/k9amqp/exchange [js]
  github.com/mvolejnik/xk6-k9-amqp v0.0.6, k6/x/k9amqp/queue [js]
```


### Build K6 with K9 AMQP extension locally
`go build .` used in development as fast compilation for syntax errors as it's way faster than `xk6 build...`.

```
$ cat build.sh 
#!/usr/bin/env bash

go build . && xk6 build latest --with xk6-k9-amqp=.

$ ./build.sh 
2024/11/23 21:38:06 [INFO] Temporary folder: /tmp/buildenv_2024-11-23-2138.79571920
2024/11/23 21:38:06 [INFO] Initializing Go module
2024/11/23 21:38:06 [INFO] exec (timeout=10s): /opt/go/bin/go mod init k6 
go: creating new go.mod: module k6
2024/11/23 21:38:06 [INFO] Replace xk6-k9-amqp => /home/mvolejnik/Git/xk6-k9-amqp
2024/11/23 21:38:06 [INFO] exec (timeout=0s): /opt/go/bin/go mod edit -replace xk6-k9-amqp=/home/mvolejnik/Git/xk6-k9-amqp 
2024/11/23 21:38:06 [INFO] exec (timeout=0s): /opt/go/bin/go mod tidy -compat=1.17 
go: warning: "all" matched no packages
2024/11/23 21:38:06 [INFO] Pinning versions
2024/11/23 21:38:06 [INFO] exec (timeout=0s): /opt/go/bin/go mod tidy -compat=1.17 
go: found xk6-k9-amqp in xk6-k9-amqp v0.0.0-00010101000000-000000000000
go: finding module for package github.com/nxadm/tail
go: found github.com/nxadm/tail in github.com/nxadm/tail v1.4.11
2024/11/23 21:38:06 [INFO] Writing main module: /tmp/buildenv_2024-11-23-2138.79571920/main.go
2024/11/23 21:38:06 [INFO] exec (timeout=0s): /opt/go/bin/go mod edit -require go.k6.io/k6@latest 
2024/11/23 21:38:06 [INFO] exec (timeout=0s): /opt/go/bin/go mod tidy -compat=1.17 
2024/11/23 21:38:07 [INFO] exec (timeout=0s): /opt/go/bin/go mod tidy -compat=1.17 
2024/11/23 21:38:07 [INFO] Build environment ready
2024/11/23 21:38:07 [INFO] Building k6
2024/11/23 21:38:07 [INFO] exec (timeout=0s): /opt/go/bin/go mod tidy -compat=1.17 
2024/11/23 21:38:07 [INFO] exec (timeout=0s): /opt/go/bin/go build -o /home/mvolejnik/Git/xk6-k9-amqp/k6 -ldflags=-w -s -trimpath 
2024/11/23 21:38:08 [INFO] Build complete: ./k6
2024/11/23 21:38:08 [INFO] Cleaning up temporary folder: /tmp/buildenv_2024-11-23-2138.79571920

xk6 has now produced a new k6 binary which may be different than the command on your system path!
Be sure to run './k6 run <SCRIPT_NAME>' from the '/home/mvolejnik/Git/xk6-k9-amqp' directory.

```

### Run RqbbitMQ

```
sudo docker run -p 5672:5672 -p 15672:15672 rabbitmq:3-management

Unable to find image 'rabbitmq:3-management' locally
3-management: Pulling from library/rabbitmq
0622fac788ed: Already exists 
6926ed160d39: Pull complete 
6c42ae7d1a2d: Pull complete 
261d4b178df4: Pull complete 
6fd08a40d273: Pull complete 
199259066d28: Pull complete 
fed2c5809243: Pull complete 
953b81faf3db: Pull complete 
f775e99f7f42: Pull complete 
c3abc98a719c: Pull complete 
Digest: sha256:9d984edac52ffeea602dd20e28b752394597ad214dcce5f66c1bd45700c8f296
Status: Downloaded newer image for rabbitmq:3-management
=INFO REPORT==== 24-May-2025::08:20:31.632035 ===
    alarm_handler: {set,{system_memory_high_watermark,[]}}
2025-05-24 08:20:33.853573+00:00 [notice] <0.44.0> Application syslog exited with reason: stopped
2025-05-24 08:20:33.857921+00:00 [notice] <0.254.0> Logging: switching to configured handler(s); following messages may not be visible in this log output
2025-05-24 08:20:33.858542+00:00 [notice] <0.254.0> Logging: configured log handlers are now ACTIVE
...
2025-05-24 08:20:38.761939+00:00 [info] <0.754.0> Management plugin: HTTP (non-TLS) listener started on port 15672
2025-05-24 08:20:38.762206+00:00 [info] <0.784.0> Statistics database started.
2025-05-24 08:20:38.762350+00:00 [info] <0.783.0> Starting worker pool 'management_worker_pool' with 3 processes in it
2025-05-24 08:20:38.775540+00:00 [info] <0.802.0> Prometheus metrics: HTTP (non-TLS) listener started on port 15692
2025-05-24 08:20:38.775714+00:00 [info] <0.688.0> Ready to start client connection listeners
2025-05-24 08:20:38.777524+00:00 [info] <0.846.0> started TCP listener on [::]:5672
 completed with 5 plugins.
2025-05-24 08:20:38.854097+00:00 [info] <0.688.0> Server startup complete; 5 plugins started.
2025-05-24 08:20:38.854097+00:00 [info] <0.688.0>  * rabbitmq_prometheus
2025-05-24 08:20:38.854097+00:00 [info] <0.688.0>  * rabbitmq_federation
2025-05-24 08:20:38.854097+00:00 [info] <0.688.0>  * rabbitmq_management
2025-05-24 08:20:38.854097+00:00 [info] <0.688.0>  * rabbitmq_management_agent
2025-05-24 08:20:38.854097+00:00 [info] <0.688.0>  * rabbitmq_web_dispatch
2025-05-24 08:20:38.966514+00:00 [info] <0.9.0> Time to start RabbitMQ: 7407 ms
2025-05-24 08:21:38.739619+00:00 [info] <0.900.0> Waiting for Mnesia tables for 30000 ms, 9 retries left
2025-05-24 08:21:38.739748+00:00 [info] <0.900.0> Successfully synced tables from a peer
```

#### Run Simple Sample


```
$ ./k6 run examples/simple.js 

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 

INFO[0000] 2024/11/23 21:41:45 INFO init amqp client with pool {ChannelsPerConn:2 ChannelsCacheSize:10} 
     execution: local
        script: examples/simple.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 10 looping VUs for 30s (gracefulStop: 30s)

INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO exchange created name=test.ex 
INFO[0000] 2024/11/23 21:41:45 INFO qeuue created name=test.q 
INFO[0000] 2024/11/23 21:41:45 INFO qeuue binded name=test.q key=test 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0000] 2024/11/23 21:41:45 INFO no available channel in pool, creating new one 
INFO[0030] 2024/11/23 21:42:15 INFO qeuue deleted name=test.q 
INFO[0030] 2024/11/23 21:42:15 INFO exchange deleted name=test.ex 
INFO[0030] 2024/11/23 21:42:15 INFO Teardown AMQP Client 

     amqp_pub_failed........: 0    0/s
     amqp_pub_sent..........: 5750 191.158496/s
     amqp_sub_failed........: 0    0/s
     amqp_sub_no_delivery...: 5750 191.158496/s
     amqp_sub_received......: 5750 191.158496/s
     data_received..........: 0 B  0 B/s
     data_sent..............: 0 B  0 B/s
     iteration_duration.....: avg=52.23ms min=50.72ms med=52.06ms max=63.14ms p(90)=53.18ms p(95)=53.7ms
     iterations.............: 5750 191.158496/s
     vus....................: 10   min=10       max=10
     vus_max................: 10   min=10       max=10


running (0m30.1s), 00/10 VUs, 5750 complete and 0 interrupted iterations
default ✓ [======================================] 10 VUs  30s

```

### Consumer Listener Sample

The sample uses `constant-vus` executor and listener to consume messages from the queue.

```
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


```
$ ./k6 run ./examples/produce-consume-listener.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /  
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/

INFO[0000] 2025/07/08 21:43:07 INFO init amqp client with pool {ChannelsPerConn:1 ChannelsCacheSize:20}
     execution: local
        script: ./examples/produce-consume-listener.js
        output: -

     scenarios: (100.00%) 2 scenarios, 42 max VUs, 1m40s max duration (incl. graceful stop):
              * consume: 2 looping VUs for 1m10s (exec: consume, gracefulStop: 30s)
              * publish: 333.33 iterations/s for 1m0s (maxVUs: 20-40, exec: produce, gracefulStop: 30s)

INFO[0000] 2025/07/08 21:43:07 INFO no available channel in pool, creating new one
INFO[0000] 2025/07/08 21:43:07 INFO exchange created name=test.ex
INFO[0000] 2025/07/08 21:43:07 INFO qeuue created name=test.q
INFO[0000] 2025/07/08 21:43:07 INFO qeuue binded name=test.q key=test
INFO[0001] 2025/07/08 21:43:08 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0005] 2025/07/08 21:43:12 INFO no available channel in pool, creating new one
INFO[0070] Consumer [2] consumed 9995 messages           source=console
INFO[0070] Consumer [4] consumed 10006 messages          source=console
INFO[0070] 2025/07/08 21:44:17 INFO Teardown AMQP Client


  █ TOTAL RESULTS

    CUSTOM
    amqp_pub_sent..........................: 20001 285.395566/s
    amqp_sub_latency.......................: avg=0      min=0       med=0       max=0     p(90)=0        p(95)=0

    EXECUTION
    iteration_duration.....................: avg=7.34ms min=70.83µs med=322.9µs max=1m10s p(90)=412.21µs p(95)=450.42µs
    iterations.............................: 20003 285.424105/s
    vus....................................: 2     min=2        max=3
    vus_max................................: 22    min=22       max=22

    NETWORK
    data_received..........................: 0 B   0 B/s
    data_sent..............................: 0 B   0 B/s




running (1m10.1s), 00/22 VUs, 20003 complete and 0 interrupted iterations
consume ✓ [======================================] 2 VUs      1m10s
publish ✓ [======================================] 00/20 VUs  1m0s  333.33 iters/s

```