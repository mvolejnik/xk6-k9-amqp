// index.d.ts

// 1. Standalone/Global Types (No 'export' at the root level)
interface Table { [key: string]: any; }
interface AmqpOptions { host?: string; port?: number; vhost?: string; username?: string; password?: string; }
interface PoolOptions { channels_per_conn?: number; channels_cache_size?: number; }
interface Publishing { content_type?: string; content_encoding?: string; delivery_mode?: number; priority?: number; correlation_id?: string; reply_to?: string; expiration?: string; message_id?: string; timestamp?: Date; type?: string; user_id?: string; app_id?: string; headers?: Table; body: string | ArrayBuffer; }
interface Delivery { Acknowledger: any; Headers: Table; ContentType: string; ContentEncoding: string; DeliveryMode: number; Priority: number; CorrelationId: string; ReplyTo: string; Expiration: string; MessageId: string; Timestamp: Date; Type: string; UserId: string; AppId: string; ConsumerTag: string; MessageCount: number; DeliveryTag: number; Redelivered: boolean; Exchange: string; RoutingKey: string; Body: string | ArrayBuffer; }
interface PublishOptions { exchange: string; key: string; mandatory?: boolean; immediate?: boolean; }
interface AmqpProduceResponse { Error: boolean; ErrorMessage: string; }
interface GetOptions { queue: string; auto_ack: boolean; }
interface AmqpGetResponse { Delivery: Delivery; Ok: boolean; Error: boolean; ErrorMessage: string; }
interface ConsumeOptions { queue: string; auto_ack: boolean; exclusive?: boolean; no_local?: boolean; no_wait?: boolean; args?: Table; size: number; }
interface AmqpConsumeResponse { Deliveries: Delivery[]; Ok: boolean; Error: boolean; ErrorMessage: string; }
interface ListenOptions { queue: string; auto_ack: boolean; exclusive?: boolean; no_local?: boolean; no_wait?: boolean; args?: Table; }
type ListenerType = (delivery: Delivery) => void | Error;
interface QueueDeclareOptions { name: string; durable?: boolean; auto_delete?: boolean; exclusive?: boolean; no_wait?: boolean; passive?: boolean; args?: Table; }
interface Queue { Name: string; Messages: number; Consumers: number; }
interface QueueDeleteOptions { name: string; if_unused?: boolean; if_empty?: boolean; no_wait?: boolean; }
interface QueueBindOptions { name: string; key: string; exchange: string; no_wait?: boolean; args?: Table; }
interface QueueUnbindOptions { name: string; key: string; exchange: string; args?: Table; }
interface QueuePurgeOptions { name: string; no_wait?: boolean; }
interface ExchangeDeclareOptions { name: string; kind: 'direct' | 'topic' | 'fanout' | 'headers' | string; durable?: boolean; auto_delete?: boolean; internal?: boolean; no_wait?: boolean; args?: Table; }
interface ExchangeDeleteOptions { name: string; if_unused?: boolean; no_wait?: boolean; }
interface ExchangeBindOptions { destination: string; source: string; key: string; no_wait?: boolean; args?: Table; }
interface ExchangeUnbindOptions { destination: string; source: string; key: string; args?: Table; }

// 2. Main Module
declare module 'k6/x/k9amqp' {
  export class Client {
    constructor(amqpOptions?: AmqpOptions, poolOptions?: PoolOptions);
    publish(opts: PublishOptions, msg: Publishing): AmqpProduceResponse;
    get(opts: GetOptions): AmqpGetResponse;
    consume(opts: ConsumeOptions): AmqpConsumeResponse;
    listen(opts: ListenOptions, listener: ListenerType): void;
    teardown(): void;
  }
}

// 3. Queue Module
declare module 'k6/x/k9amqp/queue' {
  import { Client } from 'k6/x/k9amqp';
  
  export function declare(client: Client, opts: QueueDeclareOptions): Queue;
  export function delete_(client: Client, opts: QueueDeleteOptions): void;
  export function bind(client: Client, opts: QueueBindOptions): void;
  export function unbind(client: Client, opts: QueueUnbindOptions): void;
  export function purge(client: Client, opts: QueuePurgeOptions): number;

  const queue: {
    declare: typeof declare;
    delete: typeof delete_;
    bind: typeof bind;
    unbind: typeof unbind;
    purge: typeof purge;
  };
  export default queue;
}

// 4. Exchange Module
declare module 'k6/x/k9amqp/exchange' {
  import { Client } from 'k6/x/k9amqp';

  export function declare(client: Client, opts: ExchangeDeclareOptions): void;
  export function delete_(client: Client, opts: ExchangeDeleteOptions): void;
  export function bind(client: Client, opts: ExchangeBindOptions): void;
  export function unbind(client: Client, opts: ExchangeUnbindOptions): void;

  const exchange: {
    declare: typeof declare;
    delete: typeof delete_;
    bind: typeof bind;
    unbind: typeof unbind;
  };
  export default exchange;
}
