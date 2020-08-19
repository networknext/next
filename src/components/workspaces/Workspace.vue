<template>
  <div class="container-fluid below-nav-bar">
    <div style="padding-top: 20px;" v-if="message !== ''">
      <Alert :message="message" :alertType="alertType">
        <a href="#" @click="resendVerificationEmail()">
          Resend email
        </a>
      </Alert>
    </div>
    <div class="row">
      <main role="main" class="col-md-12 col-lg-12 px-4">
        <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom" v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'">
          <SessionCounts/>
        </div>
        <router-view/>
      </main>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import MapWorkspace from './MapWorkspace.vue'
import SessionCounts from '../SessionCounts.vue'
import SessionsWorkspace from './SessionsWorkspace.vue'
import SessionToolWorkspace from './SessionToolWorkspace.vue'
import SettingsWorkspace from './SettingsWorkspace.vue'
import Alert from '@/components/Alert.vue'
import { AlertTypes } from '../types/AlertTypes'
import APIService from '@/services/api.service'

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
  private apiService: APIService
  private message: string
  private alertType: string

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
    this.message = ''
    this.alertType = ''
  }

  mounted () {
    if (this.$store.getters.isAnonymousPlus) {
      this.message = `Please confirm your email address: ${this.$store.getters.userProfile.email}`
      this.alertType = AlertTypes.INFO
    }
  }

  private resendVerificationEmail () {
    const userId = this.$store.getters.userProfile.auth0ID
    const email = this.$store.getters.userProfile.email
    this.apiService
      .resendVerificationEmail({
        user_id: userId,
        user_email: email,
        redirect: window.location.origin,
        connection: 'Username-Password-Authentication'
      })
      .then((response: any) => {
        this.message = 'Verification email was sent successfully. Please check your email for futher instructions.'
        this.alertType = AlertTypes.SUCCESS
      })
      .catch((error: Error) => {
        console.log('something went wrong with resending verification email')
        console.log(error)
        this.message = 'Something went wrong sending the verification email. Please try again later.'
        this.alertType = AlertTypes.ERROR
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style scoped lang="scss">
</style>
