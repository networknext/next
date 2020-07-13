import Vue from 'vue'
import VueRouter, { RouteConfig } from 'vue-router'

import DownloadsWorkspace from '@/components/workspaces/DownloadsWorkspace.vue'
import GameConfiguration from '@/components/GameConfiguration.vue'
import MapWorkspace from '@/components/workspaces/MapWorkspace.vue'
import SessionDetails from '@/components/SessionDetails.vue'
import SessionsWorkspace from '@/components/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/components/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/components/workspaces/SettingsWorkspace.vue'
import UserManagement from '@/components/UserManagement.vue'
import UserSessions from '@/components/UserSessions.vue'
import UserToolWorkspace from '@/components/workspaces/UserToolWorkspace.vue'

Vue.use(VueRouter)

const routes: Array<RouteConfig> = [
  {
    path: '/',
    name: 'map',
    component: MapWorkspace
  },
  {
    path: '/sessions',
    name: 'sessions',
    component: SessionsWorkspace
  },
  {
    path: '/session-tool',
    name: 'session-tool',
    component: SessionToolWorkspace,
    children: [
      {
        path: '*',
        component: SessionDetails
      }
    ]
  },
  {
    path: '/user-tool',
    name: 'user-tool',
    component: UserToolWorkspace,
    children: [
      {
        path: '*',
        component: UserSessions
      }
    ]
  },
  {
    path: '/downloads',
    name: 'downloads',
    component: DownloadsWorkspace
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsWorkspace,
    children: [
      {
        path: 'game-config',
        name: 'config',
        component: GameConfiguration
      },
      {
        path: 'users',
        name: 'users',
        component: UserManagement
      }
    ]
  }
]

const router = new VueRouter({
  routes
})

export default router
