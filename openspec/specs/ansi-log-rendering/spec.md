# ansi-log-rendering Specification

## Purpose

TBD - created by archiving change 'render-ansi-log-colors'. Update Purpose after archive.

## Requirements

### Requirement: Render ANSI SGR colors in service log output

The system SHALL parse ANSI SGR escape sequences (Select Graphic Rendition, `ESC [ ... m`) embedded in stdout and stderr log content and render them as HTML `<span>` elements with corresponding inline style attributes. Plain text characters SHALL be HTML-escaped before insertion into the output. The supported SGR code set SHALL be the union of: reset (`0`), bold (`1`), underline (`4`), foreground colors 30 through 37, bright foreground colors 90 through 97, background colors 40 through 47, and bright background colors 100 through 107.

#### Scenario: Plain text without escape sequences

- **WHEN** the log content is `"hello world\n"` with no escape sequences
- **THEN** the rendered output equals `"hello world\n"` with no `<span>` elements

#### Scenario: Single foreground color span

- **WHEN** the log content is `"\x1b[31mERROR\x1b[0m: connection refused"`
- **THEN** the rendered output contains `<span style="color:#e06c75">ERROR</span>` followed by `: connection refused`

#### Scenario: Bold combined with color

- **WHEN** the log content is `"\x1b[1m\x1b[33mWARN\x1b[0m"`
- **THEN** the rendered output contains a single `<span>` whose style includes both `font-weight:bold` and `color:#e5c07b`, wrapping the text `WARN`

#### Scenario: Multiple parameters in one SGR sequence

- **WHEN** the log content is `"\x1b[1;31mFATAL\x1b[0m"`
- **THEN** the rendered output contains a single `<span>` whose style includes both `font-weight:bold` and `color:#e06c75`, wrapping the text `FATAL`

##### Example: SGR code to style mapping

| SGR code | Effect           | Style fragment             |
| -------- | ---------------- | -------------------------- |
| 0        | reset            | closes current span        |
| 1        | bold             | `font-weight:bold`         |
| 4        | underline        | `text-decoration:underline`|
| 30       | black fg         | `color:#5c6370`            |
| 31       | red fg           | `color:#e06c75`            |
| 32       | green fg         | `color:#98c379`            |
| 33       | yellow fg        | `color:#e5c07b`            |
| 34       | blue fg          | `color:#61afef`            |
| 35       | magenta fg       | `color:#c678dd`            |
| 36       | cyan fg          | `color:#56b6c2`            |
| 37       | white fg         | `color:#abb2bf`            |
| 90       | bright black fg  | `color:#828896`            |
| 91       | bright red fg    | `color:#ff7b85`            |
| 92       | bright green fg  | `color:#b5e890`            |
| 93       | bright yellow fg | `color:#ffd97d`            |
| 94       | bright blue fg   | `color:#82c8ff`            |
| 95       | bright magenta fg| `color:#e08af0`            |
| 96       | bright cyan fg   | `color:#73d1de`            |
| 97       | bright white fg  | `color:#ffffff`            |
| 40       | black bg         | `background-color:#5c6370` |
| 41       | red bg           | `background-color:#e06c75` |
| 42       | green bg         | `background-color:#98c379` |
| 43       | yellow bg        | `background-color:#e5c07b` |
| 44       | blue bg          | `background-color:#61afef` |
| 45       | magenta bg       | `background-color:#c678dd` |
| 46       | cyan bg          | `background-color:#56b6c2` |
| 47       | white bg         | `background-color:#abb2bf` |
| 100      | bright black bg  | `background-color:#828896` |
| 101      | bright red bg    | `background-color:#ff7b85` |
| 102      | bright green bg  | `background-color:#b5e890` |
| 103      | bright yellow bg | `background-color:#ffd97d` |
| 104      | bright blue bg   | `background-color:#82c8ff` |
| 105      | bright magenta bg| `background-color:#e08af0` |
| 106      | bright cyan bg   | `background-color:#73d1de` |
| 107      | bright white bg  | `background-color:#ffffff` |


<!-- @trace
source: render-ansi-log-colors
updated: 2026-06-01
code:
  - frontend/app/utils/ansiToHtml.ts
  - frontend/app/utils/launchPolicy.ts
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/keepalive.go
  - frontend/app/utils/serviceToConfig.ts
  - CHANGELOG.md
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
-->

---
### Requirement: Strip unsupported and malformed escape sequences

The system SHALL remove all ANSI escape sequences that are not in the supported SGR subset, preserving any surrounding plain text. The set of stripped sequences SHALL include: SGR codes outside the supported subset (including 256-color `38;5;N`, truecolor `38;2;R;G;B`, blink `5`, reverse video `7`, and any other unlisted code), non-SGR CSI sequences (any `ESC [ ... <final-byte>` where the final byte is not `m`), OSC sequences (`ESC ] ...` terminated by `BEL` `\x07` or `ESC \`), DCS / SOS / PM / APC sequences (`ESC P`, `ESC X`, `ESC ^`, `ESC _` until terminator), and any `ESC [` with no terminating final byte before end-of-input. The system SHALL NOT throw, alert, or otherwise surface an error when stripping; the remaining content SHALL continue to render.

#### Scenario: 256-color escape is stripped

- **WHEN** the log content is `"\x1b[38;5;33mhi\x1b[0m"`
- **THEN** the rendered output contains the text `hi` and no `<span>` element

#### Scenario: Non-SGR CSI escape is stripped

- **WHEN** the log content is `"\x1b[2Jcleared"`
- **THEN** the rendered output equals `cleared` with no `<span>` and no `[2J` literal

#### Scenario: Unterminated CSI at end of input

- **WHEN** the log content is `"text\x1b[31"`
- **THEN** the rendered output equals `text` with no `<span>` and no `[31` literal

#### Scenario: Mixed supported and unsupported parameters in one SGR

- **WHEN** the log content is `"\x1b[1;38;5;33mhi\x1b[0m"`
- **THEN** the rendered output equals the text `hi` with no `<span>` element, because the SGR sequence contains an unsupported parameter and is treated as a whole-sequence strip

#### Scenario: OSC sequence with BEL terminator

- **WHEN** the log content is `"\x1b]0;window-title\x07text"`
- **THEN** the rendered output equals `text`

##### Example: stripping behavior cases

| Input                            | Rendered output  | Notes                              |
| -------------------------------- | ---------------- | ---------------------------------- |
| `"\x1b[38;2;255;0;0mhi\x1b[0m"`  | `hi`             | truecolor stripped                 |
| `"\x1b[5mblink\x1b[0m"`          | `blink`          | blink code unsupported             |
| `"\x1b[Hhi"`                     | `hi`             | cursor home stripped               |
| `"\x1b[zzzhi"`                   | `hi`             | malformed final byte stripped      |
| `"a\x1b[31"`                     | `a`              | unterminated trailing CSI stripped |
| `"\x1b]0;title\x1b\\after"`      | `after`          | OSC with ST terminator stripped    |


<!-- @trace
source: render-ansi-log-colors
updated: 2026-06-01
code:
  - frontend/app/utils/ansiToHtml.ts
  - frontend/app/utils/launchPolicy.ts
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/keepalive.go
  - frontend/app/utils/serviceToConfig.ts
  - CHANGELOG.md
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
-->

---
### Requirement: HTML-escape plain text and disallow non-whitelisted style attributes

The system SHALL HTML-escape all plain text fragments before placing them between span tags or directly into the output. The escape mapping SHALL be: `&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`, `"` → `&quot;`, `'` → `&#39;`. The `style` attribute of any emitted `<span>` SHALL contain only the four whitelisted CSS properties: `color`, `background-color`, `font-weight`, `text-decoration`. Property values SHALL be drawn only from the compile-time constant `SGR_COLOR_MAP` (or the literal values `bold` / `underline`); the system SHALL NOT interpolate any portion of the original log text into a style attribute.

#### Scenario: HTML special characters in plain text are escaped

- **WHEN** the log content is `"<script>alert(1)</script>"`
- **THEN** the rendered output equals `&lt;script&gt;alert(1)&lt;/script&gt;` and contains no live `<script>` tag

#### Scenario: Quotes inside plain text are escaped

- **WHEN** the log content is `"value=\"hi\""`
- **THEN** the rendered output equals `value=&quot;hi&quot;` with no literal `"` characters between tags

#### Scenario: Attacker payload inside an SGR-wrapped span is escaped

- **WHEN** the log content is `"\x1b[31m\" onmouseover=alert(1) x=\"\x1b[0m"`
- **THEN** the rendered output wraps the escaped text inside `<span style="color:#e06c75">` and no `onmouseover` attribute appears outside the controlled `style` attribute


<!-- @trace
source: render-ansi-log-colors
updated: 2026-06-01
code:
  - frontend/app/utils/ansiToHtml.ts
  - frontend/app/utils/launchPolicy.ts
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/keepalive.go
  - frontend/app/utils/serviceToConfig.ts
  - CHANGELOG.md
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
-->

---
### Requirement: Mount rendered output in ServiceLogs view

The `ServiceLogs.vue` component SHALL render log content through a `v-html`-bound `<pre>` element whose contents come from a `renderedLogs` computed value defined as `ansiToHtml` applied to the log content string carried in the `LogsResult` returned by `GetLogs` / `GetSystemLogs` (or the empty string when no content is loaded). The CSS classes `text-gray-300`, `whitespace-pre-wrap`, `break-all`, `font-mono`, and `text-sm` SHALL remain on the `<pre>` element so that monospace, line-wrapping, and color base behavior are unchanged. The existing loading and error branches SHALL be preserved; the placeholder branch SHALL follow the `log-load-feedback` capability, which extends it with status-specific wording for the `no-path` and `not-found` states.

#### Scenario: Logs containing ANSI colors render as colored spans

- **WHEN** `GetLogs` resolves with Status "ok" and Content `"\x1b[32mOK\x1b[0m booted"` and the Logs tab is opened
- **THEN** the `<pre>` element contains `<span style="color:#98c379">OK</span>` followed by ` booted`, and the literal characters `[32m` and `[0m` do not appear in the rendered DOM

#### Scenario: Empty log preserves placeholder

- **WHEN** `GetLogs` resolves with Status "ok" and empty Content
- **THEN** the component renders the existing "No logs available for {logType}" placeholder branch and does not render the `<pre>` element

#### Scenario: Existing loading state is unaffected

- **WHEN** `loading` is true and no log content is loaded
- **THEN** the component renders the existing spinner branch ("Loading logs..."), not the `<pre>` element

#### Scenario: Existing error state is unaffected

- **WHEN** the `GetLogs` promise rejects and `error.value` is set
- **THEN** the component renders the existing red-text error branch and does not render the `<pre>` element

<!-- @trace
source: fix-log-error-classification
updated: 2026-07-04
code:
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/launchctl/user.go
  - app.go
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/apple_system.go
  - internal/launchctl/manager.go
tests:
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/system_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/user_test.go
-->