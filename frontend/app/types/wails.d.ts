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
  wakeSystem: boolean
  schedule?: ScheduleConfig
  environment?: Record<string, string>
  stdoutPath?: string
  stderrPath?: string
  workingDirectory?: string
  type: 'user' | 'system' | 'apple-system'
  readOnly: boolean
  plistFormat: 'xml' | 'binary' | 'unknown'
  statusConfidence: 'verified' | 'unverified'
}

export interface CalendarEntry {
  minute?: number
  hour?: number
  day?: number
  weekday?: number
  month?: number
}

export interface ScheduleConfig {
  schedules?: CalendarEntry[]
  interval?: number
}

export interface PlistContent {
  data: string
  format: 'xml' | 'binary' | 'unknown' | ''
  convertFailed: boolean
}

export interface LogClearStatus {
  logPath: string
  exists: boolean
  userWritable: boolean
}

export interface Backup {
  id: string
  service: string
  timestamp: string
  path: string
  originalPath?: string
}

export interface ServiceConfig {
  label: string
  program?: string
  arguments?: string[]
  runAtLoad: boolean
  keepAlive: boolean
  wakeSystem: boolean
  schedule?: ScheduleConfig
  environment?: Record<string, string>
  stdoutPath?: string
  stderrPath?: string
  workingDirectory?: string
}

export type AdminModeState = 'disabled' | 'requesting' | 'enabled' | 'shutting_down'

export interface AdminModeStatus {
  state: AdminModeState
  error: string | null
}

export interface Settings {
  userLogDir: string
  systemLogDir: string
}

export interface DeleteServiceOptions {
  deleteLogs: boolean
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
          KickstartService(name: string): Promise<void>
          GetPlist(name: string): Promise<string>
          GetLogs(name: string, logType: string): Promise<string>
          ClearLogs(name: string, logType: string): Promise<void>
          ClearSystemLogs(name: string, serviceType: string, logType: string): Promise<void>
          GetLogClearStatus(name: string, serviceType: string, logType: string): Promise<LogClearStatus>
          CreateService(config: ServiceConfig): Promise<void>
          UpdateService(name: string, config: ServiceConfig): Promise<void>
          DeleteService(name: string): Promise<void>
          ListSystemServices(): Promise<Service[]>
          ListAppleSystemServices(): Promise<Service[]>
          GetSystemService(name: string, serviceType: string): Promise<Service>
          GetSystemPlist(name: string, serviceType: string): Promise<string>
          GetSystemLogs(name: string, serviceType: string, logType: string): Promise<string>
          ListAllBackups(): Promise<Backup[]>
          ListBackups(serviceName: string): Promise<Backup[]>
          GetBackupContent(serviceName: string, backupID: string): Promise<PlistContent>
          GetCurrentPlist(name: string): Promise<PlistContent>
          GetCurrentSystemPlist(name: string): Promise<PlistContent>
          RestoreBackup(serviceName: string, backupID: string): Promise<void>
          GetVersion(): Promise<string>
          CheckPermissions(): Promise<Record<string, boolean>>
          EnableAdminMode(): Promise<void>
          DisableAdminMode(): Promise<void>
          GetAdminModeStatus(): Promise<AdminModeStatus>
          StartSystemService(name: string): Promise<void>
          StopSystemService(name: string): Promise<void>
          RestartSystemService(name: string): Promise<void>
          CreateSystemService(config: ServiceConfig): Promise<void>
          UpdateSystemService(name: string, config: ServiceConfig): Promise<void>
          DeleteSystemService(name: string, options: DeleteServiceOptions): Promise<string>
          GetSettings(): Promise<Settings>
          UpdateSettings(s: Settings): Promise<void>
        }
      }
    }
  }
}

export {}
