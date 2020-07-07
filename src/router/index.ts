import Vue from 'vue'
import VueRouter, { RouteConfig } from 'vue-router'

import MapWorkspace from '@/components/workspaces/MapWorkspace.vue'
import SessionsWorkspace from '@/components/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/components/workspaces/SessionToolWorkspace.vue'
import UserToolWorkspace from '@/components/workspaces/UserToolWorkspace.vue'
import SettingsWorkspace from '@/components/workspaces/SettingsWorkspace.vue'
import DownloadsWorkspace from '@/components/workspaces/DownloadsWorkspace.vue'

Vue.use(VueRouter)

const routes: Array<RouteConfig> = [
  {
    path: '/',
    name: 'Map',
    component: MapWorkspace
  },
  {
    path: '/sessions',
    name: 'Sessions',
    component: SessionsWorkspace
  },
  {
    path: '/session-tool',
    name: 'Session Tool',
    component: SessionToolWorkspace
  },
  {
    path: '/user-tool',
    name: 'User Tool',
    component: UserToolWorkspace
  },
  {
    path: '/downloads',
    name: 'Downloads',
    component: DownloadsWorkspace
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsWorkspace
  }
]

const router = new VueRouter({
  routes
})

export default router
