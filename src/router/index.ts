import Vue from 'vue'
import VueRouter, { RouteConfig, Route, NavigationGuardNext } from 'vue-router'
import store from '@/store'

import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import GameConfiguration from '@/components/GameConfiguration.vue'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import UserManagement from '@/components/UserManagement.vue'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'
import RouteShader from '@/components/RouteShader.vue'
import AccountSettings from '@/components/AccountSettings.vue'
import SessionDetails from '@/components/SessionDetails.vue'
import UserSessions from '@/components/UserSessions.vue'

Vue.use(VueRouter)

// All navigable routes for the Portal
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
        name: 'session-details',
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
        name: 'user-sessions',
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
        path: 'account',
        name: 'account-settings',
        component: AccountSettings
      },
      {
        path: 'game-config',
        name: 'config',
        component: GameConfiguration
      },
      {
        path: 'users',
        name: 'users',
        component: UserManagement
      }/* ,
      {
        path: 'route-shader',
        name: 'shader',
        component: RouteShader
      } */
    ]
  },
  {
    path: '*',
    name: 'undefined'
  }
]

const router = new VueRouter({
  mode: 'history',
  routes
})

// Catch all for routes. This can be used for a lot of different things like separating anon portal from authorized portal etc
router.beforeEach(async (to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  if ((!store.getters.isAdmin && !store.getters.isOwner && to.name === 'settings') || to.name === 'undefined') {
    next('/')
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    return
  }
  if (to.name === 'settings') {
    store.commit('UPDATE_CURRENT_PAGE', 'account-settings')
    next('/settings/account')
    return
  }
  store.commit('UPDATE_CURRENT_PAGE', to.name)
  next()
})

export default router
