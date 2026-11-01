import type { ServiceItem } from '../types'

export const defaultServices: ServiceItem[] = [
  {
    id: 'caddy',
    name_key: 'service_caddy',
    desc_key: 'service_caddy_desc',
    port: 443,
    running: true,
    url: 'http://localhost:2019',
    credentials: {
      host: 'localhost',
      info: 'Admin: :2019 / SSL Root',
    },
  },
  {
    id: 'php-8.5',
    name_key: 'service_php85',
    desc_key: 'service_php85_desc',
    port: 9085,
    running: false,
    credentials: {
      host: '127.0.0.1',
      info: 'FastCGI / 8 Workers',
    },
  },
  {
    id: 'php-8.4',
    name_key: 'service_php84',
    desc_key: 'service_php84_desc',
    port: 9084,
    running: true,
    credentials: {
      host: '127.0.0.1',
      info: 'FastCGI / 8 Workers',
    },
  },
  {
    id: 'php-8.3',
    name_key: 'service_php83',
    desc_key: 'service_php83_desc',
    port: 9083,
    running: true,
    credentials: {
      host: '127.0.0.1',
      info: 'FastCGI / 8 Workers',
    },
  },
  {
    id: 'mysql',
    name_key: 'service_mysql',
    desc_key: 'service_mysql_desc',
    port: 3306,
    running: true,
    credentials: {
      host: '127.0.0.1',
      user: 'root',
      password: '',
      database: 'laravel',
    },
  },
  {
    id: 'redis',
    name_key: 'service_redis',
    desc_key: 'service_redis_desc',
    port: 6379,
    running: false,
    credentials: {
      host: '127.0.0.1',
      password: '',
    },
  },
  {
    id: 'mailpit',
    name_key: 'service_mailpit',
    desc_key: 'service_mailpit_desc',
    port: 1025,
    running: true,
    url: 'http://127.0.0.1:8025',
    credentials: {
      smtp_host: '127.0.0.1',
      smtp_port: 1025,
      web_port: 8025,
    },
  },
]
