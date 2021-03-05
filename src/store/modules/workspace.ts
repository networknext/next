/**
 * Basic Vuex module specific to workspace state
 */

/**
 * TODO: Namespace these
 */

export default {
  state: {
    isTour: false,
    currentTourStep: -1,
    finishedTours: [],
    currentPage: 'map',
    filter: {
      companyCode: ''
    }
  },
  getters: {
    currentPage: (state: any) => state.currentPage,
    currentFilter: (state: any) => state.filter,
    isTour: (state: any) => state.isTour,
    currentStep: (state: any) => state.currentTourStep,
    finishedTours: (state: any) => state.finishedTours
  },
  actions: {
    updateCurrentPage ({ commit }: any, currentPage: string) {
      commit('UPDATE_CURRENT_PAGE', currentPage)
    },
    updateCurrentFilter ({ commit }: any, currentFilter: string) {
      commit('UPDATE_CURRENT_FILTER', currentFilter)
    },
    toggleIsTour ({ commit }: any, isTour: boolean) {
      commit('TOGGLE_IS_TOUR', isTour)
    },
    updateCurrentTourStep ({ commit }: any, currentTourStep: number) {
      commit('UPDATE_CURRENT_TOUR_STEP', currentTourStep)
    },
    updateFinishedTours ({ commit }: any, finishedTour: string) {
      commit('UPDATE_FINISHED_TOURS', finishedTour)
    }
  },
  mutations: {
    UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
      state.currentPage = currentPage
    },
    UPDATE_CURRENT_FILTER (state: any, currentFilter: any) {
      state.filter = currentFilter
    },
    TOGGLE_IS_TOUR (state: any, isTour: boolean) {
      state.isTour = isTour
    },
    UPDATE_CURRENT_TOUR_STEP (state: any, currentTourStep: number) {
      state.currentTourStep = currentTourStep
    },
    UPDATE_FINISHED_TOURS (state: any, finishedTour: string) {
      if (!state.finishedTours.includes(finishedTour)) {
        state.finishedTours.push(finishedTour)
      }
    }
  }
}
