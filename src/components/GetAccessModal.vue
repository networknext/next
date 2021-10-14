<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-body">
            <font-awesome-icon
              icon="arrow-left"
              class="fa-w-16 fa-fw"
              v-if="!stepOne"
            />
            <div class="card-title">
              <div class="row">
                <div class="col"></div>
                <img class="logo-sizing" src="https://storage.googleapis.com/network-next-press-kit/networknext_logo_colour_black_RGB.png" />
                <div class="col"></div>
              </div>
              <div class="row">
                <div class="col"></div>
                <h2 class="header">Get Access</h2>
                <div class="col"></div>
              </div>
            </div>
            <div v-if="stepOne" class="form-group">
              <p style="text-align: center;">
                Please enter your email and create a secure password to get access to the SDK, documentation and to set up a company account.
              </p>
              <input
                type="text"
                class="form-control"
                id="email-input"
                placeholder="Email"
              />
              <small v-for="(error, index) in emailErrors" :key="index" class="text-danger">
                {{ error }}
                <br/>
              </small>
              <br />
              <input
                type="password"
                class="form-control"
                id="password-input"
                placeholder="Password"
              />
              <small v-for="(error, index) in passwordErrors" :key="index" class="text-danger">
                {{ error }}
                <br/>
              </small>
              <br />
              <input
                type="password"
                class="form-control"
                id="passwprd-input"
                placeholder="Confirm Password"
              />
              <small v-for="(error, index) in confirmPasswordErrors" :key="index" class="text-danger">
                {{ error }}
                <br/>
              </small>
            </div>
            <div v-if="!stepOne" class="form-group">
              <p style="text-align: center;">
                Please enter a company name and website so that our team can learn more about your company to help make your on boarding experience smoother.
              </p>
              <input
                type="text"
                class="form-control"
                id="first-name-input"
                placeholder="First Name"
              />
              <small v-for="(error, index) in firstNameErrors" :key="index" class="text-danger">
                {{ error }}
                <br/>
              </small>
              <br />
              <input
                type="text"
                class="form-control"
                id="last-name-input"
                placeholder="Last Name"
              />
              <small v-for="(error, index) in lastNameErrors" :key="index" class="text-danger">
                {{ error }}
                <br/>
              </small>
              <br />
              <input
                type="text"
                class="form-control"
                id="company-name-input"
                placeholder="Company Name"
              />
              <br />
              <input
                type="text"
                class="form-control"
                id="company-website-input"
                placeholder="Company Website"
              />
            </div>
            <button v-if="stepOne" class="btn btn-primary btn-block" @click="switchSteps(false)" :disabled="!validForm">
              Continue
            </button>
            <button v-if="!stepOne" class="btn btn-primary btn-block" @click="getAccess()">
              Get Access
            </button>
            <div style="padding: 1rem 0 1rem 0;">Already have an account? <router-link to="login"><strong>Log in</strong></router-link></div>
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
export default class GetAccessModal extends Vue {
  get validForm () {
    this.checkPasswords()
    this.checkNames()
    this.checkEmail()

    return (
      this.confirmPasswordErrors.length +
      this.emailErrors.length +
      this.firstNameErrors.length +
      this.lastNameErrors.length +
      this.passwordErrors.length
    ) === 0
  }

  private confirmPassword: string
  private confirmPasswordErrors: Array<string>
  private email: string
  private emailErrors: Array<string>
  private firstName: string
  private firstNameErrors: Array<string>
  private lastName: string
  private lastNameErrors: Array<string>
  private password: string
  private passwordErrors: Array<string>
  private stepOne: boolean

  constructor () {
    super()
    this.confirmPassword = ''
    this.confirmPasswordErrors = []
    this.email = ''
    this.emailErrors = []
    this.firstName = ''
    this.firstNameErrors = []
    this.lastName = ''
    this.lastNameErrors = []
    this.password = ''
    this.passwordErrors = []
    this.stepOne = true
  }

  // This function is only necessary as a helper for the WIX sign up system
  private mounted () {
    const email = this.$route.query.email || '' // The hell is a <string | (string | null)[]>!?
    if (typeof email === 'string') { // TODO: see if there is a way around this. Typescript doesn't like the (string | null)[] secondary type definition...
      this.email = email
    }
  }

  private checkPasswords () {
    console.log('Checking passwords')
  }

  private checkNames () {
    console.log('Checking names')
  }

  private checkEmail () {
    if (this.email === '') {
      this.emailErrors.push('Please enter a valid email address')
    }
  }

  private getAccess (): void {
    this.$authService.getAccess(this.email, this.password)
  }

  private switchSteps (isFirstStep: boolean) {
    this.stepOne = isFirstStep
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
    max-height: 600px;
    margin: 0px auto;
    background-color: #fff;
    border-radius: 5px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
    transition: all 0.3s ease;
    font-family: Helvetica, Arial, sans-serif;
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
