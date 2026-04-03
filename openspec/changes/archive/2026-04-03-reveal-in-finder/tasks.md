## 1. 後端實作

- [x] [P] 1.1 在 `app.go` 新增 `RevealInFinder(path string) error` 方法，使用 `exec.Command("open", "-R", path).Start()` 在 Finder 中顯示指定檔案（Reveal plist file in Finder）

## 2. 前端實作

- [x] [P] 2.1 在 `frontend/app/components/ServiceSummary.vue` 的 Plist File 區塊 copy icon 右側新增「在 Finder 中顯示」按鈕（folder icon），使用 `@click.stop` 防止觸發複製（Button does not trigger copy），呼叫後端 `RevealInFinder` 方法（Reveal plist file in Finder, Reveal in Finder for system service）
