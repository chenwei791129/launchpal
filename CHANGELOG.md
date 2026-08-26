# Changelog

## [1.22.0](https://github.com/chenwei791129/launchpal/compare/v1.21.1...v1.22.0) (2026-08-26)


### Features

* add Status and Type filters to service list pages ([75f8d2f](https://github.com/chenwei791129/launchpal/commit/75f8d2fd536de518d971a50d11a15a4d89e091ba))
* show resolved log path and file size in the Logs tab ([6a707f5](https://github.com/chenwei791129/launchpal/commit/6a707f533817238e0eec9169c999598060f85ea8))


### Bug Fixes

* raise the minimum supported macOS version to 13 Ventura ([534d770](https://github.com/chenwei791129/launchpal/commit/534d770017499d0b89c65a200d18c67816150bd9))

## [1.21.1](https://github.com/chenwei791129/launchpal/compare/v1.21.0...v1.21.1) (2026-08-25)


### Bug Fixes

* derive user service status solely from launchd ([c250caf](https://github.com/chenwei791129/launchpal/commit/c250cafb5539edf3c9eba8855a7358b302d1d68f))

## [1.21.0](https://github.com/chenwei791129/launchpal/compare/v1.20.0...v1.21.0) (2026-07-23)


### Features

* harden Admin Mode privileged-helper lifecycle ([9f5beb4](https://github.com/chenwei791129/launchpal/commit/9f5beb48185146799f710ec8e60195f8ac3c5366))
* verify privileged-helper launch integrity via root-owned protected copy ([36be6bc](https://github.com/chenwei791129/launchpal/commit/36be6bceabd83df9cee67581877a92f568896f5e))


### Bug Fixes

* harden filesystem inputs in launchctl managers and privhelper ([63a5954](https://github.com/chenwei791129/launchpal/commit/63a595401b19adc30c3856e5ec87231de53efbbd))

## [1.20.0](https://github.com/chenwei791129/launchpal/compare/v1.19.0...v1.20.0) (2026-07-04)


### Features

* auto-refresh the Logs tab with a 2s polling toggle ([9eb5c27](https://github.com/chenwei791129/launchpal/commit/9eb5c2739cae066ddc31c41308f477ef7f43fdbe))
* classify log load results and surface backend errors in Logs tab ([6b3b9f9](https://github.com/chenwei791129/launchpal/commit/6b3b9f93fc4a940fd8c01b5e754d4f71dcf99463))


### Bug Fixes

* discard superseded concurrent log loads in ServiceLogs ([a48bc8e](https://github.com/chenwei791129/launchpal/commit/a48bc8e7e7a8dfdf9f6e916c2224bab5e73f5f81))

## [1.19.0](https://github.com/chenwei791129/launchpal/compare/v1.18.0...v1.19.0) (2026-06-26)


### Features

* preserve unmodeled plist keys on service update ([#37](https://github.com/chenwei791129/launchpal/issues/37)) ([36bec75](https://github.com/chenwei791129/launchpal/commit/36bec75ce97663afcaf8c468a24bbf7b45174022))

## [1.18.0](https://github.com/chenwei791129/launchpal/compare/v1.17.0...v1.18.0) (2026-06-01)


### Features

* render ANSI SGR colors in service logs ([a646a74](https://github.com/chenwei791129/launchpal/commit/a646a74c49e09a4095412ea58eb1e5ee278af558))
* structured KeepAlive config with launch-policy radio ([15b5f89](https://github.com/chenwei791129/launchpal/commit/15b5f894440d24e06cac0710dcf76f1ceb70275c))

## [1.17.0](https://github.com/chenwei791129/launchpal/compare/v1.16.0...v1.17.0) (2026-05-28)


### Features

* clone user service via Copy button on detail page ([ea1d9e8](https://github.com/chenwei791129/launchpal/commit/ea1d9e8529a604456f66edac38e1e254c14e3c60))

## [1.16.0](https://github.com/chenwei791129/launchpal/compare/v1.15.0...v1.16.0) (2026-05-16)


### Features

* **system-daemon:** optional log file cleanup on delete ([b3d828e](https://github.com/chenwei791129/launchpal/commit/b3d828e170e0944e375bd43a92715b87fc759e54))


### Bug Fixes

* **service-form:** allow Program path to be empty when Arguments holds the executable ([b0108d8](https://github.com/chenwei791129/launchpal/commit/b0108d8b38380d078ebffad99f52b4bdba46abb8))

## [1.15.0](https://github.com/chenwei791129/launchpal/compare/v1.14.0...v1.15.0) (2026-05-09)


### Features

* customize log directory paths via settings ([69ea9f2](https://github.com/chenwei791129/launchpal/commit/69ea9f2235eba34debac5abc7392c72f8d0fad6c))


### Bug Fixes

* **frontend:** use ISO-style en-CA locale in formatTimestamp ([9ccb9b3](https://github.com/chenwei791129/launchpal/commit/9ccb9b36cd9cb924d23f54f44e2444d8155c7da3))
* **launchctl:** drop pgrep+kill fallback from UserManager.Stop ([faacae2](https://github.com/chenwei791129/launchpal/commit/faacae2af46ce657f2a549d9754f10f3fb7ead25))

## [1.14.0](https://github.com/chenwei791129/launchpal/compare/v1.13.1...v1.14.0) (2026-05-03)


### Features

* **logs:** add Clear Logs control with per-file permission dispatch ([#29](https://github.com/chenwei791129/launchpal/issues/29)) ([9f0f900](https://github.com/chenwei791129/launchpal/commit/9f0f90068ff62a1e2725160dc8a89c2f18916c01))

## [1.13.1](https://github.com/chenwei791129/launchpal/compare/v1.13.0...v1.13.1) (2026-05-02)


### Bug Fixes

* **ci:** copy launchpal-privhelper into app bundle during build ([29b6584](https://github.com/chenwei791129/launchpal/commit/29b6584ddf1197346a3f38b14ccaa8dc1c516d46))

## [1.13.0](https://github.com/chenwei791129/launchpal/compare/v1.12.1...v1.13.0) (2026-04-22)


### Features

* **admin-mode:** add session-scoped privileged helper for system daemon writes ([#25](https://github.com/chenwei791129/launchpal/issues/25)) ([3c4ed37](https://github.com/chenwei791129/launchpal/commit/3c4ed37babc6e72f4fd73409027e1731c8c492bd))

## [1.12.1](https://github.com/chenwei791129/launchpal/compare/v1.12.0...v1.12.1) (2026-04-21)


### Performance Improvements

* **launchctl:** replace per-service pgrep fork with a single ps snapshot ([953afaa](https://github.com/chenwei791129/launchpal/commit/953afaa27bd30a9155dadfae2c6d0bfc79d6f05d))

## [1.12.0](https://github.com/chenwei791129/launchpal/compare/v1.11.1...v1.12.0) (2026-04-21)


### Features

* **launchctl:** add heuristic status detection for system daemons ([281e3c8](https://github.com/chenwei791129/launchpal/commit/281e3c85c6e6098b194964129639f2f3564d4f11))
* **ui:** surface unverified status confidence with info icon ([2ffe296](https://github.com/chenwei791129/launchpal/commit/2ffe29668e2781234b51776e3a62d2f2f8d07c3b))


### Bug Fixes

* **ui:** restore StatusConfidenceIcon tooltip visibility and switch to English ([df7c510](https://github.com/chenwei791129/launchpal/commit/df7c510eb4a6bd6a55ba3c4b6abd68aa45057144))
* **ui:** translate About section from Chinese to English ([d325d28](https://github.com/chenwei791129/launchpal/commit/d325d28387d46a2e185e7c0770a12a62a787e750))

## [1.11.1](https://github.com/chenwei791129/launchpal/compare/v1.11.0...v1.11.1) (2026-04-21)


### Bug Fixes

* **launchctl:** refuse empty-label bootout/kickstart targets ([36e7032](https://github.com/chenwei791129/launchpal/commit/36e703206e9fea01db6bbcfd91ad02d7ec08e638))

## [1.11.0](https://github.com/chenwei791129/launchpal/compare/v1.10.1...v1.11.0) (2026-04-18)


### Features

* **backup:** add side-by-side diff preview before restore ([cbe0761](https://github.com/chenwei791129/launchpal/commit/cbe0761b71d70d050062a4abcee6a5de3a516486))

## [1.10.1](https://github.com/chenwei791129/launchpal/compare/v1.10.0...v1.10.1) (2026-04-14)


### Bug Fixes

* make Summary tab scrollable when content overflows ([e2916f4](https://github.com/chenwei791129/launchpal/commit/e2916f438c9558866c15bfa84d12c3258d735236))

## [1.10.0](https://github.com/chenwei791129/launchpal/compare/v1.9.0...v1.10.0) (2026-04-11)


### Features

* add cron range and enumeration syntax with expansion preview ([b1eea86](https://github.com/chenwei791129/launchpal/commit/b1eea864353341bf21b1680e60e44dde06de3cd9))
* mask environment variable values by default ([09e33ee](https://github.com/chenwei791129/launchpal/commit/09e33ee0b52db07c66d4fa2253f1d5afd97a81e2))
* support multiple calendar interval schedules with CalendarEntry ([9c6bb60](https://github.com/chenwei791129/launchpal/commit/9c6bb606a12d22015d0348a1e6f322e4a62b379a))
* support multiple schedule entries in next occurrence calculation ([90ea5b4](https://github.com/chenwei791129/launchpal/commit/90ea5b464387d5070221ee5657d1c2e4902ac8bb))
* update ServiceSummary for multiple schedule entries ([c892512](https://github.com/chenwei791129/launchpal/commit/c892512707ba17776b4e693b894b2b755e946756))
* update TypeScript types for CalendarEntry and ScheduleConfig ([c534aa5](https://github.com/chenwei791129/launchpal/commit/c534aa599b7f1e494352478b12b807dfdc1011f8))

## [1.9.0](https://github.com/chenwei791129/launchpal/compare/v1.8.0...v1.9.0) (2026-04-04)


### Features

* add Run Now (kickstart) button for user services ([94a8659](https://github.com/chenwei791129/launchpal/commit/94a8659772921f3fa4e09f8b433c63748a2b6d1e))
* migrate Start/Stop from legacy load/unload to bootstrap/bootout ([f05b713](https://github.com/chenwei791129/launchpal/commit/f05b71384cdeff38819e8f57b12bb616d5327a11))

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
