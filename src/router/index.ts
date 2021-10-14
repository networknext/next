import Vue from 'vue'
import VueRouter, { RouteConfig, Route, NavigationGuardNext } from 'vue-router'
import store from '@/store'

import Billing from '@/components/Billing.vue'
import Analytics from '@/components/Analytics.vue'
import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import ExplorationWorkspace from '@/workspaces/ExplorationWorkspace.vue'
import GameConfiguration from '@/components/GameConfiguration.vue'
import LoginModal from '@/components/LoginModal.vue'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import Notifications from '@/components/Notifications.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import GetAccessModal from '@/components/GetAccessModal.vue'
import UserManagement from '@/components/UserManagement.vue'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'
import RouteShader from '@/components/RouteShader.vue'
import AccountSettings from '@/components/AccountSettings.vue'
import SessionDetails from '@/components/SessionDetails.vue'
import Supply from '@/components/Supply.vue'
import UserSessions from '@/components/UserSessions.vue'
import { FeatureEnum } from '@/components/types/FeatureTypes'

Vue.use(VueRouter)

// All navigable routes for the Portal
const routes: Array<RouteConfig> = [
  {
    path: '/login',
    name: 'login',
    component: LoginModal
  },
  {
    path: '/map',
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
    path: '/get-access',
    name: 'get-access',
    component: GetAccessModal
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
      },
      {
        path: 'route-shader',
        name: 'shader',
        component: RouteShader
      }
    ]
  },
  {
    path: '/explore',
    name: 'explore',
    component: ExplorationWorkspace,
    children: [
      {
        path: 'notifications',
        name: 'notifications',
        component: Notifications
      },
      {
        path: 'analytics',
        name: 'analytics',
        component: Analytics
      },
      {
        path: 'billing',
        name: 'billing',
        component: Billing
      },
      {
        path: 'supply',
        name: 'supply',
        component: Supply
      }
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
  // TODO: Make sure all edge cases for illegal routing are caught here
  // TODO: Clean this up. Figure out a better way of handling user role and legal route relationships
  if (!store.getters.isAdmin && (to.name === 'supply')) {
    next('/map')
    return
  }
  if ((!store.getters.isAdmin && !store.getters.isOwner && (to.name === 'users' || to.name === 'game-config')) || to.name === 'undefined') {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  if (store.getters.isAnonymous && (to.name === 'user-sessions' || to.name === 'user-tool' || to.name === 'account-settings' || to.name === 'downloads')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (router.app.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  if (!store.getters.isSeller && (to.name === 'supply')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (router.app.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  if (!store.getters.hasAnalytics && (to.name === 'analytics')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (router.app.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  if (!store.getters.hasBilling && (to.name === 'billing')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (router.app.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  // TODO: Add in checks for different parts of the explore page with new roles TBD
  if (to.name === 'explore') {
    store.commit('UPDATE_CURRENT_PAGE', 'notifications')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/explore/notifications')
    return
  }
  if (to.name === 'settings') {
    store.commit('UPDATE_CURRENT_PAGE', 'account-settings')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/settings/account')
    return
  }
  // Email is verified
  if (to.query.message === 'Your email was verified. You can continue using the application.') {
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    // TODO: refreshToken returns a promise that should be used to optimize page loads. Look into how this effects routing
    Vue.prototype.$authService.refreshToken()
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    next('/map')
    return
  }
  // Close modal if open on map page
  if (to.name === 'session-details' && from.name === 'map') {
    router.app.$root.$emit('hideModal')
  }
  store.commit('UPDATE_CURRENT_PAGE', to.name)
  if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
    (window as any).Intercom('update')
  }
  next()
})

export default router
