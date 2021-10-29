<template>
  <div :class="[{'alert': true}, className]" role="alert" v-if="alertMessage !== ''">
    {{ alertMessage }}
    <slot v-if="showSlots"></slot>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component is a reusable alert component
 * It takes in a message and alert type props
 *  and will display an alert with the passed in
 *  message with a class attribute equivalent to the
 *  passed in alert type
 */

import { AlertType } from '@/components/types/AlertTypes'
import { EMAIL_CONFIRMATION_MESSAGE } from './types/Constants'

@Component
export default class Alert extends Vue {
  get alertMessage (): string {
    return this.message
  }

  get className (): string {
    return this.alertType
  }

  private message: string
  private alertType: AlertType
  private showSlots: boolean

  constructor () {
    super()
    this.message = ''
    this.alertType = AlertType.DEFAULT
    this.showSlots = true
  }

  public setMessage (message: string) {
    this.message = message
  }

  public setAlertType (alertType: AlertType) {
    this.alertType = alertType
  }

  public resetAlert () {
    this.alertType = AlertType.DEFAULT
    this.message = ''
  }

  public toggleSlots (toggle: boolean) {
    this.showSlots = toggle
  }

  public resendVerificationEmail () {
    const userId = this.$store.getters.userProfile.auth0ID
    const email = this.$store.getters.userProfile.email

    this.$apiService
      .resendVerificationEmail({
        user_id: userId,
        user_email: email,
        redirect: window.location.origin,
        connection: 'Username-Password-Authentication'
      })
      .then(() => {
        this.showSlots = false
        this.setMessage('Verification email was sent successfully. Please check your email for futher instructions.')
        this.setAlertType(AlertType.SUCCESS)
        setTimeout(() => {
          this.setMessage(`${EMAIL_CONFIRMATION_MESSAGE} ${this.$store.getters.userProfile.email}`)
          this.setAlertType(AlertType.INFO)
          this.showSlots = true
        }, 5000)
      })
      .catch((error: Error) => {
        this.showSlots = false
        console.log('something went wrong with resending verification email')
        console.log(error)
        this.setMessage('Something went wrong sending the verification email. Please try again later.')
        this.setAlertType(AlertType.ERROR)
        setTimeout(() => {
          this.setMessage(`${EMAIL_CONFIRMATION_MESSAGE} ${this.$store.getters.userProfile.email}`)
          this.setAlertType(AlertType.INFO)
          this.showSlots = true
        }, 5000)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  div {
    text-align: center;
  }
  .alert {
    margin-bottom: 0rem;
  }
</style>
