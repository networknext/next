export default {
  state: {
    showTable: false
  },
  getters: {
    showTable: (state: any) => state.showTable
  },
  actions: {
    toggleTable ({ commit }: any, showTable: boolean) {
      commit('TOGGLE_SESSION_TABLE', showTable)
    }
  },
  mutations: {
    TOGGLE_SESSION_TABLE (state: any, showTable: boolean) {
      state.showTable = showTable
    }
  }
}
// NAMESPACE THESE
