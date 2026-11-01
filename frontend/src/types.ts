export interface SiteItem {
  name?: string
  domain: string
  path?: string
  root_path: string
  php_version?: string
  php_port?: number
  proxy_port?: number
  project_type?: string
  type_label?: string
  is_web?: boolean
}

export interface ServiceItem {
  id: string
  name_key: string
  desc_key: string
  port: number
  running: boolean
  url?: string
  credentials?: {
    host?: string
    user?: string
    password?: string
    database?: string
    smtp_host?: string
    smtp_port?: number
    web_port?: number
    info?: string
  }
}
