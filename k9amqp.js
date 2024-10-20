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
  k9amqp.init(amqpOptions, {channels_per_conn: 5, channel_cache_size: 1050})
  exchange.declare(k9amqp, {name: "test.ex", kind: "topic", durable: true})
  exchange.declare(k9amqp, {name: "rcvr.ex", kind: "topic", durable: true})
  exchange.bind(k9amqp, {destination: "rcvr.ex", key: "test", source: "test.ex"})
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

export default function() {
  k9amqp.publish(
  { exchange: "rcvr.ex", key: "test"},
  { content_type: "application/json",    
    priority: 0,
    app_id: "k6",
    body: "{\"test\":\"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz\"}"
  }
 )
 sleep(.05)
 let msg = k9amqp.get({queue: "test.q", auto_ack: true})
 if (msg) {
  // do smth.
  //console.log(msg)
 }
 
}
