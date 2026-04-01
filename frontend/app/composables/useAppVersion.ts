const appVersion = ref('dev')
let fetched = false

export function useAppVersion() {
  if (!fetched) {
    fetched = true
    window.go?.main?.App?.GetVersion?.().then((v) => {
      if (v) appVersion.value = v
    }).catch(() => {})
  }
  return appVersion
}
