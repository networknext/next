/**
 * Basic Vuex module specific to workspace state
 */

/**
 * TODO: Namespace these
 */

export default {
  state: {
    currentPage: 'map',
    filter: {
      companyCode: ''
    }
  },
  getters: {
    currentPage: (state: any) => state.currentPage,
    currentFilter: (state: any) => state.filter
  },
  actions: {
    updateCurrentPage ({ commit }: any, currentPage: string) {
      commit('UPDATE_CURRENT_PAGE', currentPage)
    },
    updateCurrentFilter ({ commit }: any, currentFilter: string) {
      commit('UPDATE_CURRENT_FILTER', currentFilter)
    }
  },
  mutations: {
    UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
      state.currentPage = currentPage
    },
    UPDATE_CURRENT_FILTER (state: any, currentFilter: any) {
      state.filter = currentFilter
    }
  }
}
