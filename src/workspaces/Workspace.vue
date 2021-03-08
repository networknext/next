<template>
  <div class="container-fluid below-nav-bar">
    <div class="row">
      <main role="main" class="col-md-12 col-lg-12 px-4">
        <SessionCounts
          v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'"
        />
        <router-view />
      </main>
      <v-tour v-show="$store.getters.currentPage === 'map'" name="mapTour" :steps="mapTourSteps" :options="mapTourOptions" :callbacks="mapTourCallbacks"></v-tour>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import SessionCounts from '@/components/SessionCounts.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'

/**
 * This component is the base component for all other workspace components
 * It also holds the Email Verification alert and Session Count components
 *  so that they are visible across all workspaces if necessary
 */

@Component({
  components: {
    MapWorkspace,
    SessionCounts,
    SessionsWorkspace,
    SessionToolWorkspace,
    SettingsWorkspace
  }
})
export default class Workspace extends Vue {
  private mapTourSteps: Array<any>
  private mapTourOptions: any
  private mapTourCallbacks: any

  constructor () {
    super()
    this.mapTourSteps = [
      {
        target: '#map',
        header: {
          title: 'Welcome to Network Next!'
        },
        content: 'Each green dot on the <strong>Map </strong>is a player being accelerated by Network Next',
        params: {
          enabledButtons: {
            buttonSkip: true,
            buttonPrevious: false,
            buttonNext: true,
            buttonStop: false
          }
        }
      },
      {
        target: '[data-tour="sessionsLink"]',
        header: {
          title: 'Sessions Link'
        },
        content: 'Click <strong>Sessions</strong> to learn more about what Network Next does!'
      }
    ]

    this.mapTourOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'DONE'
      }
    }

    this.mapTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_TOURS', 'map')
      },
      onSkip: () => {
        this.$store.commit('UPDATE_FINISHED_TOURS', 'map')
      }
    }
  }

  private mounted () {
    if (this.$store.getters.isTour && this.$route.name === 'map' && this.$tours.mapTour && !this.$tours.mapTour.isRunning) {
      this.$tours.mapTour.start()
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style lang="scss">
  .v-step {
    background-color: white !important;
    color: black;
  }
  .v-step__header {
    padding: 1rem !important;
    background-color: white !important;
    color: black;
    font-weight: bold;
    text-align: left;
    margin-bottom: 0 !important;
  }
  .v-step__content {
    background-color: white !important;
    color: black;
    text-align: left;
  }
  .v-step__button {
    border-color: black !important;
    color: black !important;
    font-weight: lighter;
    min-width: 80px;
  }
  .v-step__button.v-step__button-next {
    background-color: rgb(0, 159, 223) !important;
    border-color: rgb(0, 159, 223) !important;
    color: white !important;
  }
  .v-step__button.v-step__button-stop {
    background-color: rgb(0, 159, 223) !important;
    border-color: rgb(0, 159, 223) !important;
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
