import './router/ComponentHooks'
import './assets/main.scss'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import { JsonRPCPlugin } from './plugins/jsonrpc'
import { AuthPlugin } from './plugins/auth'

import { Route, NavigationGuardNext } from 'vue-router'

/**
 * Main file responsible for mounting the App component,
 *  setting up all of the services,
 *  mounting the cypress instance if running e2e tests,
 *  initializing all vuex stores,
 *  initializing auth0 related functionality
 */

Vue.config.productionTip = false

const app: any = null
const win: any = window

store.dispatch('updateCurrentPage', router.currentRoute.name)
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  store.dispatch('updateCurrentPage', to.name)
  next()
})

const clientID = process.env.VUE_APP_AUTH0_CLIENTID
const domain = process.env.VUE_APP_AUTH0_DOMAIN

Vue.use(AuthPlugin, {
  domain: domain,
  clientID: clientID
})

Vue.use(JsonRPCPlugin)

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')

if (win.Cypress) {
  win.app = app
}
