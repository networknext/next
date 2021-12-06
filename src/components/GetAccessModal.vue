<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-body">
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
            <form id="get-access-form" @submit.prevent="stepOne ? switchSteps() : processNewSignup()">
              <div v-if="stepOne" class="form-group">
                <p style="text-align: center;">
                  Please enter your email and create a secure password to get access to the SDK, documentation and to set up a company account.
                </p>
                <input
                  type="email"
                  class="form-control"
                  id="email-input"
                  placeholder="Email"
                  autocomplete="off"
                  v-model="email"
                />
                <small id="email-error" v-if="emailError !== ''" class="text-danger">
                  {{ emailError }}
                  <br/>
                </small>
                <br />
                <input
                  type="password"
                  class="form-control"
                  id="password-input"
                  placeholder="Password"
                  autocomplete="off"
                  v-model="password"
                />
                <small id="password-error" class="text-danger" v-if="!validPassword">
                  Please enter a valid password
                  <br/>
                </small>
                <br v-if="password.length > 0"/>
                <div v-if="password.length > 0" class="password-checker">
                  <div>
                    <span>Your password must contain:</span>
                  </div>
                  <ul style="padding-left: inherit;">
                    <li>
                      At least 8 characters <font-awesome-icon id="length-check" icon="check" class="fa-w-16 fa-fw" :style="{'padding-left': '.5rem', 'color': password.length >= 8 ? 'green' : 'red', 'width': '2rem'}"/>
                    </li>
                    <li>
                      At least 3 of the following:
                    </li>
                    <ul>
                      <li>
                        Lower case letters (a-z) <font-awesome-icon id="lower-check" icon="check" class="fa-w-16 fa-fw" :style="{'padding-left': '.5rem', 'color': hasLowerCase ? 'green' : 'red', 'width': '2rem'}"/>
                      </li>
                      <li>
                        Upper case letters (A-Z) <font-awesome-icon id="upper-check" icon="check" class="fa-w-16 fa-fw" :style="{'padding-left': '.5rem', 'color': hasUpperCase ? 'green' : 'red', 'width': '2rem'}"/>
                      </li>
                      <li>
                        Numbers (0-9) <font-awesome-icon id="number-check" icon="check" class="fa-w-16 fa-fw" :style="{'padding-left': '.5rem', 'color': hasNumbers ? 'green' : 'red', 'width': '2rem'}"/>
                      </li>
                      <li>
                        Special characters (ex. !@#$%^&*) <font-awesome-icon id="special-check" icon="check" class="fa-w-16 fa-fw" :style="{'padding-left': '.5rem', 'color': hasCharacters ? 'green' : 'red', 'width': '2rem'}"/>
                      </li>
                    </ul>
                  </ul>
                </div>
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
                  autocomplete="off"
                  v-model="firstName"
                />
                <small id="first-name-error" v-if="firstNameError !== ''" class="text-danger">
                  {{ firstNameError }}
                  <br/>
                </small>
                <br />
                <input
                  type="text"
                  class="form-control"
                  id="last-name-input"
                  placeholder="Last Name"
                  autocomplete="off"
                  v-model="lastName"
                />
                <small id="last-name-error" v-if="lastNameError !== ''" class="text-danger">
                  {{ lastNameError }}
                  <br/>
                </small>
                <br />
                <input
                  type="text"
                  class="form-control"
                  id="company-name-input"
                  placeholder="Company Name"
                  autocomplete="off"
                  v-model="companyName"
                />
                <small id="company-name-error" v-if="companyNameError !== ''" class="text-danger">
                  {{ companyNameError }}
                  <br/>
                </small>
                <br />
                <input
                  type="text"
                  class="form-control"
                  id="company-website-input"
                  placeholder="Company Website"
                  autocomplete="off"
                  v-model="companyWebsite"
                />
                <small id="company-website-error" v-if="companyWebsiteError !== ''" class="text-danger">
                  {{ companyWebsiteError }}
                  <br/>
                </small>
              </div>
              <button type="submit" class="btn btn-primary btn-block">
                {{ stepOne ? 'Continue' : 'Get Access' }}
              </button>
            </form>
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
  get hasCharacters () {
    const regex = new RegExp(/([#$%&'*+/=?^!_`{|}~-])/)
    return regex.test(this.password)
  }

  get hasLowerCase () {
    const regex = new RegExp(/([a-z])/)
    return regex.test(this.password)
  }

  get hasNumbers () {
    const regex = new RegExp(/([0-9])/)
    return regex.test(this.password)
  }

  get hasUpperCase () {
    const regex = new RegExp(/([A-Z])/)
    return regex.test(this.password)
  }

  private companyName: string
  private companyNameError: string
  private companyWebsite: string
  private companyWebsiteError: string
  private email: string
  private emailError: string
  private firstName: string
  private firstNameError: string
  private lastName: string
  private lastNameError: string
  private password: string
  private stepOne: boolean
  private validCompanyName: boolean
  private validEmail: boolean
  private validFirstName: boolean
  private validLastName: boolean
  private validPassword: boolean
  private validWebsite: boolean

  constructor () {
    super()
    this.companyName = ''
    this.companyNameError = ''
    this.companyWebsite = ''
    this.companyWebsiteError = ''
    this.email = ''
    this.emailError = ''
    this.firstName = ''
    this.firstNameError = ''
    this.lastName = ''
    this.lastNameError = ''
    this.password = ''
    this.stepOne = true
    this.validCompanyName = false
    this.validEmail = false
    this.validFirstName = false
    this.validLastName = false
    this.validPassword = false
    this.validWebsite = false
  }

  // This function is only necessary as a helper for the WIX sign up system
  private mounted () {
    const email = this.$route.query.email || '' // The hell is a <string | (string | null)[]>!?
    if (typeof email === 'string') { // TODO: see if there is a way around this. Typescript doesn't like the (string | null)[] secondary type definition...
      this.email = email
    }
    // TODO: Find a better way of doing this
    this.checkCompanyName(false)
    this.checkEmail(false)
    this.checkFirstName(false)
    this.checkLastName(false)
    this.checkPassword(false)
    this.checkWebsite(false)
  }

  // TODO: Add better checks for all check* functions
  private checkCompanyName (checkLength: boolean) {
    const regex = new RegExp(/^[a-z ,.'-]+$/i)
    this.validCompanyName = !checkLength || (this.companyName.length > 0 && regex.test(this.companyName))
    this.companyNameError = this.validCompanyName ? '' : 'Please enter a valid company name'
  }

  private checkEmail (checkLength: boolean) {
    const regex = new RegExp(/(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])/)
    this.validEmail = !checkLength || (this.email.length > 0 && regex.test(this.email))
    this.emailError = this.validEmail ? '' : 'Please enter a valid email address'
  }

  private checkFirstName (checkLength: boolean) {
    const regex = new RegExp(/^[a-zA-Z]+$/)
    this.validFirstName = !checkLength || (this.firstName.length > 0 && regex.test(this.firstName))
    this.firstNameError = this.validFirstName ? '' : 'Please enter a valid first name'
  }

  private checkLastName (checkLength: boolean) {
    const regex = new RegExp(/^[a-zA-Z]+$/)
    this.validLastName = !checkLength || (this.lastName.length > 0 && regex.test(this.lastName))
    this.lastNameError = this.validLastName ? '' : 'Please enter a valid last name'
  }

  private checkPassword (checkLength: boolean) {
    const regex = new RegExp(/([A-Za-z0-9!#$%&'*+/=?^_`{|}~-]){3,}/)
    this.validPassword = !checkLength || (this.password.length >= 8 && regex.test(this.password))
  }

  private checkWebsite (checkLength: boolean) {
    const regex = new RegExp(/((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)/)
    this.validWebsite = !checkLength || (this.companyWebsite.length > 0 && regex.test(this.companyWebsite))
    this.companyWebsiteError = this.validWebsite ? '' : 'Please enter a valid website. IE: https://networknext.com'
  }

  private processNewSignup (): void {
    // TODO: Find a better way of doing this
    this.checkCompanyName(true)
    this.checkFirstName(true)
    this.checkLastName(true)
    this.checkWebsite(true)
    if (!this.validCompanyName || !this.validFirstName || !this.validLastName || !this.validWebsite) {
      return
    }

    // Send this off to the backend to record the new sign up in hubspot and don't wait for the response
    this.$apiService.processNewSignup({
      company_name: this.companyName,
      company_website: this.companyWebsite,
      email: this.email,
      first_name: this.firstName,
      last_name: this.lastName
    })

    this.$authService.login(this.email, this.password, window.location.origin)
      .catch((err: Error) => {
        console.log('Something went wrong processing the new sign up information')
        console.log(err)
      })
  }

  private switchSteps () {
    // TODO: Find a better way of doing this
    this.checkEmail(true)
    this.checkPassword(true)
    if (!this.validEmail || !this.validPassword) {
      return
    }

    this.$authService.getAccess(
      this.email,
      this.password
    )
      .then(() => {
        this.stepOne = false
      })
      .catch((err: Error) => {
        console.log('Something went wrong during the sign up process')
        console.log(err)
        this.emailError = 'Email has already been used to sign up or is invalid'
        setTimeout(() => {
          this.emailError = ''
        }, 5000)
        this.stepOne = true
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .back-btn {
    margin-bottom: 1rem;
    margin-left: -0.75rem;
    margin-top: -0.5rem;
    cursor: pointer;
    font-size: 24px;
  }
  .logo-sizing {
    width: 320px;
    height: 37px;
  }
  .header {
    padding-top: 1rem;
  }
  .password-checker {
    width: 100%;
    height: 200px;
    border-color: #ced4da;
    border-radius: .25rem;
    border-width: 1px;
    border-style: solid;
    padding: 14px 16px;
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
