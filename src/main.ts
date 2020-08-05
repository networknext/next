import './router/ComponentHooks'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import APIService from './services/api.service'
import AuthService, { UserProfile, NNAuth0Profile } from './services/auth.service'
import { Route, NavigationGuardNext } from 'vue-router'

function mountCypress (win: any, app: any) {
  win.app = app
}

Vue.config.productionTip = false

// Add api service as a Vue property so it can be used in all Vue components
const apiService = new APIService()
Vue.prototype.$apiService = apiService

// Add auth service as a Vue property so it can be used in all Vue components
const authService = new AuthService()
Vue.prototype.$authService = authService

let app: any = null
const win: any = window

store.dispatch('updateCurrentPage', router.currentRoute.name)
router.beforeEach((to: Route, from: Route, next: NavigationGuardNext<Vue>) => {
  store.dispatch('updateCurrentPage', to.name)
  next()
})
authService.lockClient.checkSession({
  scope: 'openid profile email user_metadata app_metadata'
}, (error: auth0.Auth0Error, authResult: AuthResult | undefined) => {
  if (!error && authResult) {
    authService.lockClient.getUserInfo(authResult.accessToken, (error: auth0.Auth0Error, profile: NNAuth0Profile) => {
      if (!error) {
        const roles = profile['https://networknext.com/userRoles'] || { roles: [] }
        const userProfile: UserProfile = {
          auth0ID: profile.sub,
          company: '',
          email: profile.email || '',
          idToken: authResult.idToken,
          name: profile.name,
          roles: roles.roles,
          verified: profile.email_verified || false,
          routeShader: null
        }
        store.commit('UPDATE_USER_PROFILE', userProfile)
        apiService.fetchAllBuyers().then((response: any) => {
          const allBuyers = response.buyers || []
          store.commit('UPDATE_ALL_BUYERS', allBuyers)
        })
        app = new Vue({
          router,
          store,
          render: (h) => h(App)
        }).$mount('#app')
        if (win.Cypress) {
          mountCypress(win, app)
        }
      }
    })
  } else {
    // TODO: Come up with a way to handle errors better
    if (error.error !== 'login_required') {
      console.log(error)
    }
    app = new Vue({
      router,
      store,
      render: (h) => h(App)
    }).$mount('#app')
    if (win.Cypress) {
      mountCypress(win, app)
    }
  }
})
