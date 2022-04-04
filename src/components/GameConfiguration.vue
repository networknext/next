<template>
  <div
    class="card-body"
    id="config-page"
    v-if="(
      $store.getters.isOwner &&
      $store.getters.userProfile.companyName !== ''
    ) || $store.getters.isAdmin
    ">
    <TermsOfServiceModal v-if="showTOS"/>
    <h5 class="card-title">Game Configuration</h5>
    <p class="card-text">
      Manage how your game connects to Network Next.
    </p>
    <Alert ref="responseAlert" />
    <form v-on:submit.prevent="checkTOS()">
      <div class="form-group" id="pubKey">
        <label>Company Name</label>
        <input
          type="text"
          class="form-control"
          id="company-input"
          :disabled="true"
          :value="companyName"
        />
        <br />
        <label>Public Key</label>
        <textarea
          class="form-control"
          placeholder="Enter your base64-encoded public key"
          id="pubkey-input"
          v-model="pubKey"
        ></textarea>
      </div>
      <button
        id="game-config-button"
        type="submit"
        class="btn btn-primary btn-sm"
        :disabled="pubKey === ''"
      >Save game configuration</button>
      <p class="text-muted text-small mt-2"></p>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { cloneDeep } from 'lodash'
import { UPDATE_PUBLIC_KEY_SUCCESS } from './types/Constants'
import { ErrorTypes } from './types/ErrorTypes'
import TermsOfServiceModal from '@/components/TermsOfServiceModal.vue'

/**
 * This component displays all of the necessary information for the game configuration tab
 *  within the settings page of the Portal and houses all the associated logic and api calls
 */

/**
 * TODO: Clean up template
 * TODO: Pretty sure the card-body can be taken out into a wrapper component - same with route shader and user management...
 */

@Component({
  components: {
    Alert,
    TermsOfServiceModal
  }
})
export default class GameConfiguration extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    responseAlert: Alert;
  }

  private companyName: string
  private pubKey: string
  private showTOS: boolean
  private userProfile: UserProfile

  constructor () {
    super()
    this.companyName = ''
    this.pubKey = ''
    this.showTOS = false
    this.userProfile = newDefaultProfile()
  }

  private mounted () {
    this.userProfile = cloneDeep(this.$store.getters.userProfile)
    if (this.companyName === '') {
      this.companyName = this.userProfile.companyName || ''
    }
    if (this.pubKey === '') {
      this.pubKey = this.userProfile.pubKey || ''
    }

    // TODO: Make a modal events bus rather than using the root application bus
    this.$root.$on('showTOSModal', this.showTOSModalCallback)
    this.$root.$on('hideTOSModal', this.hideTOSModalCallback)
  }

  private beforeDestroy () {
    this.$root.$off('showTOSModal')
    this.$root.$off('hideTOSModal')
  }

  private showTOSModalCallback () {
    if (!this.showTOS) {
      this.showTOS = true
    }
  }

  private hideTOSModalCallback (accepted: boolean) {
    if (this.showTOS) {
      this.showTOS = false
    }

    if (accepted) {
      this.updatePubKey()
    }
  }

  private checkTOS () {
    if (this.$store.getters.userProfile.buyerID === '') {
      // Launch TOS modal
      this.showTOS = true
      return
    }

    this.updatePubKey()
  }

  private updatePubKey () {
    this.$apiService
      .updateGameConfiguration({
        new_public_key: this.pubKey
      })
      .then(() => {
        this.$apiService.sendPublicKeyEnteredSlackNotification({ email: this.$store.getters.userProfile.email, company_name: this.$store.getters.userProfile.companyName, company_code: this.$store.getters.userProfile.companyCode })

        // Give a Looker seat to the Owner of the account
        return this.$apiService.updateUserRoles({
          user_id: this.$store.getters.userProfile.auth0ID,
          roles: [
            {
              name: 'Explorer'
            },
            {
              name: 'Owner'
            }
          ]
        })
      })
      .then(() => {
        this.$refs.responseAlert.setMessage(UPDATE_PUBLIC_KEY_SUCCESS)
        this.$refs.responseAlert.setAlertType(AlertType.SUCCESS)
        setTimeout(() => {
          if (this.$refs.responseAlert) {
            this.$refs.responseAlert.resetAlert()
          }
        }, 5000)

        return this.$authService.refreshToken()
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the public key')
        console.log(error)
        this.$refs.responseAlert.setMessage(ErrorTypes.UPDATE_PUBLIC_KEY_FAILURE)
        this.$refs.responseAlert.setAlertType(AlertType.ERROR)
        setTimeout(() => {
          if (this.$refs.responseAlert) {
            this.$refs.responseAlert.resetAlert()
          }
        }, 5000)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  #game-config-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #game-config-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
</style>
