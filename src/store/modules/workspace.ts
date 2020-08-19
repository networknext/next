/**
 * Basic Vuex module specific to workspace state
 */

/**
 * TODO: Namespace these
 */

export default {
  state: {
    currentPage: 'map'
  },
  getters: {
    currentPage: (state: any) => state.currentPage
  },
  actions: {
    updateCurrentPage ({ commit }: any, currentPage: string) {
      commit('UPDATE_CURRENT_PAGE', currentPage)
    }
  },
  mutations: {
    UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
      state.currentPage = currentPage
    }
  }
}
