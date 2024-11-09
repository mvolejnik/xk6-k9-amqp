import http from 'k6/http';
import { sleep } from 'k6';
import k9amqp from 'k6/x/k9amqp';
import queue from 'k6/x/k9amqp/queue';
import exchange from 'k6/x/k9amqp/exchange';

export const options = {
  vus: 1000,
  duration: '30s',
}

const amqpOptions = {
  host : __ENV.AMQP_HOST || "localhost",
  port : __ENV.AMQP_PORT || 5672,
  vhost : __ENV.AMQP_VHOST || "/",
  username : __ENV.AMQP_USERNAME || "guest",
  password : __ENV.AMQP_PASSWORD || "guest"
}

export function setup() {
  k9amqp.init(amqpOptions, {channels_per_conn: 1, channel_cache_size: 1})
  exchange.declare(k9amqp, {name: "test.ex", kind: "topic", durable: true})
  exchange.declare(k9amqp, {name: "rcvr.ex", kind: "topic", durable: true})
  exchange.bind(k9amqp, {destination: "test.ex", key: "test", source: "rcvr.ex"})
  queue.declare(k9amqp, {name: "test.q", durable: true, args: {"x-message-ttl": 300000}})
  queue.bind(k9amqp, {name: "test.q", key: "test", exchange: "test.ex"})
}

export function teardown(data) {
  let q = queue.declare(k9amqp, {name: 'test.q', passive: true})
  console.log('queue status: ', q)
  queue.delete(k9amqp, {name: "test.q"})
  exchange.delete(k9amqp, {name: 'test.ex'})
  exchange.delete(k9amqp, {name: 'rcvr.ex'})
  k9amqp.teardown()
}

{ // TODO Init just once and share between VUs
  k9amqp.init(amqpOptions, {channels_per_conn: 1, channel_cache_size: 1})
}

export default function() {
  k9amqp.publish(
  { exchange: "rcvr.ex", key: "test"},
  { content_type: "application/json",    
    priority: 0,
    app_id: "k6",
    body: "{\"test\":\"fe95QtmPWcZ1e6UA4SAzFE1lSNbO60hfkpr6D8RSfT7JEfAK7MC1rWX0noKq6XcLxuGnS6s95RJnqxakFAKxUkXIyS77GnXSygIQQmjrTvIiIXkIVGnXkCgbyrDUSM6V8FFuQGEswmh7si9qHQa3q6NauHmtgOfFOdvv6qj2nGs6UNSOLLlpOhjysF2sHVL5pitHRvZaP80a5Cj6R3nopkmh2ZCgoovWBpFZC1yGL2IBqxE40tzLMVORTapTm23PT1gwCvLdL7ykvHNLnhOH5hS5Bu1SuU0lcn4BjY1sRd6nDgzt1I9ys6pkXJf1K4SGZsd7UYGPjdqDC2cLEmSH6KTk0W2Na4vIxr1nkXJyUpvv9yXZLnuPa82SpjbVeJ6aNhlJuSJQReSUOeQGU3c7s0dlQnZ4miePKX3TXMNCDu1eOMKcypAD9aIFpguV32egOcJLX8HxCQ21Q41m8wcMumw0xxWorLHxMd6eZJIcrmsOJip8H0Lf\"}"
  }
 )
 sleep(.05)
 let msg = k9amqp.get({queue: "test.q", auto_ack: true})
 msg = k9amqp.get({queue: "test.q", auto_ack: true})
 if (msg) {
  // do smth.
  //console.log(msg)
 }
}
