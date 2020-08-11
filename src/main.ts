import './router/ComponentHooks'
import './assets/main.scss'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import APIService from './services/api.service'
import AuthService, { UserProfile, NNAuth0Profile } from './services/auth.service'
import { Route, NavigationGuardNext } from 'vue-router'
import { CreateElement } from 'vue/types/umd'

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
        const email = profile.email || ''
        const domain = email.split('@')[1]
        const token = authResult.idToken
        const userProfile: UserProfile = {
          auth0ID: profile.sub,
          company: '',
          email: profile.email || '',
          idToken: token,
          name: profile.name,
          roles: roles.roles,
          verified: profile.email_verified || false,
          routeShader: null,
          domain: domain,
          pubKey: '',
          buyerID: ''
        }
        const promises = [
          apiService.fetchUserAccount({ user_id: userProfile.auth0ID }, token),
          apiService.fetchGameConfiguration({ domain: domain }, token),
          apiService.fetchAllBuyers(token)
        ]
        Promise.all(promises)
          .then((responses: any) => {
            userProfile.buyerID = responses[0].account.buyer_id
            userProfile.company = responses[1].game_config.company
            userProfile.pubKey = responses[1].game_config.public_key
            userProfile.routeShader = responses[1].customer_route_shader
            const allBuyers = responses[2].buyers || []
            store.commit('UPDATE_USER_PROFILE', userProfile)
            store.commit('UPDATE_ALL_BUYERS', allBuyers)
            app = new Vue({
              router,
              store,
              render: (h: CreateElement) => h(App)
            }).$mount('#app')
            if (win.Cypress) {
              mountCypress(win, app)
            }
          })
          .catch((error: Error) => {
            console.log('Something went wrong fetching init data')
            console.log(error)
          })
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
      render: (h: CreateElement) => h(App)
    }).$mount('#app')
    if (win.Cypress) {
      mountCypress(win, app)
    }
  }
})
