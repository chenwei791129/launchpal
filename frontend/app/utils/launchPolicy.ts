import type { KeepAliveConfig } from '~/types/wails'

// LaunchPolicy is the single mutually-exclusive launch-behavior selection
// shared by the create modal and the edit form. It replaces the former pair of
// independent RunAtLoad / KeepAlive checkboxes.
export type LaunchPolicy = 'onDemand' | 'runAtLoad' | 'keepAlive'

// emptyKeepAlive is the disabled ("no keep alive") state.
export function emptyKeepAlive(): KeepAliveConfig {
  return { enabled: false, mode: '' }
}

// cloneKeepAlive deep-copies a KeepAliveConfig, including its nested maps, so
// form edits never mutate the source service object held elsewhere in state.
export function cloneKeepAlive(ka: KeepAliveConfig | undefined): KeepAliveConfig {
  if (!ka) return emptyKeepAlive()
  return {
    enabled: ka.enabled,
    mode: ka.mode,
    successfulExit: ka.successfulExit,
    crashed: ka.crashed,
    afterInitialDemand: ka.afterInitialDemand,
    networkState: ka.networkState,
    pathState: ka.pathState ? { ...ka.pathState } : undefined,
    otherJobEnabled: ka.otherJobEnabled ? { ...ka.otherJobEnabled } : undefined,
  }
}

// deriveLaunchPolicy maps an existing service's runAtLoad + keepAlive onto the
// radio selection. KeepAlive takes precedence because launchd implies
// RunAtLoad from KeepAlive: a service carrying both lands on `keepAlive`.
export function deriveLaunchPolicy(src: { runAtLoad: boolean; keepAlive?: KeepAliveConfig }): LaunchPolicy {
  if (src.keepAlive?.enabled) return 'keepAlive'
  if (src.runAtLoad) return 'runAtLoad'
  return 'onDemand'
}

// cloneLaunchPolicy is the clone-specific variant: a Keep Alive source is
// preserved, but any other source defaults to On Demand rather than carrying
// RunAtLoad into the clone (matching the legacy "clone forces RunAtLoad off").
export function cloneLaunchPolicy(src: { runAtLoad: boolean; keepAlive?: KeepAliveConfig }): LaunchPolicy {
  return deriveLaunchPolicy(src) === 'keepAlive' ? 'keepAlive' : 'onDemand'
}

// hasEffectiveKeepAliveSubKey reports whether a dictionary-mode KeepAlive has
// any condition set — editable bool sub-keys or preserved non-editable
// networkState / pathState / otherJobEnabled.
export function hasEffectiveKeepAliveSubKey(ka: KeepAliveConfig): boolean {
  return (
    ka.successfulExit !== undefined ||
    ka.crashed !== undefined ||
    ka.afterInitialDemand !== undefined ||
    ka.networkState !== undefined ||
    (ka.pathState !== undefined && Object.keys(ka.pathState).length > 0) ||
    (ka.otherJobEnabled !== undefined && Object.keys(ka.otherJobEnabled).length > 0)
  )
}

// normalizeKeepAliveForSubmit downgrades a dictionary-mode KeepAlive with no
// effective sub-key to the boolean form, mirroring the backend empty-dict rule
// so the frontend submission and the written plist agree.
export function normalizeKeepAliveForSubmit(ka: KeepAliveConfig): KeepAliveConfig {
  if (ka.mode === 'dictionary' && !hasEffectiveKeepAliveSubKey(ka)) {
    return { enabled: true, mode: 'boolean' }
  }
  return ka
}

// applyLaunchPolicy maps the radio selection + edited KeepAlive into the
// runAtLoad / keepAlive fields of a ServiceConfig:
//   - onDemand   → runAtLoad false, keepAlive disabled
//   - runAtLoad  → runAtLoad true,  keepAlive disabled
//   - keepAlive  → runAtLoad false (launchd implies it), keepAlive enabled
// ThrottleInterval is an independent top-level field (not governed by the
// launch policy) and is round-tripped by the caller, so it is intentionally
// not handled here — clearing it here would drop a pre-existing value when a
// service that has ThrottleInterval but no KeepAlive is edited.
export function applyLaunchPolicy(
  policy: LaunchPolicy,
  keepAlive: KeepAliveConfig,
): { runAtLoad: boolean; keepAlive: KeepAliveConfig } {
  if (policy === 'runAtLoad') {
    return { runAtLoad: true, keepAlive: emptyKeepAlive() }
  }
  if (policy === 'keepAlive') {
    // Boolean (or unset) mode is a plain Keep Alive that launchd writes as
    // `true`; dictionary mode round-trips its sub-keys (empty dict downgrades).
    const ka = keepAlive.mode === 'dictionary'
      ? normalizeKeepAliveForSubmit({ ...keepAlive, enabled: true })
      : { enabled: true, mode: 'boolean' as const }
    return { runAtLoad: false, keepAlive: ka }
  }
  return { runAtLoad: false, keepAlive: emptyKeepAlive() }
}
