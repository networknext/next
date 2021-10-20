import Vue from 'vue'
import VueRouter, { RouteConfig, Route, NavigationGuardNext } from 'vue-router'
import store from '@/store'

import { FeatureEnum } from '@/components/types/FeatureTypes'

import AccountSettings from '@/components/AccountSettings.vue'
import Analytics from '@/components/Analytics.vue'
import Billing from '@/components/Billing.vue'
import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import ExplorationWorkspace from '@/workspaces/ExplorationWorkspace.vue'
import GameConfiguration from '@/components/GameConfiguration.vue'
import GetAccessModal from '@/components/GetAccessModal.vue'
import LoginModal from '@/components/LoginModal.vue'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import Notifications from '@/components/Notifications.vue'
import ResetPasswordModal from '@/components/ResetPasswordModal.vue'
import SessionDetails from '@/components/SessionDetails.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import Supply from '@/components/Supply.vue'
import UserManagement from '@/components/UserManagement.vue'
import UserSessions from '@/components/UserSessions.vue'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'

Vue.use(VueRouter)

// All navigable routes for the Portal
const routes: Array<RouteConfig> = [
  {
    path: '/downloads',
    name: 'downloads',
    component: DownloadsWorkspace
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
    path: '/get-access',
    name: 'get-access',
    component: GetAccessModal
  },
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
    path: '/password-reset',
    name: 'password-reset',
    component: ResetPasswordModal
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
      }
    ]
  },
  {
    path: '*',
    name: '404'
  }
]

const router = new VueRouter({
  mode: 'history',
  routes
})

const AnonymousRoutes = [
  'map',
  'sessions',
  'session-details',
  'session-tool',
  'get-access',
  'login'
]

const AnonymousPlusRoutes = [
  'map',
  'sessions',
  'session-details',
  'session-tool',
  'user-sessions',
  'user-tool',
  'get-access',
  'login'
]

const ViewerRoutes = [
  'map',
  'sessions',
  'session-details',
  'session-tool',
  'user-sessions',
  'user-tool',
  'downloads',
  'settings',
  'account-settings',
  'get-access',
  'login'
]

const OwnerRoutes = [
  'map',
  'sessions',
  'session-details',
  'session-tool',
  'user-sessions',
  'user-tool',
  'downloads',
  'settings',
  'account-settings',
  'config',
  'users',
  'explore',
  'notifications',
  'get-access',
  'login'
]

// Add or remove these to open up beta features
const BetaRoutes = [
  'billing',
  'supply',
  'analytics'
]

// Catch all for routes. This can be used for a lot of different things like separating anon portal from authorized portal etc
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  if (to.name === 'session-details' && from.name === 'map') {
  }
  // Email is verified - catch this event, refresh the user's token and go to the map
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

  // Anonymous filters
  if (store.getters.isAnonymous && AnonymousRoutes.indexOf(to.name || '') === -1) {
    store.commit('UPDATE_CURRENT_PAGE', '/map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }

  // AnonymousPlus filters
  if (store.getters.isAnonymousPlus && AnonymousPlusRoutes.indexOf(to.name || '') === -1) {
    store.commit('UPDATE_CURRENT_PAGE', '/map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }

  if (!store.getters.isAnonymous && !store.getters.isAnonymousPlus && !store.getters.isOwner && ViewerRoutes.indexOf(to.name || '') === -1) {
    store.commit('UPDATE_CURRENT_PAGE', '/map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }

  // Owner Filters
  if (store.getters.Owner && OwnerRoutes.indexOf(to.name || '') === -1) {
    store.commit('UPDATE_CURRENT_PAGE', '/map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }

  // If user isn't an admin and they are trying to access beta content block them
  if (!store.getters.isAdmin && BetaRoutes.indexOf(to.name || '') !== -1) {
    store.commit('UPDATE_CURRENT_PAGE', '/map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }

  // Beta / Premium features given to the user at a buyer level
  if (!store.getters.isSeller && (to.name === 'supply')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  if (!store.getters.isAdmin && !store.getters.hasAnalytics && (to.name === 'analytics')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }
  if (!store.getters.isAdmin && !store.getters.hasBilling && (to.name === 'billing')) {
    store.commit('UPDATE_CURRENT_PAGE', 'map')
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('update')
    }
    next('/map')
    return
  }

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

  // Close modal if open on map page
  if (to.name === 'session-details' && from.name === 'map') {
    router.app.$root.$emit('hideMapPointsModal')
  }

  store.commit('UPDATE_CURRENT_PAGE', to.name)
  if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
    (window as any).Intercom('update')
  }
  next()
})

export default router
