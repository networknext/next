<template>
  <div>
    <div class="
              d-flex
              justify-content-between
              flex-wrap
              flex-md-nowrap
              align-items-center
              pt-3
              pb-2
              mb-3
              border-bottom
            "
    >
      <div>
        <h1 class="h2">
          Analytics (Beta)
        </h1>
        <div>Curated dashboards that provide a look into your data. This feature is in beta and will be receiving continuous updates.</div>
      </div>
      <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4">
        <Alert ref="verifyAlert"></Alert>
      </div>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" style="max-width: 400px;">
        <div class="mr-auto"></div>
        <BuyerFilter v-if="$store.getters.isAdmin" :includeAll="false" :liveOnly="false"/>
      </div>
    </div>
    <Analytics />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import BuyerFilter from '@/components/BuyerFilter.vue'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import Analytics from '../components/Analytics.vue'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert,
    Analytics,
    BuyerFilter
  }
})
export default class AnalyticsWorkspace extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    verifyAlert: Alert;
  }

  private mounted () {
    // If the network is down, show an error
    if (this.$store.getters.killLoops) {
      this.showErrorAlert()
    }
  }

  private showErrorAlert () {
    this.$refs.verifyAlert.toggleSlots(false)
    this.$refs.verifyAlert.setMessage(ErrorTypes.SYSTEM_FAILURE)
    this.$refs.verifyAlert.setAlertType(AlertType.ERROR)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
