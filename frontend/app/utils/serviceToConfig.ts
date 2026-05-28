import type { Service, ServiceConfig } from '~/types/wails'

// stdoutPath / stderrPath are intentionally omitted: CreateServiceModal always
// re-derives log paths from `composeLogPaths(serviceType, settings, label)` at
// submit time, so passing the source service's log paths would be discarded
// and misleading.
export function serviceToConfig(svc: Service): ServiceConfig {
  return {
    label: svc.label,
    program: svc.program,
    arguments: svc.arguments ? [...svc.arguments] : [],
    runAtLoad: svc.runAtLoad,
    keepAlive: svc.keepAlive,
    wakeSystem: svc.wakeSystem,
    schedule: svc.schedule ? { ...svc.schedule } : undefined,
    environment: svc.environment ? { ...svc.environment } : undefined,
    workingDirectory: svc.workingDirectory,
  }
}
