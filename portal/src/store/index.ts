import Vue from 'vue'
import Vuex from 'vuex'
import modules from './modules'

Vue.use(Vuex)

/**
 * Basic Vuex store leveraging modules
 */

export default new Vuex.Store({
  strict: true,
  modules
})
