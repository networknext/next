/**
 * Basic Vuex module specific to workspace state
 */

/**
 * TODO: Namespace these
 */

export default {
  state: {
    isTour: false,
    isSignUpTour: false,
    finishedTours: [],
    finishedSignUpTours: [],
    currentPage: 'map',
    filter: {
      companyCode: ''
    },
    sessionCountAlertRef: null
  },
  getters: {
    currentPage: (state: any) => state.currentPage,
    currentFilter: (state: any) => state.filter,
    isTour: (state: any) => state.isTour,
    isSignUpTour: (state: any) => state.isSignUpTour,
    finishedTours: (state: any) => state.finishedTours,
    finishedSignUpTours: (state: any) => state.finishedSignUpTours,
    sessionCountAlert: (state: any) => state.sessionCountAlertRef
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
    updateFinishedTours ({ commit }: any, finishedTour: string) {
      commit('UPDATE_FINISHED_TOURS', finishedTour)
    },
    toggleIsSignUpTour ({ commit }: any, isSignUpTour: boolean) {
      commit('TOGGLE_IS_SIGN_UP_TOUR', isSignUpTour)
    },
    updateFinishedSignUpTours ({ commit }: any, finishedSignUpTour: string) {
      commit('UPDATE_FINISHED_SIGN_UP_TOURS', finishedSignUpTour)
    },
    setSessionCountAlertMessage ({ commit }: any, sessionCountAlertMessage: any) {
      commit('SET_SESSION_COUNT_ALERT_MESSAGE', sessionCountAlertMessage)
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
    TOGGLE_IS_SIGN_UP_TOUR (state: any, isSignUpTour: boolean) {
      state.isSignUpTour = isSignUpTour
    },
    UPDATE_FINISHED_TOURS (state: any, finishedTour: string) {
      if (!state.finishedTours.includes(finishedTour)) {
        state.finishedTours.push(finishedTour)
      }
    },
    UPDATE_FINISHED_SIGN_UP_TOURS (state: any, finishedSignUpTours: string) {
      if (!state.finishedSignUpTours.includes(finishedSignUpTours)) {
        state.finishedSignUpTours.push(finishedSignUpTours)
      }
    },
    SET_SESSION_COUNT_ALERT_MESSAGE (state: any, sessionCountAlertMessage: string) {
      state.sessionCountAlertMessage = sessionCountAlertMessage
    }
  }
}
