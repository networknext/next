import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { EMAIL_VERIFICATION_URL } from '@/components/types/Constants'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { Auth0DecodedHash } from 'auth0-js'
import { cloneDeep } from 'lodash'
import Vue from 'vue'

/**
 * Basic Vuex module specific to authorization/authentication
 */

/**
 * TODO: Namespace these
 */

const state = {
  userProfile: newDefaultProfile(),
  allBuyers: []
}

const getters = {
  idToken: (state: any) => state.userProfile.idToken,
  isAdmin: (state: any) => state.userProfile.roles.indexOf('Admin') !== -1,
  isOwner: (state: any, getters: any) => state.userProfile.roles.indexOf('Owner') !== -1 || getters.isAdmin,
  isExplorer: (state: any, getters: any) => state.userProfile.roles.indexOf('Explorer') !== -1 || getters.isAdmin,
  isAnonymous: (state: any, getters: any) => getters.idToken === '',
  isAnonymousPlus: (state: any, getters: any) => !getters.isAnonymous ? !state.userProfile.verified : false,
  isBuyer: (state: any) => (state.userProfile.pubKey !== ''),
  hasAnalytics: (state: any, getters: any) => ((state.userProfile.hasAnalytics && getters.isExplorer) || getters.isAdmin),
  hasBilling: (state: any, getters: any) => ((state.userProfile.hasBilling && getters.isExplorer) || getters.isAdmin),
  hasTrial: (state: any) => state.userProfile.hasTrial,
  isSeller: (state: any, getters: any) => (state.userProfile.seller || getters.isAdmin),
  userProfile: (state: any) => state.userProfile,
  allBuyers: (state: any) => state.allBuyers,
  registeredToCompany: (state: any) => (state.userProfile.companyCode !== '')
}

const actions = {
  processAuthChange ({ dispatch, getters }: any, authResult: Auth0DecodedHash) {
    const token = authResult.idToken || ''
    if (token === '') {
      return
    }

    const userProfile = cloneDeep(state.userProfile)
    const idTokenPayload = authResult.idTokenPayload
    const nnScope = idTokenPayload['https://networknext.com/userData'] || {}
    const roles: Array<any> = nnScope.roles || { roles: [] }
    const companyCode: string = nnScope.company_code || nnScope.customer_code || ''
    const newsletterConsent: boolean = nnScope.newsletter || false
    const verified: boolean = nnScope.verified || false

    userProfile.roles = roles
    userProfile.idToken = token
    userProfile.auth0ID = idTokenPayload.sub
    userProfile.companyCode = companyCode
    userProfile.newsletterConsent = newsletterConsent
    userProfile.verified = verified

    dispatch('updateUserProfile', userProfile)

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
        let allBuyers: Array<any> = []
        const userProfile = cloneDeep(state.userProfile)
        if (getters.registeredToCompany) {
          allBuyers = responses[2].buyers || []
          userProfile.pubKey = responses[1].game_config.public_key || ''
        } else {
          allBuyers = responses[1].buyers || []
        }

        const userInformation = responses[0].account

        userProfile.email = userInformation.email || ''
        userProfile.avatar = userInformation.avatar || ''

        userProfile.buyerID = userInformation.buyer_id || ''
        userProfile.seller = userInformation.seller || false
        userProfile.firstName = userInformation.first_name || ''
        userProfile.lastName = userInformation.last_name || ''
        userProfile.companyName = userInformation.company_name || ''
        userProfile.hasAnalytics = userInformation.analytics || false
        userProfile.hasBilling = userInformation.billing || false
        userProfile.hasTrial = userInformation.trial || true
        userProfile.signedTOS = userInformation.signed_tos || false
        userProfile.domains = responses[0].domains || []

        dispatch('updateUserProfile', userProfile)

        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
          (window as any).Intercom('boot', {
            api_base: process.env.VUE_APP_INTERCOM_BASE_API,
            app_id: process.env.VUE_APP_INTERCOM_ID,
            email: userProfile.email,
            user_id: userProfile.auth0ID,
            avatar: userProfile.avatar,
            unsubscribed_from_emails: newsletterConsent,
            company: companyCode
          })
        }

        dispatch('updateAllBuyers', allBuyers)
        const defaultFilter: Filter = {
          companyCode: (userProfile.buyerID === '' || getters.isAdmin) ? '' : userProfile.companyCode,
          dateRange: DateFilterType.CURRENT_MONTH
        }
        dispatch('updateCurrentFilter', defaultFilter)

        if (getters.isAdmin) {
          const isDemo = Vue.$cookies.get('isDemo')
          if (isDemo) {
            dispatch('toggleIsDemo', true)
          }
        }

        const query = window.location.search
        if (query.includes(EMAIL_VERIFICATION_URL)) {
          dispatch('toggleIsSignUpTour', true)
          if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
            setTimeout(() => {
              Vue.prototype.$gtag.event('Account verified', {
                event_category: 'Account Creation'
              })
            }, 5000)
          }
        }
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
