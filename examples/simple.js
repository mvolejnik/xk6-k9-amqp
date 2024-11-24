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

// Inits K9 AMQP Client
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
 let msg = client.get({queue: "test.q", auto_ack: true})
 msg = client.get({queue: "test.q", auto_ack: true})
 if (msg) {
  // do smth.
  //console.log(msg)
 }
 
}
