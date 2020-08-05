<template>
  <div class="card-body" id="config-page">
    <h5 class="card-title">
      Game Configuration
    </h5>
    <p class="card-text">
      Manage how your game connects to Network Next.
    </p>
    <Alert :message="message" :alertType="alertType" v-if="message !== ''"/>
    <form v-on:submit.prevent="updatePubKey()">
      <div class="form-group" id="pubKey">
        <label>
          Company Name
        </label>
        <input type="text"
                class="form-control"
                placeholder="Enter your company name"
                id="company-input"
                :disabled="!$store.getters.isOwner && !$store.getters.isAdmin"
                v-model="company"
        >
        <br>
        <label>
          Public Key
        </label>
        <textarea class="form-control"
                  placeholder="Enter your base64-encoded public key"
                  id="pubkey-input"
                  :disabled="!$store.getters.isOwner && !$store.getters.isAdmin"
                  v-model="pubKey"
        >
        </textarea>
      </div>
      <button type="submit"
              class="btn btn-primary btn-sm"
              v-if="$store.getters.isOwner && $store.getters.isAdmin"
      >
        Save game configuration
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import APIService from '@/services/api.service'
import Alert from '@/components/Alert.vue'
import { AlertTypes } from './types/AlertTypes'
import { UserProfile } from '@/services/auth.service'

@Component({
  components: {
    Alert
  }
})
export default class GameConfiguration extends Vue {
  private apiService: APIService
  private company: string
  private pubKey: string
  private message: string
  private alertType: string
  private userProfile: UserProfile

  constructor () {
    super()
    this.userProfile = JSON.parse(JSON.stringify(this.$store.getters.userProfile))
    console.log(this.userProfile)
    this.apiService = Vue.prototype.$apiService
    this.company = this.userProfile.company || ''
    this.pubKey = this.userProfile.pubKey || ''
    this.message = ''
    this.alertType = ''
  }

  private mounted () {
    this.pubKey = this.$store.getters.userProfile.pubKey
  }

  private updatePubKey () {
    const domain = this.$store.getters.userProfile.domain || ''

    this.apiService
      .updateGameConfiguration({ name: this.company, domain: domain, new_public_key: this.pubKey })
      .then((response) => {
        this.userProfile.pubKey = response.game_config.public_key
        this.userProfile.company = response.game_config.company
        this.userProfile.buyerID = response.game_config.buyer_id
        this.$store.commit('UPDATE_USER_PROFILE', this.userProfile)
        this.alertType = AlertTypes.INFO
        this.message = 'Updated public key successfully'
        setTimeout(() => {
          this.message = ''
        }, 5000)
      })
      .catch((e) => {
        console.log('Something went wrong updating the public key')
        console.log(e)
        this.alertType = AlertTypes.ERROR
        this.message = 'Failed to update public key'
        setTimeout(() => {
          this.message = ''
        }, 5000)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
