import { UserProfile } from '@/services/auth.service'

export default {
  state: {
    userProfile: null,
    allBuyers: []
  },
  getters: {
    idToken: (state: any) => state.userProfile ? state.userProfile.idToken : null,
    isAdmin: (state: any) => state.userProfile ? state.userProfile.roles.indexOf('Admin') !== -1 : false,
    isOwner: (state: any) => state.userProfile ? state.userProfile.roles.indexOf('Owner') !== -1 : false,
    isViewer: (state: any) => state.userProfile ? state.userProfile.roles.indexOf('Viewer') !== -1 : false,
    isAnonymous: (state: any) => state.userProfile === null,
    isAnonymousPlus: (state: any, getters: any) => state.userProfile ? state.userProfile.verified && !getters.isAdmin && !getters.isOwner && !getters.isViewer : false,
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
