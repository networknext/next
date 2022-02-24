import './router/ComponentHooks'
import './assets/main.scss'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'uplot/dist/uPlot.min.css'
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

// Setup feature flags - This should be moved to a database so that we can flip features on and off on the fly - potentially do A/B testing
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
  },
  {
    name: FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT,
    description: 'Leverage Looker API for user tool and session tool',
    value: false
  }
]

// Check to see if the flags should come from the database or not (future feature)
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

// Setup vue tour - This handles all of our sign up and new user tours
require('vue-tour/dist/vue-tour.css')
Vue.use(VueTour)

// Setup up google tag manager for analytics
const gtagID = process.env.VUE_APP_GTAG_ID || ''
if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS) && gtagID !== '') {
  Vue.use(VueGtag, {
    config: { id: gtagID }
  }, router)
}

// Setup auth0 plugin - This manages all of the auth related functions of the Portal (login, signup, auth related state initialization)
const clientID = process.env.VUE_APP_AUTH0_CLIENTID
const domain = process.env.VUE_APP_AUTH0_DOMAIN
Vue.use(AuthPlugin, {
  domain: domain,
  clientID: clientID,
  store: store,
  flagService: Vue.prototype.$flagService,
  gtagService: Vue.prototype.$gtag
})

// Setup jsonrpc plugin - This handles all communication with the backend both auth'd and unauth'd
Vue.use(JSONRPCPlugin)

// Process authorization - User accessing webpage is either logged in or not and this function handles the state initialization for those two scenarios
Vue.prototype.$authService.processAuthentication()
  .then(() => {
    // Check to see if the user should be redirected after login or send them to the map if not
    const query = window.location.search
    if (query.includes('redirectURI')) {
      const parseURI = query.split('=')[1]

      // Push the user to the page they requested before logging in - This may fail if the permissions of the user are not sufficient for the route
      router.push({ path: parseURI }).catch(() => { /** Catch insufficient privileges error here. May want to hook this to a tracker of some kind or something in the future */ })
    } else if (window.location.hash !== '' || query.includes('signup')) {
      router.push('/map')
    }

    // Mount the application
    const app = new Vue({
      router,
      store,
      render: h => h(App)
    }).$mount('#app')

    // Add application access if this is an e2e test (manually set up state, etc)
    const win: any = window
    if (win.Cypress) {
      win.app = app
    }
  })
  .catch((err: Error) => {
    // If there is anything wrong with the login process, wipe any auth data with a logout and send them back to the anonymous map
    console.log('Something went wrong processing login')
    console.log(err)
    Vue.prototype.$authService.logout()
  })
