export interface Service {
  name: string
  label: string
  status: 'running' | 'stopped' | 'unknown'
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
        }
      }
    }
  }
}

export {}
