<template>
  <div class="card-body">
    <h5 class="card-title">
      User Details
    </h5>
    <p class="card-text">
      Update user account profile.
    </p>
    <Alert ref="accountResponseAlert"/>
    <form @submit.prevent="updateAccountDetails()">
      <div class="form-group">
        <label for="firstName">
          First Name
        </label>
        <input type="text" class="form-control form-control-sm" id="firstName" v-model="firstName" placeholder="Enter your first name" @change="checkFirstName()"/>
        <small v-for="(error, index) in firstNameErrors" :key="index" class="text-danger">
          {{ error }}
          <br/>
        </small>
      </div>
      <div class="form-group">
        <label for="lastName">
          Last Name
        </label>
        <input type="text" class="form-control form-control-sm" id="lastName" v-model="lastName" placeholder="Enter your last name" @change="checkLastName()"/>
        <small v-for="(error, index) in lastNameErrors" :key="index" class="text-danger">
          {{ error }}
          <br/>
        </small>
      </div>
      <div class="form-group">
        <div class="form-check">
          <input type="checkbox" class="form-check-input" id="newsletterConsent" v-model="newsletterConsent"/>
          <small>
            I would like to receive the Network Next newsletter
          </small>
        </div>
      </div>
      <button id="account-settings-button" type="submit" class="btn btn-primary btn-sm">
        Update User Details
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
    <hr class="mt-4 mb-4">
    <h5 class="card-title">
      Company Details
    </h5>
    <p class="card-text">
      Create or assign yourself to a company account.
    </p>
    <Alert ref="companyResponseAlert"/>
    <form @submit.prevent="setupCompanyAccount()">
      <div class="form-group">
        <label for="companyName">
          Company Name
        </label>
        <input :disabled="$store.getters.userProfile.companyName !== ''" type="text" class="form-control form-control-sm" id="companyName" v-model="companyName" placeholder="Enter your company name" @change="checkCompanyName()"/>
        <small class="form-text text-muted">
          This is the name of the company that you would like your account to be assigned to. This is not necessary for existing company assignment and is case and white space sensitive.
        </small>
        <small v-for="(error, index) in companyNameErrors" :key="index" class="text-danger">
          {{ error }}
          <br/>
        </small>
      </div>
      <div class="form-group">
        <label for="companyCode">
          Company Code
        </label>
        <input :disabled="$store.getters.userProfile.companyCode !== ''" type="text" class="form-control form-control-sm" id="companyCode" v-model="companyCode" placeholder="Enter your company code" @change="checkCompanyCode()"/>
        <small class="form-text text-muted">
          This is the unique string associated to your company account and to be used in your company's subdomain. To assign this user account to an existing company, type in your companies existing code. Examples: mycompany, my-company, my-company-name
        </small>
        <small v-for="(error, index) in companyCodeErrors" :key="index" class="text-danger">
          {{ error }}
          <br/>
        </small>
      </div>
      <button v-if="$store.getters.userProfile.companyCode === '' && $store.getters.userProfile.companyName === ''" id="account-settings-button" type="submit" class="btn btn-primary btn-sm">
        Setup Company Account
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
    <form v-if="false">
      <div class="form-group">
        <label for="newPassword">
          Update Password
        </label>
        <input type="password" class="form-control form-control-sm" id="newPassword" v-model="newPassword" @change="checkNewPassword()" placeholder="Enter your new password"/>
        <small v-for="(error, index) in newPasswordErrors" :key="index" class="text-danger">
          {{ error }}
          <br/>
        </small>
      </div>
      <div class="form-group" v-if="false && validPassword">
        <label for="confirmPassword">
          Confirm Password
        </label>
        <input type="password" class="form-control form-control-sm" id="confirmPassword" v-model="confirmPassword" @change="checkConfirmPassword()" placeholder="Confirm password your new password"/>
        <small class="form-text text-muted" v-if="confirmPassword.length === 0">
          Update your password
        </small>
        <small v-for="(error, index) in confirmPasswordErrors" :key="index" class="text-danger">
          {{ error }}
          <br/>
        </small>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" :disabled="!validPasswordForm" style="margin-top: 1rem;" v-if="false">
        Save
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from './Alert.vue'
import { AlertType } from './types/AlertTypes'
import { cloneDeep } from 'lodash'

/**
 * This component displays all of the necessary information for the user management tab
 *  within the settings page of the Portal and houses all the associated logic and api calls
 */

/**
 * TODO: Clean up template
 * TODO: Pretty sure the card-body can be taken out into a wrapper component - same with route shader and game config...
 */

@Component({
  components: {
    Alert
  }
})
export default class AccountSettings extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    companyResponseAlert: Alert;
    accountResponseAlert: Alert;
  }

  get validCompanyInfo (): boolean {
    this.checkCompanyName()
    this.checkCompanyCode()
    return this.companyNameErrors.length === 0 && this.companyCodeErrors.length === 0
  }

  get validUserDetails (): boolean {
    this.checkFirstName()
    this.checkLastName()
    return this.firstNameErrors.length === 0 && this.lastNameErrors.length === 0
  }

  private companyName: string
  private companyNameErrors: Array<string>

  private companyCode: string
  private companyCodeErrors: Array<string>

  private firstName: string
  private firstNameErrors: Array<string>

  private lastName: string
  private lastNameErrors: Array<string>

  private newsletterConsent: boolean
  private AlertType: any

  private newPassword: string
  private validPassword: boolean
  private validPasswordForm: boolean
  private newPasswordErrors: Array<string>
  private confirmPassword: string
  private confirmPasswordErrors: Array<string>

  private unwatchProfile: any

  constructor () {
    super()
    this.companyName = ''
    this.companyNameErrors = []

    this.companyCode = ''
    this.companyCodeErrors = []

    this.firstName = ''
    this.firstNameErrors = []

    this.lastName = ''
    this.lastNameErrors = []

    this.newsletterConsent = false
    this.AlertType = AlertType

    this.newPassword = ''
    this.confirmPassword = ''
    this.validPassword = false
    this.validPasswordForm = false
    this.newPasswordErrors = []
    this.confirmPasswordErrors = []
  }

  private mounted () {
    this.unwatchProfile = this.$store.watch(
      (state: any, getters: any) => {
        return getters.userProfile
      },
      () => {
        const storedFirstName: string = this.$store.getters.userProfile.firstName
        const storedLastName: string = this.$store.getters.userProfile.lastName
        const storedCompanyName: string = this.$store.getters.userProfile.companyName
        const storedCompanyCode: string = this.$store.getters.userProfile.companyCode

        this.firstName = this.firstName !== storedFirstName && storedFirstName !== '' ? storedFirstName : this.firstName
        this.lastName = this.lastName !== storedLastName && storedLastName !== '' ? storedLastName : this.lastName
        this.companyName = this.companyName !== storedCompanyName && storedCompanyName !== '' ? storedCompanyName : this.companyName
        this.companyCode = this.companyCode !== storedCompanyCode && storedCompanyCode !== '' ? storedCompanyCode : this.companyCode
      }
    )

    const userProfile = cloneDeep(this.$store.getters.userProfile)
    this.firstName = userProfile.firstName || ''
    this.lastName = userProfile.lastName || ''
    this.newsletterConsent = userProfile.newsletterConsent || false

    this.companyName = userProfile.companyName || ''
    this.companyCode = userProfile.companyCode || ''
    this.checkCompanyName()
    this.checkCompanyCode()
    // this.checkConfirmPassword()
  }

  private beforeDestroy () {
    this.unwatchProfile()
  }

  private checkFirstName () {
    this.firstNameErrors = []
    if (this.firstName.length === 0) {
      this.firstNameErrors.push('Please enter your first name')
    }

    if (this.firstName.length > 2048) {
      this.firstNameErrors.push('First name is to long, please enter a name that is less that 2048 characters')
    }

    const regex = new RegExp('([A-Za-z][^!?<>()\-_=+|[\]{}@#$%^&*;:"\',.`~\\])\w+')
    if (!regex.test(this.firstName)) {
      this.firstNameErrors.push('A valid first name must include at least one letter')
    }
  }

  private checkLastName () {
    this.lastNameErrors = []
    if (this.lastName.length === 0) {
      this.lastNameErrors.push('Please enter your last name')
    }

    if (this.lastName.length > 2048) {
      this.lastNameErrors.push('Last name is to long, please enter a name that is less that 2048 characters')
    }

    const regex = new RegExp('([A-Za-z][^!?<>()\-_=+|[\]{}@#$%^&*;:"\',.`~\\])\w+')
    if (!regex.test(this.lastName)) {
      this.firstNameErrors.push('A valid last name must include at least one letter')
    }
  }

  private checkCompanyName () {
    this.companyNameErrors = []
    if (this.companyName.length > 256) {
      this.companyNameErrors.push('Please choose a company name that is at most 256 characters')
    }
  }

  private checkCompanyCode () {
    this.companyCodeErrors = []
    this.companyCode = this.companyCode.toLowerCase()
    if (this.companyCode.length === 0) {
      return
    }
    if (this.companyCode.length > 32) {
      this.companyCodeErrors.push('Please choose a company code that is at most 32 characters')
    }
    const regex = new RegExp('^([a-z])+(-?[a-z])*$')
    if (!regex.test(this.companyCode)) {
      this.companyCodeErrors.push('Please choose a company code that contains character padded hyphens and no special characters')
    }
  }

  private checkNewPassword () {
    this.newPasswordErrors = []
    this.validPassword = false
    if (this.newPassword.length === 0) {
      return
    }
    if (this.newPassword.length < 8) {
      this.newPasswordErrors.push('Please choose a password that is at least 8 characters')
    }
    if (this.newPassword.length > 32) {
      this.newPasswordErrors.push('Please choose a password that is at most 32 characters')
    }
    let regex = new RegExp('[0-9]')
    if (!regex.test(this.newPassword)) {
      this.newPasswordErrors.push('Please choose a password that contains at least 1 number')
    }
    regex = new RegExp('[a-z]')
    if (!regex.test(this.newPassword)) {
      this.newPasswordErrors.push('Please choose a password that contains at least 1 lower case letter')
    }
    regex = new RegExp('[A-Z]')
    if (!regex.test(this.newPassword)) {
      this.newPasswordErrors.push('Please choose a password that contains at least 1 upper case letter')
    }
    regex = new RegExp('[*.!@$%^#&:;,.?/~_|]')
    if (!regex.test(this.newPassword)) {
      this.newPasswordErrors.push('Please choose a password that contains at least 1 special character: *.!@$%^#&:;,.?/~_|')
    }
    if (this.newPasswordErrors.length === 0) {
      this.validPassword = true
    }
  }

  private checkConfirmPassword () {
    this.confirmPasswordErrors = []
    this.validPasswordForm = false || this.newsletterConsent
    if (this.confirmPassword.length === 0) {
      return
    }
    if (this.newPassword !== this.confirmPassword) {
      this.confirmPasswordErrors.push('Confirmation password does not match')
    }
    if (this.confirmPasswordErrors.length === 0) {
      this.validPasswordForm = true
    }
  }

  private updateAccountDetails () {
    let changed = false
    const newsletter = this.$store.getters.userProfile.newsletterConsent
    const options: any = {
      first_name: '',
      last_name: '',
      newsletter: newsletter
    }
    if (this.validUserDetails && (this.$store.getters.userProfile.firstName !== this.firstName || this.$store.getters.userProfile.lastName !== this.lastName)) {
      options.first_name = this.firstName
      options.last_name = this.lastName
      changed = true
    }
    if (newsletter !== this.newsletterConsent) {
      options.newsletter = this.newsletterConsent
      changed = true
    }

    if (!changed) {
      return
    }

    this.$apiService
      .updateAccountDetails(options)
      .then(() => {
        // TODO: refreshToken returns a promise that should be used to optimize the loading of new tabs
        this.$authService.refreshToken()
        this.$refs.accountResponseAlert.setMessage('Account details updated successfully')
        this.$refs.accountResponseAlert.setAlertType(AlertType.SUCCESS)
        setTimeout(() => {
          if (this.$refs.accountResponseAlert) {
            this.$refs.accountResponseAlert.resetAlert()
          }
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the user account settings')
        console.log(error)
        this.firstName = this.$store.getters.userProfile.firstName
        this.lastName = this.$store.getters.userProfile.lastName
        this.newsletterConsent = this.$store.getters.userProfile.newsletterConsent
        this.$refs.accountResponseAlert.setMessage('Failed to update account details')
        this.$refs.accountResponseAlert.setAlertType(AlertType.ERROR)
        setTimeout(() => {
          if (this.$refs.accountResponseAlert) {
            this.$refs.accountResponseAlert.resetAlert()
          }
        }, 5000)
      })
  }

  private setupCompanyAccount () {
    // Check for a valid company info form that is not equal to what is currently there. IE someone assigned to a company wants to update their newsletter settings but not change their company info
    if (!this.validCompanyInfo || (this.$store.getters.userProfile.firstName === '' && this.$store.getters.userProfile.lastName === '')) {
      this.$refs.companyResponseAlert.setMessage('Please update your first and last name before setting up a company account')
      this.$refs.companyResponseAlert.setAlertType(AlertType.ERROR)
      setTimeout(() => {
        if (this.$refs.companyResponseAlert) {
          this.$refs.companyResponseAlert.resetAlert()
        }
      }, 5000)
      return
    }

    this.$apiService
      .setupCompanyAccount({ company_name: this.companyName, company_code: this.companyCode })
      .then(() => {
        // TODO: refreshToken returns a promise that should be used to optimize the loading of new tabs
        this.$authService.refreshToken()
        this.$refs.companyResponseAlert.setMessage('Account settings updated successfully')
        this.$refs.companyResponseAlert.setAlertType(AlertType.SUCCESS)
        setTimeout(() => {
          if (this.$refs.companyResponseAlert) {
            this.$refs.companyResponseAlert.resetAlert()
          }
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the account settings')
        console.log(error)
        this.companyName = ''
        this.companyCode = ''
        this.$refs.companyResponseAlert.setMessage('Failed to update company details')
        this.$refs.companyResponseAlert.setAlertType(AlertType.ERROR)
        setTimeout(() => {
          if (this.$refs.companyResponseAlert) {
            this.$refs.companyResponseAlert.resetAlert()
          }
        }, 5000)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  #account-settings-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #account-settings-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
</style>
