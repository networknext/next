<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-body">
            <div class="card-title">
              <div class="row" v-if="stepOne">
                <div class="col"></div>
                <img class="logo-sizing" src="https://storage.googleapis.com/network-next-press-kit/networknext_logo_colour_black_RGB.png" />
                <div class="col"></div>
              </div>
              <div class="row">
                <h3 class="header">{{ stepOne ? 'Forgot Your Password?' : 'Check Your Email'}}</h3>
              </div>
            </div>
            <form @submit.prevent="resetPassword()">
              <div class="form-group">
                <p style="text-align: center;">
                  {{ stepOne ? 'Enter your email address and we will send you instructions to reset your password.' : `Please check the email address ${email} for instructions to reset your password.` }}
                </p>
                <input
                  type="text"
                  class="form-control"
                  id="email-input"
                  placeholder="Email address"
                  autocomplete="off"
                  v-model="email"
                  v-if="stepOne"
                />
                <small class="text-danger" v-if="emailError !== ''">
                  {{ emailError }}
                  <br/>
                </small>
              </div>
              <button type="submit" class="btn btn-primary btn-block" v-if="stepOne">
                Continue
              </button>
              <button type="submit" class="btn btn-outline-secondary btn-block" v-if="!stepOne">
                Resend email
              </button>
            </form>
            <div style="padding: 1rem 0 1rem 0; text-align: center;"><router-link to="/map"><strong>Back to Portal</strong></router-link></div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component opens up a login form modal
 */

@Component
export default class ResetPasswordModal extends Vue {
  private email: string
  private emailError: string
  private stepOne: boolean
  private validEmail: boolean

  constructor () {
    super()
    this.email = ''
    this.emailError = ''
    this.stepOne = true
    this.validEmail = false
  }

  // Leaving this here to make sending forgotten password links easier in the future
  private mounted () {
    const email = this.$route.query.email || '' // The hell is a <string | (string | null)[]>!?
    if (typeof email === 'string') { // TODO: see if there is a way around this. Typescript doesn't like the (string | null)[] secondary type definition...
      this.email = email
    }
    // TODO: Find a better way of doing this
    this.checkEmail(false)
  }

  private checkEmail (checkLength: boolean) {
    const regex = new RegExp(/(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])/)
    this.validEmail = !checkLength || (this.email.length > 0 && regex.test(this.email))
    this.emailError = this.validEmail ? '' : 'Please enter a valid email address'
  }

  private resetPassword (): void {
    // TODO: Find a better way of doing this
    this.checkEmail(true)
    if (!this.validEmail) {
      return
    }
    this.$apiService.sendResetPasswordEmail({ email: this.email })
      .then(() => {
        this.stepOne = false
      })
      .catch((err: Error) => {
        this.emailError = 'Could not send password reset email. Please verify that the email is linked to a valid account and try again'
        console.log(err)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .logo-sizing {
    width: 320px;
    height: 37px;
  }
  .header {
    padding-top: 1rem;
    text-align: center;
    width: 100%;
  }
  .modal-mask {
    position: fixed;
    z-index: 9998;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgb(0, 0, 0);
    display: table;
  }

  .modal-wrapper {
    display: table-cell;
    vertical-align: middle;
  }

  .modal-container {
    max-width: 400px;
    max-height: 800px;
    margin: 0px auto 10%;
    background-color: #fff;
    border-radius: 5px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
    transition: all 0.3s ease;
  }

  .modal-header h3 {
    margin-top: 0;
    color: #42b983;
  }

  .modal-body {
    margin: 20px 0;
  }

  .modal-default-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }

  /*
  * The following styles are auto-applied to elements with
  * transition="modal" when their visibility is toggled
  * by Vue.js.
  *
  * You can easily play with the modal transition by editing
  * these styles.
  */

  .modal-enter {
    opacity: 0;
  }

  .modal-leave-active {
    opacity: 0;
  }

  .modal-enter .modal-container,
  .modal-leave-active .modal-container {
    -webkit-transform: scale(1.1);
    transform: scale(1.1);
  }

  .my-custom-scrollbar {
    position: relative;
    height: 300px;
    overflow: auto;
  }
  .table-wrapper-scroll-y {
    display: block;
  }
</style>
