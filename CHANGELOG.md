# Changelog

## [1.8.0](https://github.com/chenwei791129/launchpal/compare/v1.7.0...v1.8.0) (2026-04-03)


### Features

* add Reveal in Finder button to service summary ([086d987](https://github.com/chenwei791129/launchpal/commit/086d987be52ac446e0d5c390c19664cce373c7dd))
* add RevealInFinder backend method ([a90443c](https://github.com/chenwei791129/launchpal/commit/a90443c94d49d4a8fbdf5a106ce8fb2298afc5ba))
* add WakeSystem support for scheduled services ([c6d341a](https://github.com/chenwei791129/launchpal/commit/c6d341a1bd09cc1f7aa5edbd098e94ad3b940197))
* preview next run times for CalendarInterval schedules ([6b8c137](https://github.com/chenwei791129/launchpal/commit/6b8c137b33dae07a41e789354464bb6709766bc9))


### Performance Improvements

* **launchctl:** batch status query and tail log reading ([6a7e2c0](https://github.com/chenwei791129/launchpal/commit/6a7e2c0eb0074dc6afba9e26b18e7ba539b0e359))

## [1.7.0](https://github.com/chenwei791129/launchpal/compare/v1.6.0...v1.7.0) (2026-03-31)


### Features

* add build-time version injection via ldflags ([c23208d](https://github.com/chenwei791129/launchpal/commit/c23208de96e7b58fd8fef1b0f1f0e4371a62788c))

## [1.6.0](https://github.com/chenwei791129/launchpal/compare/v1.5.0...v1.6.0) (2026-03-31)


### Features

* add update-homebrew job to release workflow ([68a418a](https://github.com/chenwei791129/launchpal/commit/68a418adead082768cebf727f7b4936fca7ebd4d))

## [1.5.0](https://github.com/chenwei791129/launchpal/compare/v1.4.1...v1.5.0) (2026-03-31)


### Features

* add custom DMG background with Gatekeeper hint ([f8d9008](https://github.com/chenwei791129/launchpal/commit/f8d9008a1c29e07cf262d01957ef7b7b8f83d67e))

## [1.4.1](https://github.com/chenwei791129/launchpal/compare/v1.4.0...v1.4.1) (2026-03-31)


### Bug Fixes

* upgrade Go version to 1.24 for tool directive support ([aaa1450](https://github.com/chenwei791129/launchpal/commit/aaa1450bc3e0283d71a4e079eac400efc74349e5))

## [1.4.0](https://github.com/chenwei791129/launchpal/compare/v1.3.0...v1.4.0) (2026-03-31)


### Features

* package application as DMG with drag-to-Applications install ([0820b46](https://github.com/chenwei791129/launchpal/commit/0820b46d09cf50b017e16d5241101f44e215983f))

## [1.3.0](https://github.com/chenwei791129/launchpal/compare/v1.2.0...v1.3.0) (2026-03-29)


### Features

* add shell-args utility for quoted argument parsing ([9c4b1e5](https://github.com/chenwei791129/launchpal/commit/9c4b1e532c9f2a7633674c2a3c09023347ff0667))


### Bug Fixes

* use shell-like parsing for service arguments ([444b79a](https://github.com/chenwei791129/launchpal/commit/444b79aec01ab3d60b1ad6950cf7a38b4dfecaae))

## [1.2.0](https://github.com/chenwei791129/launchpal/compare/v1.1.0...v1.2.0) (2026-03-29)


### Features

* add environment variables configuration in create service modal ([8d24061](https://github.com/chenwei791129/launchpal/commit/8d24061a26e6ab63b98585498f6d9128d424fb58))
* add environment variables configuration in edit service page ([d61d891](https://github.com/chenwei791129/launchpal/commit/d61d891e81be5490f15f83422aaf2934057bd417))


### Bug Fixes

* add h-full to edit tab to enable scrollbar ([748f325](https://github.com/chenwei791129/launchpal/commit/748f32519172309833c0d44b1c63f9c7b5e34bcb))

## [1.1.0](https://github.com/chenwei791129/launchpal/compare/v1.0.0...v1.1.0) (2026-03-25)


### Features

* add schedule configuration UI for new services ([10a2f90](https://github.com/chenwei791129/launchpal/commit/10a2f90c66feec671505f4cda9d8b40af717230b))
* add schedule display and editing in service detail page ([451bf64](https://github.com/chenwei791129/launchpal/commit/451bf641430a5deaed20600f203e3beb55a72e48))
* add StartInterval support for scheduled services ([629e2b5](https://github.com/chenwei791129/launchpal/commit/629e2b5f3e82dde020002f28fd6c8a924920c790))
* add StartInterval validation in Create and Update ([4bf7f9a](https://github.com/chenwei791129/launchpal/commit/4bf7f9a91f3ede92134bf36b5b2d5c16905dd297))
* show hint when multiple StartCalendarInterval entries exist ([dbfecc0](https://github.com/chenwei791129/launchpal/commit/dbfecc0973f66e7a50e3a87b780d66e69236ab35))


### Bug Fixes

* clarify cron syntax limitations in schedule form hint ([d055462](https://github.com/chenwei791129/launchpal/commit/d0554624135cf735e9374958bfd7b74812d69d40))
* prevent watch loop in ScheduleForm between modelValue and emit ([d85ec57](https://github.com/chenwei791129/launchpal/commit/d85ec578bdbed54ddb8b2cd25cd19c039ed717e9))
* show blue indicator for loaded services and correct action button ([ffe5d28](https://github.com/chenwei791129/launchpal/commit/ffe5d28d7839d6026fe58a5321ae616734630d29))
* write empty StartCalendarInterval for every-minute schedule ([5188c76](https://github.com/chenwei791129/launchpal/commit/5188c76e3ca1fd2f37b6dc5bc3424aac43fd41a6))

## 1.0.0 (2026-01-28)


### Features

* add backup restore functionality in settings page ([0bd9278](https://github.com/chenwei791129/launchpal/commit/0bd9278102c0287fd06914f9fa065d66c56637db))
* add click to copy for stdout and stderr paths ([2667f1b](https://github.com/chenwei791129/launchpal/commit/2667f1b0757963e227f91a5f7e33d9e0f0ebb64e))
* add Create Service modal with form ([1bafe8e](https://github.com/chenwei791129/launchpal/commit/1bafe8ec07a1a083ebd09e6fd48d146e1deb89c1))
* add issue and pull request templates for better contribution guidelines ([9d0ba7c](https://github.com/chenwei791129/launchpal/commit/9d0ba7c235527a788645c995b5266aa23c578481))
* add LaunchPal - macOS LaunchAgent GUI manager ([20c3772](https://github.com/chenwei791129/launchpal/commit/20c37720c387175ea001a28bceca0b56f68ce2aa))
* add plist format indicator in service summary ([7b6b896](https://github.com/chenwei791129/launchpal/commit/7b6b896eac1efd809e9f3fd10ccd8b5b6eb0652d))
* add read-only support for system LaunchDaemons ([#2](https://github.com/chenwei791129/launchpal/issues/2)) ([e7d3b63](https://github.com/chenwei791129/launchpal/commit/e7d3b63adcfcf8a38b2122d0ec344a076dd2646e))
* add settings page with version and about info ([2fb9009](https://github.com/chenwei791129/launchpal/commit/2fb900997344728fcbfbface9203ea8ddad1e59c))
* add TailwindCSS with dark theme configuration ([e35053c](https://github.com/chenwei791129/launchpal/commit/e35053c1f81028b99300140d063d2f21e426e13c))
* add Wails app bindings for service management ([5af0d5b](https://github.com/chenwei791129/launchpal/commit/5af0d5b99000dfae2157dbd2374c7fbacb6422e1))
* add XML syntax highlighting for plist in Inspect tab ([9c7b38d](https://github.com/chenwei791129/launchpal/commit/9c7b38d0e270bd174298131d9c9a42f84c44992f))
* auto-generate log paths when creating new service ([067b944](https://github.com/chenwei791129/launchpal/commit/067b944b7f39155eee29927a385b1e2540a286c5))
* create main layout with sidebar and status bar ([ffe8507](https://github.com/chenwei791129/launchpal/commit/ffe85072a579a6b8b78dd88d3d0da327e8f5e5ab))
* create service detail page with summary, logs, and inspect tabs ([acac62c](https://github.com/chenwei791129/launchpal/commit/acac62cd774c8b4a4197ecd9d7d1e42345d4cb13))
* create services list page with search and actions ([977b1bb](https://github.com/chenwei791129/launchpal/commit/977b1bb6188d14eba24d15dfa5b2577cdac2d617))
* define Service types and Manager interface ([c4cb11e](https://github.com/chenwei791129/launchpal/commit/c4cb11e88d31e59af50f124af39eb1824beeb9ce))
* display plist file path in service summary with copy support ([374a70f](https://github.com/chenwei791129/launchpal/commit/374a70f1fb0b04c4b3bdaa9ae1d5360742db7bd3))
* implement backup system with auto-backup on update ([8a67f44](https://github.com/chenwei791129/launchpal/commit/8a67f44293522274cd4c176eee8be61dc0f6611c))
* implement UserManager for LaunchAgents ([1ce6efb](https://github.com/chenwei791129/launchpal/commit/1ce6efb2c0c64949bb750b6eac21a34a5cce73da))
* initialize Wails v2 project structure ([d62e5d2](https://github.com/chenwei791129/launchpal/commit/d62e5d2e62cb00df2ae8d5e28681e592f18f9905))
* replace vanilla frontend with Nuxt 4 ([3819d25](https://github.com/chenwei791129/launchpal/commit/3819d25db25631b35e809fa7716cf3a8c98b6122))
* store original plist path in backup metadata ([3ea2269](https://github.com/chenwei791129/launchpal/commit/3ea2269a2f48573f7cd3e11fb30df70d1d989a70))


### Bug Fixes

* adjust table column widths to prevent text overlap ([0db94d2](https://github.com/chenwei791129/launchpal/commit/0db94d2ba7eecf320a38138764340b049bf7c281))
* always show action buttons instead of on hover ([2ba7942](https://github.com/chenwei791129/launchpal/commit/2ba7942523fea1fff4a2a13aeaed97c74f9546c4))
* auto-backup before deleting service ([6dc6cf2](https://github.com/chenwei791129/launchpal/commit/6dc6cf2180357e6b7b1c01ab50853eca7adcfae4))
* **ci:** build frontend before running tests ([fd27f9a](https://github.com/chenwei791129/launchpal/commit/fd27f9aed823daf41e2d6435b5798ac79c9e2b3e))
* handle binary plist format in system services ([85f323a](https://github.com/chenwei791129/launchpal/commit/85f323a640d723245e3c0e15b7c67354142f9b80))
* hide native window title to prevent text overlap ([8b5f41c](https://github.com/chenwei791129/launchpal/commit/8b5f41c542ec7ac662c5b1f4c96dbee013e37997))
* improve text wrapping in ServiceSummary to prevent overflow ([9dccf30](https://github.com/chenwei791129/launchpal/commit/9dccf30cd7b7041cfff5d6e5caca0704c5b47bdb))
* make dev builds and opens app ([7ac3b19](https://github.com/chenwei791129/launchpal/commit/7ac3b1956011136f142da7fd4b5ce80f792a4681))
* prevent text overflow from background boxes in path fields ([a574d93](https://github.com/chenwei791129/launchpal/commit/a574d93c3118eb0090d6724986875c81b4f2cba6))
* rename Auto badge to RunAtLoad for clarity ([9a2c85e](https://github.com/chenwei791129/launchpal/commit/9a2c85e89ce579d4fe388949b2b3ea87697d0a07))
* search by service label only, not file path ([2a306cb](https://github.com/chenwei791129/launchpal/commit/2a306cbc6d9fd4a21631e8ee79358052770e2e04))
* show status indicator for loaded services ([48a9209](https://github.com/chenwei791129/launchpal/commit/48a9209bfb4456475adddfe849b3e7e90c1092eb))
* sidebar button background alignment ([1dc280f](https://github.com/chenwei791129/launchpal/commit/1dc280ffacbeef45d33dbbac20b300b238f95838))
* skip pgrep fallback for common shells to avoid false matches ([2e8c09d](https://github.com/chenwei791129/launchpal/commit/2e8c09dc5f17d7e99d3606557d3e317e04aa367a))
* try to kill process if launchctl unload fails ([d38d565](https://github.com/chenwei791129/launchpal/commit/d38d565ef76041edf5557e815ea40bfab9469c06))
* use pgrep fallback when launchctl doesn't report PID ([46bfa08](https://github.com/chenwei791129/launchpal/commit/46bfa0806527a7a5117236feabc22bcad9aa60f7))
