import Vue from 'vue'
import VueRouter, { RouteConfig, Route, NavigationGuardNext } from 'vue-router'
import store from '@/store'

import { FeatureEnum } from '@/components/types/FeatureTypes'

import AccountSettings from '@/components/AccountSettings.vue'
import Analytics from '@/components/Analytics.vue'
import Discovery from '@/components/Discovery.vue'
import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import ExplorationWorkspace from '@/workspaces/ExplorationWorkspace.vue'
import GameConfiguration from '@/components/GameConfiguration.vue'
import GetAccessModal from '@/components/GetAccessModal.vue'
import LoginModal from '@/components/LoginModal.vue'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import ResetPasswordModal from '@/components/ResetPasswordModal.vue'
import Saves from '@/components/Saves.vue'
import SessionDetails from '@/components/SessionDetails.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import Supply from '@/components/Supply.vue'
import Usage from '@/components/Usage.vue'
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
        path: 'analytics',
        name: 'analytics',
        component: Analytics
      },
      {
        path: 'usage',
        name: 'usage',
        component: Usage,
        children: [
          {
            path: '*',
            name: 'invoice'
          }
        ]
      },
      {
        path: 'discovery',
        name: 'discovery',
        component: Discovery
      },
      {
        path: 'saves',
        name: 'saves',
        component: Saves
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
  'login',
  'password-reset'
]

const AnonymousPlusRoutes = [
  'map',
  'sessions',
  'session-details',
  'session-tool',
  'user-sessions',
  'user-tool'
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
  'account-settings'
]

const ExplorerRoutes = [
  'map',
  'sessions',
  'session-details',
  'session-tool',
  'user-sessions',
  'user-tool',
  'downloads',
  'settings',
  'account-settings',
  'explore',
  'usage',
  'invoice',
  'analytics'
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
  'usage',
  'analytics'
]

// Add or remove these to open up beta features
const BetaRoutes = [
  'discovery',
  'saves',
  'supply'
]

function updateCurrentPage (name: string) {
  store.commit('UPDATE_CURRENT_PAGE', name)
  if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
    (window as any).Intercom('update')
  }
}

router.onError(() => {
  updateCurrentPage('map')
  router.push('/map')
})

// Catch all for routes. This can be used for a lot of different things like separating anon portal from authorized portal etc
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  if (to.name === '404') {
    next(new Error('Route does not exist'))
    return
  }
  // Email is verified - catch this event, refresh the user's token and go to the map
  if (to.query.message === 'Your email was verified. You can continue using the application.') {
    // TODO: refreshToken returns a promise that should be used to optimize page loads. Look into how this effects routing
    Vue.prototype.$authService.refreshToken()
    updateCurrentPage('map')
    next('/map')
    return
  }

  // Anonymous filters
  if (store.getters.isAnonymous && AnonymousRoutes.indexOf(to.name || '') === -1) {
    // Prompt user to login and try the route again afterwards
    updateCurrentPage('login')
    next('/login?redirectURI=' + to.fullPath)
    return
  }

  // AnonymousPlus filters
  if (store.getters.isAnonymousPlus && AnonymousPlusRoutes.indexOf(to.name || '') === -1) {
    next(new Error('Insufficient privileges'))
    return
  }

  if (!store.getters.isAnonymous && !store.getters.isAnonymousPlus && !store.getters.isExplorer && !store.getters.isOwner && !store.getters.isAdmin && ViewerRoutes.indexOf(to.name || '') === -1) {
    next(new Error('Insufficient privileges'))
    return
  }

  // Explorer Filters
  if (store.getters.isExplorer && !store.getters.isOwner && !store.getters.isAdmin && ExplorerRoutes.indexOf(to.name || '') === -1) {
    next(new Error('Insufficient privileges'))
    return
  }

  // Owner Filters
  if (store.getters.isOwner && !store.getters.isAdmin && OwnerRoutes.indexOf(to.name || '') === -1) {
    next(new Error('Insufficient privileges'))
    return
  }

  // If user isn't an admin and they are trying to access beta content block them
  if (!store.getters.isAdmin && BetaRoutes.indexOf(to.name || '') !== -1) {
    next(new Error('Insufficient privileges'))
    return
  }

  if (to.name === 'explore') {
    if (store.getters.hasBilling) {
      updateCurrentPage('usage')
      next('/explore/usage')
      return
    }

    if (store.getters.hasAnalytics) {
      updateCurrentPage('analytics')
      next('/explore/analytics')
      return
    }

    next(new Error('Insufficient privileges'))
    return
  }

  // Beta / Premium features given to the user at a buyer level
  if (!store.getters.hasBilling && (to.name === 'usage' || to.name === 'invoice')) {
    next(new Error('Insufficient privileges'))
    return
  }
  if (!store.getters.hasAnalytics && (to.name === 'analytics')) {
    next(new Error('Insufficient privileges'))
    return
  }
  if (!store.getters.isAdmin && !store.getters.hasAnalytics && (to.name === 'discovery')) {
    next(new Error('Insufficient privileges'))
    return
  }
  if (!store.getters.isAdmin && !store.getters.isSeller && (to.name === 'supply')) {
    next(new Error('Insufficient privileges'))
    return
  }

  if (to.name === 'settings') {
    updateCurrentPage('account-settings')
    next('/settings/account')
    return
  }

  // Close modal if open on map page
  if (to.name === 'session-details' && from.name === 'map') {
    router.app.$root.$emit('hideMapPointsModal')
  }

  updateCurrentPage(to.name || '')
  next()
})

export default router
