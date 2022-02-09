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
        path: 'discovery',
        name: 'discovery',
        component: Discovery
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
  'explore',
  'usage',
  'invoice',
  'analytics'
]

const OwnerRoutes = [
  'config',
  'users'
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

function checkMapModal (toName: string, fromName: string) {
  // Close modal if open on map page
  if (toName === 'session-details' && fromName === 'map') {
    router.app.$root.$emit('hideMapPointsModal')
  }
}

router.onError(() => {
  updateCurrentPage('map')
  router.push('/map')
})

// Catch all for routes. This can be used for a lot of different things like separating anon portal from authorized portal etc
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  const toName = to.name || ''
  const fromName = from.name || ''

  if (toName === '404') {
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

  if (store.getters.isAnonymous && AnonymousRoutes.indexOf(toName) === -1) {
    // Prompt user to login and try the route again afterwards
    updateCurrentPage('/login')
    next('/login?redirectURI=' + to.fullPath)
    return
  }

  // Anonymous filters
  if (store.getters.isAnonymous && AnonymousRoutes.indexOf(toName) !== -1) {
    checkMapModal(toName, fromName)

    updateCurrentPage(toName)
    next()
    return
  }

  // AnonymousPlus filters
  if (store.getters.isAnonymousPlus && AnonymousPlusRoutes.indexOf(toName) !== -1) {
    checkMapModal(toName, fromName)

    updateCurrentPage(toName)
    next()
    return
  }

  // Viewer filters (User that is setup and verified but doesn't have a company and/or any roles)
  if (!(store.getters.isAnonymous || store.getters.isAnonymousPlus) && ViewerRoutes.indexOf(toName) !== -1) {
    checkMapModal(toName, fromName)

    if (toName === 'settings') {
      updateCurrentPage('account-settings')
      next('/settings/account')
      return
    }

    updateCurrentPage(toName)
    next()
    return
  }

  // Explorer Filters
  if (store.getters.isExplorer && ExplorerRoutes.indexOf(toName) !== -1) {
    if (toName === 'explore') {
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
    } else {
      updateCurrentPage(toName)
      next()
      return
    }

    next(new Error('Insufficient privileges'))
    return
  }

  // Owner Filters
  if (store.getters.isOwner && OwnerRoutes.indexOf(toName) !== -1) {
    updateCurrentPage(toName)
    next()
    return
  }

  // If user isn't an admin and they are trying to access beta content block them
  if (store.getters.isAdmin && BetaRoutes.indexOf(toName) !== -1) {
    updateCurrentPage(toName)
    next()
    return
  }

  next(new Error('Insufficient privileges'))
})

export default router
