import './router/ComponentHooks'
import './assets/main.scss'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import { JSONRPCPlugin } from './plugins/jsonrpc'
import { AuthPlugin, AuthService } from './plugins/auth'
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

Vue.prototype.$authService.processAuthentication()
  .then(() => {
    const query = window.location.search
    if (query.includes('redirectURI')) {
      const parseURI = query.split('=')[1]
      router.push({ path: parseURI }).catch(() => { console.log('Routing to invoice failed') })
    } else if (window.location.hash !== '' || query.includes('signup')) {
      router.push('/map')
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
  .catch((err: Error) => {
    console.log('Something went wrong processing login')
    console.log(err)
    Vue.prototype.$authService.logout()
  })
