<template>
  <n-config-provider :theme="theme" :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-layout position="absolute">
        <n-layout-header bordered class="header">
          <div class="header-content">
            <div class="left">
              <n-popover trigger="click" placement="bottom-start" :show="showMenu" @update:show="showMenu = $event"
                v-if="isMobile">
                <template #trigger>
                  <n-button text class="menu-trigger">
                    <n-icon size="24">
                      <menu-outline />
                    </n-icon>
                  </n-button>
                </template>
                <div class="mobile-menu">
                  <n-menu 
                    :value="currentPath" 
                    :options="menuOptions" 
                    :expanded-keys="expandedKeys"
                    @update:value="handleMobileSelect" 
                  />
                </div>
              </n-popover>
              <n-text class="logo">
                ME Frp 镜缘映射 - 服务端
              </n-text>
            </div>
            <div class="right">
              <n-switch v-model:value="darkmodeSwitch" @update:value="toggleDark" :rail-style="railStyle">
                <template #checked>
                  <n-icon><moon-outline /></n-icon>
                </template>
                <template #unchecked>
                  <n-icon><sunny-outline /></n-icon>
                </template>
              </n-switch>
            </div>
          </div>
        </n-layout-header>

        <n-layout has-sider position="absolute" style="top: 64px;">
          <n-layout-sider v-show="!isMobile" bordered collapse-mode="width" :collapsed-width="64" :width="240"
            :collapsed="isCollapsed" show-trigger :native-scrollbar="false" @collapse="isCollapsed = true"
            @expand="isCollapsed = false">
            <n-menu :value="currentPath" :options="menuOptions" :expanded-keys="expandedKeys" :collapsed-width="64"
              :collapsed-icon-size="24" :icon-size="22" @update:value="handleSelect"
              @update:expanded-keys="handleExpand" />
          </n-layout-sider>

          <n-layout :native-scrollbar="false">
            <n-layout-content style="padding: 24px;">
              <router-view ref="overviewRef" />
            </n-layout-content>
          </n-layout>
        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>

<style lang="scss" scoped>
html,
body,
#app {
  height: 100%;
  margin: 0;
  padding: 0;
}

.header {
  height: 64px;
  background: var(--n-color);
  border-bottom: 1px solid var(--n-border-color);
  position: relative;
  z-index: 1;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 64px;
  padding: 0 24px;

  .left {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.logo {
  font-size: 18px;
  font-weight: 500;
}

.menu-trigger {
  @media (min-width: 769px) {
    display: none;
  }
}

.right {
  display: flex;
  align-items: center;
}

:deep(.n-menu-item-content) {
  justify-content: flex-start !important;
}

:deep(.n-menu-item-content-header) {
  align-items: center;
}

.n-layout-sider {
  height: calc(100vh - 64px);
}

.n-layout-content {
  min-height: calc(100vh - 64px);
}

@media (max-width: 700px) {
  .n-layout-content {
    padding: 16px !important;
  }
}

.mobile-menu {
  min-width: 200px;
  margin: -8px -16px;
  background-color: var(--n-color);
  border-radius: 3px;
  box-shadow: var(--n-box-shadow);
}

:deep(.n-popover) {
  padding: 0;
}

:deep(.n-popover-content) {
  padding: 0;
}
</style>

<script setup lang="ts">
import { themeOverrides } from './constants/theme'
import { ref, computed, watch, h, onMounted, onBeforeUnmount } from 'vue'
import { darkTheme, useOsTheme } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import {
  HomeOutline,
  ServerOutline,
  SunnyOutline,
  MoonOutline,
  MenuOutline,
} from '@vicons/ionicons5'
import { useRouter, useRoute } from 'vue-router'

console.log(`    __  _________   ______         
   /  |/  / ____/  / ____/________ 
  / /|_/ / __/    / /_  / ___/ __ \\
 / /  / / /___   / __/ / /  / /_/ /
/_/  /_/_____/  /_/   /_/  / .___/ 
                          /_/      `);
console.log("Copyright 2025, The ME Frp Project Team.");
console.log("No redistribution allowed.");

const router = useRouter()
const route = useRoute()

const osThemeRef = useOsTheme()
const theme = ref(osThemeRef.value === 'dark' ? darkTheme : null)
const darkmodeSwitch = ref(osThemeRef.value === 'dark')

const expandedKeys = ref<string[]>(['/proxies'])

const toggleDark = (value: boolean) => {
  theme.value = value ? darkTheme : null
}

const currentPath = computed(() => route.path)

// 监听路由变化，保持菜单展开
watch(
  () => route.path,
  () => {
    // 始终保持隧道菜单展开
    expandedKeys.value = ['/proxies']
  },
  { immediate: true }
)

const renderIcon = (icon: any) => {
  return () => h(icon)
}

const menuOptions: MenuOption[] = [
  {
    label: '概览',
    key: '/',
    icon: renderIcon(HomeOutline)
  },
  {
    label: '隧道',
    key: '/proxies',
    icon: renderIcon(ServerOutline),
    children: [
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'TCP'),
        key: '/proxies/tcp'
      },
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'UDP'),
        key: '/proxies/udp'
      },
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'HTTP'),
        key: '/proxies/http'
      },
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'HTTPS'),
        key: '/proxies/https'
      },
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'TCPMUX'),
        key: '/proxies/tcpmux'
      },
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'STCP'),
        key: '/proxies/stcp'
      },
      {
        label: () => h('span', { style: 'font-weight: 500' }, 'SUDP'),
        key: '/proxies/sudp'
      }
    ]
  }
]

const handleSelect = (key: string) => {
  router.push(key)
}

const handleExpand = (keys: string[]) => {
  expandedKeys.value = keys
}

const showMenu = ref(false)
const isMobile = ref(window.innerWidth <= 700)
const isCollapsed = ref(false)

// 监听窗口大小变化
const handleWindowResize = () => {
  isMobile.value = window.innerWidth <= 700
}

onMounted(() => {
  window.addEventListener('resize', handleWindowResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleWindowResize)
})

// 移动端菜单选择处理
const handleMobileSelect = (key: string) => {
  router.push(key)
  showMenu.value = false
}

// Switch 按钮样式
const railStyle = ({ focused, checked }: { focused: boolean; checked: boolean }) => {
  return {
    background: checked ? themeOverrides.common?.primaryColor : undefined,
    boxShadow: focused ? `0 0 0 2px ${themeOverrides.common?.primaryColorSuppl}` : undefined
  }
}

const overviewRef = ref()
</script>
