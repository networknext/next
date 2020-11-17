import { UserProfile } from '@/components/types/AuthTypes.ts'

/**
 * Basic Vuex module specific to authorization/authentication
 */

/**
 * TODO: Namespace these
 */

const defaultProfile = {
  auth0ID: '',
  companyCode: '',
  companyName: '',
  buyerID: '',
  domains: [],
  email: '',
  idToken: '',
  name: '',
  verified: false,
  routeShader: null,
  pubKey: '',
  newsletterConsent: false,
  roles: []
} as UserProfile

export default {
  state: {
    userProfile: defaultProfile,
    allBuyers: [],
    abTesters: [
      '2b9c891211588152',
      'b8e4f84ca63b2021',
      '02a337e6ec5b60b5'
    ]
  },
  getters: {
    idToken: (state: any) => state.userProfile.idToken,
    isAdmin: (state: any) => state.userProfile.roles.indexOf('Admin') !== -1,
    isOwner: (state: any) => state.userProfile.roles.indexOf('Owner') !== -1,
    isViewer: (state: any) => state.userProfile.roles.indexOf('Viewer') !== -1,
    isAnonymous: (state: any, getters: any) => getters.idToken === '',
    isAnonymousPlus: (state: any, getters: any) => !getters.isAnonymous ? !state.userProfile.verified : false,
    isABTester: (state: any, getters: any) => (getters.isBuyer && state.abTesters.includes(getters.userProfile.id)) || getters.isAdmin,
    isBuyer: (state: any) => (state.userProfile.pubKey !== ''),
    userProfile: (state: any) => state.userProfile,
    allBuyers: (state: any) => state.allBuyers,
    registeredToCompany: (state: any) => (state.userProfile.companyCode !== '')
  },
  actions: {
    updateUserProfile ({ commit }: any, userProfile: any) {
      commit('UPDATE_USER_PROFILE', userProfile)
    },
    updateAllBuyers ({ commit }: any, allBuyers: string) {
      commit('UPDATE_ALL_BUYERS', allBuyers)
    }
  },
  mutations: {
    UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
      state.userProfile = userProfile
    },
    UPDATE_ALL_BUYERS (state: any, allBuyers: any) {
      state.allBuyers = allBuyers
    }
  }
}
