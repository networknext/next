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
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  // TODO: Make sure these are doing what we want them to do.
  // TODO: store.getters.isAdmin doesn't work here. store.getters shows that everything is initialized correctly but accessing any of the members within getters, doesn't work?!
  // BUG: Re-routes valid users to the map when it should just refresh the page...
  if ((!store.getters.isAdmin && !store.getters.isOwner && (to.name === 'users' || to.name === 'game-config')) || to.name === 'undefined') {
    store.commit('UPDATE_CURRENT_PAGE', 'map');
    (window as any).Intercom('update')
    next('/')
    return
  }
  if (to.name === 'settings') {
    store.commit('UPDATE_CURRENT_PAGE', 'account-settings');
    (window as any).Intercom('update')
    next('/settings/account')
    return
  }
  // Email is verified
  if (to.query.message === 'Your email was verified. You can continue using the application.') {
    store.commit('UPDATE_CURRENT_PAGE', 'map');
    (window as any).Intercom('update')
    next('/')
    return
  }
  store.commit('UPDATE_CURRENT_PAGE', to.name);
  (window as any).Intercom('update')
  next()
})

export default router
