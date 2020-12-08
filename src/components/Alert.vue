<template>
  <div :class="[{'alert': true}, className]" role="alert">
    {{ alertMessage }}
    <slot v-if="showSlots"></slot>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'

/**
 * This component is a reusable alert component
 * It takes in a message and alert type props
 *  and will display an alert with the passed in
 *  message with a class attribute equivalent to the
 *  passed in alert type
 */

/**
 * TODO: Add helper function that make it easier to set the message and alert type
 *  It is kind of a pain to deal with when there are multiple alerts on the page
 *  Similar idea to the sessions count component
 */

import { AlertType } from '@/components/types/AlertTypes'

@Component
export default class Alert extends Vue {
  @Prop({ required: false, type: String, default: '' }) message!: string
  @Prop({ required: false, type: String, default: AlertType.DEFAULT }) alertType!: string

  get alertMessage (): string {
    return this.currentMessage
  }

  get className (): string {
    return this.currentClass
  }

  private givenMessage: string
  private givenClass: string
  private currentMessage: string
  private currentClass: string
  private showSlots: boolean

  constructor () {
    super()
    this.givenMessage = this.message
    this.givenClass = this.alertType
    this.currentMessage = this.givenMessage
    this.currentClass = this.givenClass
    this.showSlots = true
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
      .then((response: any) => {
        this.showSlots = false
        this.currentMessage =
          'Verification email was sent successfully. Please check your email for futher instructions.'
        this.currentClass = AlertType.SUCCESS
        setTimeout(() => {
          this.currentMessage = this.givenMessage
          this.currentClass = this.givenClass
          this.showSlots = true
        }, 5000)
      })
      .catch((error: Error) => {
        this.showSlots = false
        console.log('something went wrong with resending verification email')
        console.log(error)
        this.currentMessage =
          'Something went wrong sending the verification email. Please try again later.'
        this.currentClass = AlertType.ERROR
        setTimeout(() => {
          this.currentMessage = this.givenMessage
          this.currentClass = this.givenClass
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
