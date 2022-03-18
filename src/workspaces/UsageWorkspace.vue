<template>
  <div>
    <div class="
              d-flex
              justify-content-between
              flex-wrap
              flex-md-nowrap
              align-items-center
              mb-3
            "
    >
      <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4">
        <Alert ref="verifyAlert"></Alert>
      </div>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" style="max-width: 400px;" v-if="$store.getters.isAdmin">
        <div class="mr-auto"></div>
        <BuyerFilter :includeAll="false" :liveOnly="false"/>
      </div>
    </div>
    <div class="card" style="margin-bottom: 250px;">
      <Usage />
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import BuyerFilter from '@/components/BuyerFilter.vue'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import Usage from '@/components/Usage.vue'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert,
    BuyerFilter,
    Usage
  }
})
export default class UsageWorkspace extends Vue {
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
