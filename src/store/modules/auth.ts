import { UserProfile } from '@/services/auth.service'
export default {
  state: {
    userProfile: null,
    allBuyers: [],
    abTesters: [
      '2b9c891211588152',
      'b8e4f84ca63b2021',
      '02a337e6ec5b60b5'
    ]
  },
  getters: {
    idToken: (state: any) => state.userProfile ? state.userProfile.idToken : null,
    isAdmin: (state: any) => state.userProfile ? state.userProfile.roles.indexOf('Admin') !== -1 : false,
    isOwner: (state: any) => state.userProfile ? state.userProfile.roles.indexOf('Owner') !== -1 : false,
    isViewer: (state: any) => state.userProfile ? state.userProfile.roles.indexOf('Viewer') !== -1 : false,
    isAnonymous: (state: any) => state.userProfile === null,
    isAnonymousPlus: (state: any, getters: any) => !getters.isAnonymous ? !state.userProfile.verified : false,
    isABTester: (state: any, getters: any) => (getters.isBuyer && state.abTesters.includes(getters.userProfile.id)) || getters.isAdmin,
    isBuyer: (state: any) => (state.userProfile ? state.userProfile.buyerID !== '' : false),
    userProfile: (state: any) => state.userProfile,
    allBuyers: (state: any) => state.allBuyers
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
