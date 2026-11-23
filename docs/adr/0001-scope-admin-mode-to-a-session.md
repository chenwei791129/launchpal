# Scope Admin Mode to a LaunchPal session

LaunchPal is unsigned, so it cannot register a persistent privileged helper through `SMAppService`. Prompting on every operation would interrupt routine service management, while running the whole app as root or installing a persistent unsigned daemon would grant more privilege for longer than needed. Keep the GUI unprivileged and start a separate, narrowly scoped helper only after the user enables Admin Mode. The helper process ends with the session; Apple System Services remain read-only.

The helper executable can persist even though the root process does not. Prefer a verified root-owned copy over the user-writable app bundle. The build-time bundle hash is defense in depth for initial provisioning and updates, not the trust anchor: those paths still have to start from the bundle, so they cannot provide the same assurance as the protected copy.
