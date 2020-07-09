import './router/ComponentHooks'
import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'
import APIService from './services/api.service'

Vue.config.productionTip = false

// Add api service as a Vue property so it can be used in all Vue components
Vue.prototype.$apiService = new APIService()

new Vue({
  router,
  store,
  render: (h) => h(App)
}).$mount('#app')
