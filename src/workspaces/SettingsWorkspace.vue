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
    ">
      <h1 class="h2">
        Settings
      </h1>
      <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4">
        <Alert ref="verifyAlert">
          <a href="#" @click="$refs.verifyAlert.resendVerificationEmail()">
            Resend email
          </a>
        </Alert>
      </div>
    </div>
    <div class="card" style="margin-bottom: 250px;">
      <div class="card-header">
        <ul class="nav nav-tabs card-header-tabs">
          <li class="nav-item" v-if="!$store.getters.isAnonymous">
            <router-link to="/settings/account" class="nav-link" :class="{ active: $store.getters.currentPage === 'account-settings'}">Account Settings</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.registeredToCompany && ($store.getters.isAdmin || $store.getters.isOwner)">
            <router-link to="/settings/game-config" class="nav-link" :class="{ active: $store.getters.currentPage === 'config'}">Game Settings</router-link>
          </li>
          <li class="nav-item" v-if="$flagService.isEnabled(FeatureEnum.FEATURE_ROUTE_SHADER) && ($store.getters.isAdmin || $store.getters.isOwner)">
            <router-link to="/settings/route-shader" class="nav-link" :class="{ active: $store.getters.currentPage === 'shader'}">Route Shader</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.registeredToCompany && ($store.getters.isAdmin || $store.getters.isOwner)">
            <router-link to="/settings/users" class="nav-link" :class="{ active: $store.getters.currentPage === 'users'}">Users</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import { EMAIL_CONFIRMATION_MESSAGE, RELOAD_MESSAGE } from '@/components/types/Constants'
import { Component, Vue } from 'vue-property-decorator'
import { FeatureEnum } from '../components/types/FeatureTypes'

/**
 * This component holds the workspace elements related to the settings page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert
  }
})
export default class SettingsWorkspace extends Vue {
  private FeatureEnum: any
  // Register the alert component to access its set methods
  $refs!: {
    verifyAlert: Alert;
  }

  private created () {
    this.FeatureEnum = FeatureEnum
  }

  private mounted () {
    if (this.$store.getters.isAnonymousPlus) {
      this.$refs.verifyAlert.setMessage(`${EMAIL_CONFIRMATION_MESSAGE} ${this.$store.getters.userProfile.email}`)
      this.$refs.verifyAlert.setAlertType(AlertType.INFO)
    }

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
