import Vue from 'vue'
import VueRouter, { RouteConfig, Route, NavigationGuardNext } from 'vue-router'
import store from '@/store'

import DownloadsWorkspace from '@/components/workspaces/DownloadsWorkspace.vue'
import GameConfiguration from '@/components/GameConfiguration.vue'
import MapWorkspace from '@/components/workspaces/MapWorkspace.vue'
import SessionsWorkspace from '@/components/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/components/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/components/workspaces/SettingsWorkspace.vue'
import UserManagement from '@/components/UserManagement.vue'
import UserToolWorkspace from '@/components/workspaces/UserToolWorkspace.vue'
import RouteShader from '@/components/RouteShader.vue'
import SessionDetails from '@/components/SessionDetails.vue'
import UserSessions from '@/components/UserSessions.vue'

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
        path: 'game-config',
        name: 'config',
        component: GameConfiguration
      },
      {
        path: 'users',
        name: 'users',
        component: UserManagement
      },
      {
        path: 'route-shader',
        name: 'shader',
        component: RouteShader
      }
    ]
  },
  {
    path: '*',
    name: 'undefined'
  }
]

const router = new VueRouter({
  routes
})

router.beforeEach(async (to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  if ((!store.getters.isAdmin && !store.getters.isOwner && to.name === 'settings') || to.name === 'undefined') {
    next('/')
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    return
  }
  if (to.name === 'settings') {
    store.commit('UPDATE_CURRENT_PAGE', 'users')
    next('/settings/users')
    return
  }
  store.commit('UPDATE_CURRENT_PAGE', to.name)
  next()
})

export default router
