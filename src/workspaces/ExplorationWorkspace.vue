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
      <h1 class="h2">
        Exploration
      </h1>
      <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4">
        <Alert ref="verifyAlert"></Alert>
      </div>
    </div>
    <div class="card" style="margin-bottom: 250px;">
      <div class="card-header">
        <ul class="nav nav-tabs card-header-tabs">
          <li class="nav-item">
            <router-link to="/explore/notifications" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'notifications'}">Notifications</router-link>
          </li>
          <li class="nav-item" v-if="false">
            <router-link to="/explore/analytics" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'analytics'}">Analytics</router-link>
          </li>
          <li class="nav-item" v-if="false && $store.getters.isSeller">
            <router-link to="/explore/supply" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'supply'}">Supply</router-link>
          </li>
          <li class="nav-item" v-if="false">
            <router-link to="/explore/invoicing" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'invoicing'}">Invoicing</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import { RELOAD_MESSAGE } from '@/components/types/Constants'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert
  }
})
export default class ExplorationWorkspace extends Vue {
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
    this.$refs.verifyAlert.setMessage(RELOAD_MESSAGE)
    this.$refs.verifyAlert.setAlertType(AlertType.ERROR)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
