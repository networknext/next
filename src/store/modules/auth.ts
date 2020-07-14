import { UserProfile } from '@/services/auth.service'

export default {
  state: {
    userProfile: null
  },
  getters: {
    idToken: (state: any) => state.userProfile ? state.userProfile.idToken : null,
    isAdmin: (state: any) => state.userProfile ? state.userProfile.roles.find((role: string) => { return role === 'Admin' }) : false,
    isOwner: (state: any) => state.userProfile ? state.userProfile.roles.find((role: string) => { return role === 'Owner' }) : false,
    isAnonymous: (state: any) => state.userProfile === null,
    isAnonymousPlus: (state: any) => state.userProfile ? state.userProfile.verified : false,
    userProfile: (state: any) => state.userProfile
  },
  actions: {
    updateUserProfile ({ commit }: any, currentPage: string) {
      commit('UPDATE_USER_PROFILE', currentPage)
    }
  },
  mutations: {
    UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
      state.userProfile = userProfile
    }
  }
}
