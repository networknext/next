import './router/ComponentHooks'
import './assets/main.scss'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import APIService from './services/api.service'
// import AuthService, { UserProfile, NNAuth0Profile } from './services/auth.service'
import { AuthPlugin } from './plugins/auth'
// import { Auth0Plugin } from './plugins/auth0'

import { Route, NavigationGuardNext } from 'vue-router'
// import { CreateElement } from 'vue/types/umd'

/**
 * Main file responsible for mounting the App component,
 *  setting up all of the services,
 *  mounting the cypress instance if running e2e tests,
 *  initializing all vuex stores,
 *  initializing auth0 related functionality
 */

/**
 * TODO: Clean all of this up a bit
 * TODO: Potentially find a better auth solution
 */
// function mountCypress (win: any, app: any) {
//   win.app = app
// }

Vue.config.productionTip = false

// Add api service as a Vue property so it can be used in all Vue components
const apiService = new APIService()
Vue.prototype.$apiService = apiService

// Add auth service as a Vue property so it can be used in all Vue components
// const authService = new AuthService()
// Vue.prototype.$authService = authService

const app: any = null
const win: any = window

store.dispatch('updateCurrentPage', router.currentRoute.name)
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  store.dispatch('updateCurrentPage', to.name)
  next()
})

const clientID = 'Kx0mbNIMZtMNA71vf9iatCp3N6qi1GfL'
const domain = 'networknext.auth0.com'

Vue.use(AuthPlugin, {
  domain: domain,
  clientID: clientID
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
