<template>
  <div class="container-fluid below-nav-bar">
    <div class="row">
      <main role="main" class="col-md-12 col-lg-12 px-4">
        <SessionCounts
          v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'"
        />
        <router-view />
        <MapPointsModal v-show="showMapPointsModal" :points="modalPoints" />
        <NotificationsModal v-show="($store.getters.isOwner || $store.getters.isAdmin) && showNotificationsModal"/>
        <TermsOfServiceModal ref="tosModal" :deniable="this.$store.getters.userProfile.buyerID === ''" />
      </main>
      <v-tour v-show="$store.getters.currentPage === 'map'" name="mapTour" :steps="mapTourSteps" :options="mapTourOptions" :callbacks="mapTourCallbacks"></v-tour>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import LoginModal from '@/components/LoginModal.vue'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import NotificationsModal from '@/components/NotificationsModal.vue'
import SessionCounts from '@/components/SessionCounts.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import TermsOfServiceModal from '@/components/modals/TermsOfServiceModal.vue'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import MapPointsModal from '@/components/MapPointsModal.vue'

/**
 * This component is the base component for all other workspace components
 * It also holds the Email Verification alert and Session Count components
 *  so that they are visible across all workspaces if necessary
 */

@Component({
  components: {
    LoginModal,
    MapPointsModal,
    MapWorkspace,
    NotificationsModal,
    SessionCounts,
    SessionsWorkspace,
    SessionToolWorkspace,
    SettingsWorkspace,
    TermsOfServiceModal
  }
})
export default class Workspace extends Vue {
  private unwatchUserProfile: any
  private mapTourSteps: Array<any>
  private mapTourOptions: any
  private mapTourCallbacks: any
  private showMapPointsModal: boolean
  private showNotificationsModal: boolean
  private showTOSModal: boolean
  private modalPoints: Array<any>

  $refs!: {
    tosModal: TermsOfServiceModal;
  }

  constructor () {
    super()
    this.mapTourSteps = [
      {
        target: '#map',
        header: {
          title: 'Map'
        },
        content: 'Each green dot on the <strong>Map </strong>is a player being accelerated by Network Next.',
        params: {
          enabledButtons: {
            buttonSkip: false,
            buttonPrevious: false,
            buttonNext: true,
            buttonStop: false
          },
          placement: 'bottom'
        }
      },
      {
        target: '[data-tour="sessionsLink"]',
        header: {
          title: 'Sessions'
        },
        content: 'Click <strong>Sessions</strong> to learn more about what Network Next does!',
        params: {
          enabledButtons: {
            buttonSkip: false,
            buttonPrevious: false,
            buttonNext: true,
            buttonStop: true
          }
        }
      }
    ]

    this.mapTourOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'OK'
      }
    }

    this.mapTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_TOURS', 'map')

        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          Vue.prototype.$gtag.event('Map tour finished', {
            event_category: 'Tours'
          })
        }
      }
    }

    this.showMapPointsModal = false
    this.showNotificationsModal = false
    this.showTOSModal = false
    this.modalPoints = []
  }

  private mounted () {
    this.checkTOSRequired()

    this.unwatchUserProfile = this.$store.watch(
      (state: any, getters: any) => {
        return getters.userProfile
      },
      () => {
        this.checkTOSRequired()
      }
    )

    if (this.$store.getters.isTour && this.$route.name === 'map' && this.$tours.mapTour && !this.$tours.mapTour.isRunning) {
      this.$tours.mapTour.start()
    }

    // TODO: Make a modal events bus rather than using the root application bus
    this.$root.$on('showMapPointsModal', this.showMapPointsModalCallback)
    this.$root.$on('hideMapPointsModal', this.hideMapPointsModalCallback)

    this.$root.$on('showNotificationsModal', this.showNotificationsModalCallback)
    this.$root.$on('hideNotificationsModal', this.hideNotificationsModalCallback)

    this.$root.$on('showTOSModal', this.showTOSModalCallback)
  }

  private beforeDestroy () {
    this.unwatchUserProfile()

    this.$root.$off('showMapPointsModal')
    this.$root.$off('hideMapPointsModal')

    this.$root.$off('showNotificationsModal')
    this.$root.$off('hideNotificationsModal')

    this.$root.$off('showTOSModal')
  }

  private showMapPointsModalCallback (points: Array<any>) {
    if (!this.showMapPointsModal) {
      this.modalPoints = points
      this.showMapPointsModal = true
    }
  }

  private hideMapPointsModalCallback () {
    if (this.showMapPointsModal) {
      this.showMapPointsModal = false
    }
  }

  private showNotificationsModalCallback () {
    if (!this.showNotificationsModal) {
      this.showNotificationsModal = true
    }
  }

  private hideNotificationsModalCallback () {
    if (this.showNotificationsModal) {
      this.showNotificationsModal = false
    }
  }

  private showTOSModalCallback () {
    if (!this.$refs.tosModal.isVisible()) {
      this.$refs.tosModal.toggleShowModal()
    }
  }

  private checkTOSRequired () {
    if (
      this.$store.getters.isOwner &&
      this.$store.getters.userProfile.buyerID !== '' &&
      !this.$store.getters.userProfile.signedTOS &&
      !this.$refs.tosModal.isVisible()
    ) {
      this.$refs.tosModal.toggleShowModal()
    } else {
      if (this.$refs.tosModal.isVisible()) {
        this.$refs.tosModal.toggleHideModal()
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style lang="scss">
  .v-step {
    background-color: white !important;
  }
  .v-step__header {
    padding: 1rem !important;
    padding-bottom: 28px !important;
    background-color: white !important;
    color: black;
    font-size: 24px;
    font-weight: bold;
    text-align: left;
    margin-bottom: 0 !important;
  }
  .v-step__content {
    background-color: white !important;
    padding-bottom: 28px !important;
    font-size: 14px;
    color: #1A1A1A;
    text-align: left;
    margin: 0 !important;
  }
  .v-step__button {
    border-color: #1A1A1A !important;
    color: #1A1A1A !important;
    font-size: 18px;
    font-weight: 400;
    min-width: 80px;
  }
  .v-step__button.v-step__button-next {
    background-color: #009FDF !important;
    border-color: rgb(0, 159, 223) !important;
    color: white !important;
  }
  .v-step__button.v-step__button-stop {
    background-color: #009FDF !important;
    border-color: #009FDF !important;
    color: white !important;
  }
  .v-step[x-placement^="top"] .v-step__arrow.v-step__arrow--dark{
    border-top-color: white !important;
  }
  .v-step[x-placement^="bottom"] .v-step__arrow.v-step__arrow--dark{
    border-bottom-color: white !important;
  }
  .v-step[x-placement^="right"] .v-step__arrow.v-step__arrow--dark{
    border-right-color: white !important;
  }
  .v-step[x-placement^="left"] .v-step__arrow.v-step__arrow--dark{
    border-left-color: white !important;
  }
</style>
