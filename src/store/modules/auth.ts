import { UserProfile } from '@/components/types/APITypes'

export default {
  state: {
    userProfile: null
  },
  getters: {
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
