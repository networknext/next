<template>
  <div class="container-fluid below-nav-bar">
    <Alert :message="alertMessage" :alertType="alertType" v-if="message !== ''">
      <a href="#" @click="resendVerificationEmail()">
        Resend email
      </a>
    </Alert>
    <div class="row">
      <main role="main" class="col-md-12 col-lg-12 px-4">
        <SessionCounts
          v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'"
        />
        <router-view />
      </main>
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
import Alert from '@/components/Alert.vue'
import { AlertTypes } from '@/components/types/AlertTypes'

/**
 * This component is the base component for all other workspace components
 * It also holds the Email Verification alert and Session Count components
 *  so that they are visible across all workspaces if necessary
 */

@Component({
  components: {
    Alert,
    MapWorkspace,
    SessionCounts,
    SessionsWorkspace,
    SessionToolWorkspace,
    SettingsWorkspace
  }
})
export default class Workspace extends Vue {
  get alertMessage () {
    return this.message
  }

  private message: string
  private alertType: string
  private vueInstance: any
  private unwatch: any

  constructor () {
    super()
    this.alertType = AlertTypes.INFO
    this.vueInstance = Vue
    this.message = ''
    this.unwatch = this.$store.watch(
      (_, getters: any) => getters.isAnonymousPlus,
      (showAlert: boolean) => {
        // Not sure why this is necessary but Watch seems to need a function call
        this.updateAlert(showAlert)
      }
    )
  }

  private destroy () {
    this.unwatch()
  }

  // Not sure why this is necessary but Vue is ignoring all updates to message
  private updateAlert (showAlert: boolean) {
    this.message = showAlert ? `Please confirm your email address: ${this.$store.getters.userProfile.email}` : ''
  }

  private resendVerificationEmail () {
    const userId = this.$store.getters.userProfile.auth0ID
    const email = this.$store.getters.userProfile.email

    this.$apiService
      .resendVerificationEmail({
        user_id: userId,
        user_email: email,
        redirect: window.location.origin,
        connection: 'Username-Password-Authentication'
      })
      .then((response: any) => {
        this.message =
          'Verification email was sent successfully. Please check your email for futher instructions.'
        this.alertType = AlertTypes.SUCCESS
      })
      .catch((error: Error) => {
        console.log('something went wrong with resending verification email')
        console.log(error)
        this.message =
          'Something went wrong sending the verification email. Please try again later.'
        this.alertType = AlertTypes.ERROR
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style scoped lang="scss">
</style>
