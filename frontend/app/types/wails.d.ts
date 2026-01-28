export interface Service {
  name: string
  label: string
  status: 'running' | 'stopped' | 'unknown' | 'loaded'
  pid?: number
  path: string
  program?: string
  arguments?: string[]
  runAtLoad: boolean
  keepAlive: boolean
  schedule?: ScheduleConfig
  environment?: Record<string, string>
  stdoutPath?: string
  stderrPath?: string
  workingDirectory?: string
  type: 'user' | 'system' | 'apple-system'
  readOnly: boolean
}

export interface ScheduleConfig {
  minute?: number
  hour?: number
  day?: number
  weekday?: number
  month?: number
}

export interface ServiceConfig {
  label: string
  program?: string
  arguments?: string[]
  runAtLoad: boolean
  keepAlive: boolean
  schedule?: ScheduleConfig
  environment?: Record<string, string>
  stdoutPath?: string
  stderrPath?: string
  workingDirectory?: string
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          ListServices(): Promise<Service[]>
          GetService(name: string): Promise<Service>
          StartService(name: string): Promise<void>
          StopService(name: string): Promise<void>
          RestartService(name: string): Promise<void>
          GetPlist(name: string): Promise<string>
          GetLogs(name: string, logType: string): Promise<string>
          CreateService(config: ServiceConfig): Promise<void>
          UpdateService(name: string, config: ServiceConfig): Promise<void>
          DeleteService(name: string): Promise<void>
          ListSystemServices(): Promise<Service[]>
          ListAppleSystemServices(): Promise<Service[]>
          GetSystemService(name: string, serviceType: string): Promise<Service>
          GetSystemPlist(name: string, serviceType: string): Promise<string>
          GetSystemLogs(name: string, serviceType: string, logType: string): Promise<string>
          CheckPermissions(): Promise<Record<string, boolean>>
        }
      }
    }
  }
}

export {}
