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

/**
 * Main file responsible for mounting the App component,
 *  setting up all of the services,
 *  mounting the cypress instance if running e2e tests,
 *  initializing all vuex stores,
 *  initializing auth0 related functionality
 */

Vue.config.productionTip = false

const clientID = process.env.VUE_APP_AUTH0_CLIENTID
const domain = process.env.VUE_APP_AUTH0_DOMAIN

Vue.use(AuthPlugin, {
  domain: domain,
  clientID: clientID
})

Vue.use(JSONRPCPlugin)
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

if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
  Vue.use(VueGtag, {
    config: { id: 'UA-141272717-2' }
  }, router)
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
