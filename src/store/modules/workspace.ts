/**
 * Basic Vuex module specific to workspace state
 */

import { DateFilterType, Filter } from '@/components/types/FilterTypes'

/**
 * TODO: Namespace these
 */

export default {
  state: {
    currentPage: 'map',
    filter: {
      companyCode: '',
      dateRange: DateFilterType.CURRENT_MONTH
    },
    finishedTours: [],
    finishedSignUpTours: [],
    isTour: false,
    isSignUpTour: false,
    killLoops: false,
    sessionCountAlertRef: null,
    viewport: null
  },
  getters: {
    currentPage: (state: any) => state.currentPage,
    currentFilter: (state: any) => state.filter,
    isTour: (state: any) => state.isTour,
    isSignUpTour: (state: any) => state.isSignUpTour,
    killLoops: (state: any) => state.killLoops,
    finishedTours: (state: any) => state.finishedTours,
    finishedSignUpTours: (state: any) => state.finishedSignUpTours,
    sessionCountAlert: (state: any) => state.sessionCountAlertRef,
    currentViewport: (state: any) => state.viewport
  },
  actions: {
    updateCurrentPage ({ commit }: any, currentPage: string) {
      commit('UPDATE_CURRENT_PAGE', currentPage)
    },
    updateCurrentFilter ({ commit }: any, newFilter: Filter) {
      commit('UPDATE_CURRENT_FILTER', newFilter)
    },
    toggleIsTour ({ commit }: any, isTour: boolean) {
      commit('TOGGLE_IS_TOUR', isTour)
    },
    toggleKillLoops ({ commit }: any, killLoops: boolean) {
      commit('TOGGLE_KILL_LOOPS', killLoops)
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
    },
    updateCurrentViewport ({ commit }: any, viewport: any) {
      commit('UPDATE_CURRENT_VIEWPORT', viewport)
    }
  },
  mutations: {
    UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
      state.currentPage = currentPage
    },
    UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
      state.filter = newFilter
    },
    TOGGLE_IS_TOUR (state: any, isTour: boolean) {
      state.isTour = isTour
    },
    TOGGLE_IS_SIGN_UP_TOUR (state: any, isSignUpTour: boolean) {
      state.isSignUpTour = isSignUpTour
    },
    TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
      state.killLoops = killLoops
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
    },
    UPDATE_CURRENT_VIEWPORT (state: any, viewport: any) {
      state.viewport = viewport
    }
  }
}
