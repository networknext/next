export default {
  state: {
    currentPage: 'map',
    ready: false
  },
  getters: {
    currentPage: (state: any) => state.currentPage,
    isReady: (state: any) => state.ready
  },
  actions: {
    updateCurrentPage ({ commit }: any, currentPage: string) {
      commit('UPDATE_CURRENT_PAGE', currentPage)
    },
    toggleReady ({ commit }: any, ready: string) {
      commit('TOGGLE_APP_READY', ready)
    }
  },
  mutations: {
    UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
      state.currentPage = currentPage
    },
    TOGGLE_APP_READY (state: any, ready: boolean) {
      state.ready = ready
    }
  }
}
