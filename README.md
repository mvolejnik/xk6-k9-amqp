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

"K6 K9 AMQP" - it's a pun. K6 for Grafana K6 testing tool, K-9 for Canine/(police) dog. K6 and K-9, both makes world a better place to live in. Even not noticing what happen when a dog see a rabbit...

### Simple Example

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
  const client = new k9amqp.Client()
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

### Build K6 with K9 AMQP extension
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

### Run Simple Example


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