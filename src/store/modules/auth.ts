import { UserProfile } from '@/components/types/AuthTypes'
import { cloneDeep } from 'lodash'
import Vue from 'vue'

/**
 * Basic Vuex module specific to authorization/authentication
 */

/**
 * TODO: Namespace these
 */

const defaultProfile: UserProfile = {
  auth0ID: '',
  companyCode: '',
  companyName: '',
  buyerID: '',
  seller: false,
  domains: [],
  firstName: '',
  lastName: '',
  email: '',
  idToken: '',
  verified: false,
  routeShader: null,
  pubKey: '',
  newsletterConsent: false,
  roles: []
}

const state = {
  userProfile: defaultProfile,
  allBuyers: []
}

const getters = {
  idToken: (state: any) => state.userProfile.idToken,
  isAdmin: (state: any) => state.userProfile.roles.indexOf('Admin') !== -1,
  isOwner: (state: any) => state.userProfile.roles.indexOf('Owner') !== -1,
  isViewer: (state: any) => state.userProfile.roles.indexOf('Viewer') !== -1,
  isAnonymous: (state: any, getters: any) => getters.idToken === '',
  isAnonymousPlus: (state: any, getters: any) => !getters.isAnonymous ? !state.userProfile.verified : false,
  isBuyer: (state: any) => (state.userProfile.pubKey !== ''),
  isSeller: (state: any) => (state.userProfile.seller),
  userProfile: (state: any) => state.userProfile,
  allBuyers: (state: any) => state.allBuyers,
  registeredToCompany: (state: any) => (state.userProfile.companyCode !== '')
}

const actions = {
  processAuthChange ({ dispatch, getters }: any) {
    if (getters.isAnonymous) {
      return
    }
    const userProfile = cloneDeep(state.userProfile)
    let promises = []
    if (getters.registeredToCompany) {
      promises = [
        Vue.prototype.$apiService.fetchUserAccount({ user_id: userProfile.auth0ID }),
        Vue.prototype.$apiService.fetchGameConfiguration(),
        Vue.prototype.$apiService.fetchAllBuyers()
      ]
    } else {
      promises = [
        Vue.prototype.$apiService.fetchUserAccount({ user_id: userProfile.auth0ID }),
        Vue.prototype.$apiService.fetchAllBuyers()
      ]
    }
    return Promise.all(promises)
      .then((responses: any) => {
        let allBuyers = []
        if (getters.registeredToCompany) {
          allBuyers = responses[2].buyers
          userProfile.pubKey = responses[1].game_config.public_key
        } else {
          allBuyers = responses[1].buyers
        }
        userProfile.buyerID = responses[0].account.id || responses[0].account.buyer_id || '' // TODO: remove the ".id" case after deploy
        userProfile.seller = responses[0].account.seller || false
        userProfile.firstName = responses[0].account.first_name || ''
        userProfile.lastName = responses[0].account.last_name || ''
        userProfile.companyName = responses[0].account.company_name || ''
        userProfile.domains = responses[0].domains || []
        dispatch('updateUserProfile', userProfile)
        dispatch('updateAllBuyers', allBuyers)
        dispatch('updateCurrentFilter', { companyCode: (userProfile.buyerID === '' || getters.isAdmin) ? '' : userProfile.companyCode })
      })
      .catch((error: Error) => {
        console.log('Something went wrong fetching user details')
        console.log(error.message)

        // The portal is basically useless with out this info so don't bother trying to load any other data
        dispatch('toggleKillLoops', true)
      })
  },
  updateUserProfile ({ commit }: any, userProfile: any) {
    commit('UPDATE_USER_PROFILE', userProfile)
  },
  updateAllBuyers ({ commit }: any, allBuyers: string) {
    commit('UPDATE_ALL_BUYERS', allBuyers)
  }
}

const mutations = {
  UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
    state.userProfile = userProfile
  },
  UPDATE_ALL_BUYERS (state: any, allBuyers: any) {
    state.allBuyers = allBuyers
  }
}

export default {
  state: state,
  getters: getters,
  actions: actions,
  mutations: mutations
}
