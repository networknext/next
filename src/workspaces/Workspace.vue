<template>
  <div class="container-fluid below-nav-bar">
    <div class="row">
      <main role="main" class="col-md-12 col-lg-12 px-4">
        <SessionCounts
          v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'"
        />
        <router-view />
      </main>
      <v-tour name="introTour" :steps="tourSteps" :options="customOptions"></v-tour>
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
  private tourSteps: Array<any>
  private customOptions: any
  private callbacks: any

  constructor () {
    super()
    this.tourSteps = [
      {
        target: '#map',
        header: {
          title: 'Welcome to Network Next!'
        },
        content: 'Each green dot on the <strong>Map</strong> is a player being accelerated by Network Next',
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
      },
      {
        target: '[data-tour="0"]',
        header: {
          title: 'Top Sessions'
        },
        content: 'Click on this <strong>Session ID</strong> to view more stats (such as latency, packet loss and jitter improvements).'
      },
      {
        target: '#latency-chart-1',
        header: {
          title: 'Top Sessions'
        },
        content: 'Stats about a specific session can be viewed in this <strong>Session Tool</strong>. These are real-time improvements to latency, jitter, and packet loss.'
      },
      {
        target: '[data-tour="signUpButton"]',
        header: {
          title: 'Get Access'
        },
        content: '<strong>Try it for your game for FREE!</strong> Just create an account and log in to try Network Next: <ul><li>Download the open source SDK and documentation.</li><li>Integrate the SDK into your game.</li></ul> Now you\'re in control of the network. Please contact us in <strong>chat</strong> (lower right) if you have any questions.',
        params: {
          enableScrolling: true
        }
      }
    ]

    this.customOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'DONE'
      }
    }

    this.callbacks = {
      onFinish: () => {
        this.$store.commit('TOGGLE_IS_TOUR', false)
        this.$store.commit('UPDATE_CURRENT_TOUR_STEP', -1)
      },
      onSkip: () => {
        this.$store.commit('TOGGLE_IS_TOUR', false)
        this.$store.commit('UPDATE_CURRENT_TOUR_STEP', -1)
      }
    }
  }

  private mounted () {
    console.log(this.$tours)
    if (this.$store.getters.isTour) {
      this.$tours.introTour.start()
    }
  }

  private beforeDestroy () {
    if (this.$store.getters.isTour) {
      this.$tours.introTour.stop()
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style scoped lang="scss">
</style>
