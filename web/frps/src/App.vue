<template>
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides">
    <NMessageProvider>
      <NLayout position="absolute">
        <NLayoutHeader bordered class="header">
          <div class="header-content">
            <div class="left">
              <NPopover trigger="click" placement="bottom-start" :show="showMenu" @update:show="showMenu = $event"
                v-if="isMobile">
                <template #trigger>
                  <NButton text class="menu-trigger">
                    <NIcon size="24">
                      <menu-outline />
                    </NIcon>
                  </NButton>
                </template>
                <div class="mobile-menu">
                  <NMenu 
                    :value="currentPath" 
                    :options="menuOptions" 
                    :expanded-keys="expandedKeys"
                    @update:value="handleMobileSelect" 
                  />
                </div>
              </NPopover>
              <NText class="logo">
                ME Frp 镜缘映射 - 服务端
              </NText>
            </div>
            <div class="right">
              <NSwitch v-model:value="darkmodeSwitch" @update:value="toggleDark" :rail-style="railStyle">
                <template #checked>
                  <NIcon><moon-outline /></NIcon>
                </template>
                <template #unchecked>
                  <NIcon><sunny-outline /></NIcon>
                </template>
              </NSwitch>
            </div>
          </div>
        </NLayoutHeader>

        <NLayout has-sider position="absolute" style="top: 64px;">
          <NLayoutSider v-show="!isMobile" bordered collapse-mode="width" :collapsed-width="64" :width="240"
            :collapsed="isCollapsed" show-trigger :native-scrollbar="false" @collapse="isCollapsed = true"
            @expand="isCollapsed = false">
            <NMenu :value="currentPath" :options="menuOptions" :expanded-keys="expandedKeys" :collapsed-width="64"
              :collapsed-icon-size="24" :icon-size="22" @update:value="handleSelect"
              @update:expanded-keys="handleExpand" />
          </NLayoutSider>

          <NLayout :native-scrollbar="false">
            <NLayoutContent style="padding: 24px;">
              <router-view ref="overviewRef" />
            </NLayoutContent>
          </NLayout>
        </NLayout>
      </NLayout>
    </NMessageProvider>
  </NConfigProvider>
</template>

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

:deep(.NPopover) {
  padding: 0;
}

:deep(.NPopover-content) {
  padding: 0;
}
</style>