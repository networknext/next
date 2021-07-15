<template>
  <div class="card-body" id="config-page">
    <h5 class="card-title">Game Configuration</h5>
    <p class="card-text">Manage how your game connects to Network Next.</p>
    <Alert ref="responseAlert" />
    <form v-on:submit.prevent="updatePubKey()">
      <div class="form-group" id="pubKey">
        <label>Company Name</label>
        <input
          type="text"
          class="form-control"
          id="company-input"
          :disabled="true"
          v-bind:value="companyName"
        />
        <br />
        <label>Public Key</label>
        <textarea
          class="form-control"
          placeholder="Enter your base64-encoded public key"
          id="pubkey-input"
          :disabled="!$store.getters.isOwner && !$store.getters.isAdmin"
          v-model="pubKey"
        ></textarea>
      </div>
      <button
        id="game-config-button"
        type="submit"
        class="btn btn-primary btn-sm"
        v-if="$store.getters.isOwner || $store.getters.isAdmin"
      >Save game configuration</button>
      <p class="text-muted text-small mt-2"></p>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import { UserProfile } from '@/components/types/AuthTypes'
import { cloneDeep } from 'lodash'

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
    Alert
  }
})
export default class GameConfiguration extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    responseAlert: Alert;
  }

  private companyName: string
  private pubKey: string
  private userProfile: UserProfile

  constructor () {
    super()
    this.companyName = ''
    this.pubKey = ''
    this.userProfile = {} as UserProfile
  }

  private mounted () {
    this.userProfile = cloneDeep(this.$store.getters.userProfile)
    if (this.companyName === '') {
      this.companyName = this.userProfile.companyName || ''
    }
    if (this.pubKey === '') {
      this.pubKey = this.userProfile.pubKey || ''
    }
  }

  private updatePubKey () {
    this.$apiService
      .updateGameConfiguration({
        new_public_key: this.pubKey
      })
      .then((response: any) => {
        this.userProfile.pubKey = response.public_key
        this.$store.commit('UPDATE_USER_PROFILE', this.userProfile)
        this.$refs.responseAlert.setMessage('Updated public key successfully')
        this.$refs.responseAlert.setAlertType(AlertType.SUCCESS)
        setTimeout(() => {
          if (this.$refs.responseAlert) {
            this.$refs.responseAlert.resetAlert()
          }
        }, 5000)
        this.$apiService.fetchAllBuyers()
          .then((response: any) => {
            const allBuyers = response.buyers
            this.$store.commit('UPDATE_ALL_BUYERS', allBuyers)
          })
        this.$apiService.sendPublicKeyEnteredSlackNotification({ email: this.$store.getters.userProfile.email, company_name: this.$store.getters.userProfile.companyName, company_code: this.$store.getters.userProfile.companyCode })
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the public key')
        console.log(error)
        this.$refs.responseAlert.setMessage('Failed to update public key')
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
