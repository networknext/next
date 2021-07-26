import './router/ComponentHooks'
import './assets/main.scss'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import { JSONRPCPlugin } from './plugins/jsonrpc'
import { AuthPlugin } from './plugins/auth'
import VueGtag from 'vue-gtag'
import { FlagPlugin } from './plugins/flags'
import { FeatureEnum, Flag } from './components/types/FeatureTypes'
import VueTour from 'vue-tour'

/**
 * Main file responsible for mounting the App component,
 *  setting up all of the services,
 *  mounting the cypress instance if running e2e tests,
 *  initializing all vuex stores,
 *  initializing auth0 related functionality
 */

Vue.config.productionTip = false

const flags: Array<Flag> = [
  {
    name: FeatureEnum.FEATURE_EXPLORE,
    description: 'Integrate Looker into the portal under a new navigation tab called "Explore"',
    value: false
  },
  {
    name: FeatureEnum.FEATURE_INTERCOM,
    description: 'Integrate intercom',
    value: false
  },
  {
    name: FeatureEnum.FEATURE_ROUTE_SHADER,
    description: 'Route shader page for users to update their route shader',
    value: false
  },
  {
    name: FeatureEnum.FEATURE_IMPERSONATION,
    description: 'Feature to allow admins to impersonate a customer in a read only state',
    value: false
  },
  {
    name: FeatureEnum.FEATURE_ANALYTICS,
    description: 'Google analytics and tag manager hooks',
    value: false
  },
  {
    name: FeatureEnum.FEATURE_TOUR,
    description: 'New product tour to replace intercom',
    value: false
  }
]

const useAPI = process.env.VUE_APP_USE_API_FLAGS === 'true'
Vue.use(FlagPlugin, {
  flags: flags,
  useAPI: useAPI,
  apiService: Vue.prototype.$apiService
})

if (useAPI) {
  Vue.prototype.$flagService.fetchAllRemoteFeatureFlags()
} else {
  Vue.prototype.$flagService.fetchEnvVarFeatureFlags()
}

require('vue-tour/dist/vue-tour.css')

Vue.use(VueTour)

const gtagID = process.env.VUE_APP_GTAG_ID || ''

if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS) && gtagID !== '') {
  Vue.use(VueGtag, {
    config: { id: gtagID }
  }, router)
}

const clientID = process.env.VUE_APP_AUTH0_CLIENTID
const domain = process.env.VUE_APP_AUTH0_DOMAIN

Vue.use(AuthPlugin, {
  domain: domain,
  clientID: clientID
})

Vue.use(JSONRPCPlugin)

// This is VERY hacky. It would be much better to do this within the router but going that route (no pun intended) mounts half the app before hitting the redirect which is funky
// TODO: Look into a lifecycle hook that handles this better...
if (window.location.pathname === '/get-access') {
  Vue.prototype.$authService.signUp(window.location.search.split('?email=')[1])
} else {
  Vue.prototype.$authService.processAuthentication().then(() => {
    const query = window.location.search
    if (query.includes('Your%20email%20was%20verified.%20You%20can%20continue%20using%20the%20application.')) {
      store.commit('TOGGLE_IS_SIGN_UP_TOUR', true)
      if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
        setTimeout(() => {
          Vue.prototype.$gtag.event('Account verified', {
            event_category: 'Account Creation'
          })
        }, 5000)
      }
    }

    if (query.includes('signup')) {
      setTimeout(() => {
        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          Vue.prototype.$gtag.event('Auth0 account created', {
            event_category: 'Account Creation'
          })
        }
        Vue.prototype.$apiService.sendSignUpSlackNotification({ email: store.getters.userProfile.email })
      }, 5000)
    }

    if (query.includes('code=') && query.includes('state=')) {
      router.push('/map')
    }

    const isReturning = localStorage.returningUser || 'false'
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_TOUR)) {
      if (!(isReturning === 'true') && store.getters.isAnonymous) {
        store.commit('TOGGLE_IS_TOUR', true)
        localStorage.returningUser = true
      }
    }
    const app = new Vue({
      router,
      store,
      render: h => h(App)
    }).$mount('#app')

    const win: any = window

    if (win.Cypress) {
      win.app = app
    }
  })
}
