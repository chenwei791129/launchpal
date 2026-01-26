<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="px-4 py-3 border-b border-surface-100">
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 text-sm mb-3">
        <NuxtLink to="/" class="text-gray-400 hover:text-white transition-colors">
          Services
        </NuxtLink>
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-white">Settings</span>
      </div>

      <!-- Title -->
      <div class="flex items-center gap-3">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <h1 class="text-lg font-semibold text-white">Settings</h1>
      </div>
    </header>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 space-y-6">
      <!-- Version Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">Version</h2>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-primary-600/20 rounded-lg flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <p class="text-white font-medium">LaunchPal</p>
              <p class="text-gray-400 text-sm">Version {{ appVersion }}</p>
            </div>
          </div>
          <span class="px-3 py-1 bg-primary-600/20 text-primary-400 text-sm rounded-full">
            {{ appVersion }}
          </span>
        </div>
      </section>

      <!-- Backup Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">Backup Storage</h2>
        <div class="flex items-start gap-3">
          <div class="w-10 h-10 bg-surface-200 rounded-lg flex items-center justify-center flex-shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-white font-medium mb-1">Backup Directory</p>
            <p class="text-gray-400 text-sm mb-2">Service backups are stored at:</p>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-surface-500 rounded-lg text-sm text-gray-300 font-mono truncate">
                {{ backupPath }}
              </code>
              <button
                class="p-2 rounded-lg hover:bg-surface-200 text-gray-400 hover:text-white transition-colors flex-shrink-0"
                title="Copy path"
                @click="copyBackupPath"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- About Section -->
      <section class="bg-surface-400 rounded-xl p-4">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-4">About</h2>
        <div class="space-y-4">
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 bg-gradient-to-br from-primary-500 to-primary-700 rounded-2xl flex items-center justify-center shadow-lg">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h3 class="text-xl font-semibold text-white">LaunchPal</h3>
              <p class="text-gray-400">macOS LaunchAgent 圖形化管理工具</p>
            </div>
          </div>

          <div class="border-t border-surface-100 pt-4">
            <p class="text-gray-400 text-sm leading-relaxed">
              LaunchPal 是一款專為 macOS 設計的 LaunchAgent 管理工具，提供直觀的圖形化介面，
              讓您輕鬆管理系統服務。支援服務的啟動、停止、重啟，以及 plist 檔案的檢視與編輯。
            </p>
          </div>

          <div class="border-t border-surface-100 pt-4">
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p class="text-gray-500 mb-1">Platform</p>
                <p class="text-white">macOS</p>
              </div>
              <div>
                <p class="text-gray-500 mb-1">Framework</p>
                <p class="text-white">Wails + Vue</p>
              </div>
              <div>
                <p class="text-gray-500 mb-1">License</p>
                <p class="text-white">MIT</p>
              </div>
              <div>
                <p class="text-gray-500 mb-1">Language</p>
                <p class="text-white">Go + TypeScript</p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
const appVersion = 'v0.1.0'
const backupPath = '~/.launchpal/backups/'

async function copyBackupPath() {
  try {
    await navigator.clipboard.writeText(backupPath)
  } catch (e) {
    console.error('Failed to copy to clipboard:', e)
  }
}
</script>
