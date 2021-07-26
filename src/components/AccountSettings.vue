<template>
  <div class="card-body">
    <h5 class="card-title">
      Account Settings
    </h5>
    <p class="card-text">
      Update user account profile.
    </p>
    <Alert ref="responseAlert"/>
    <form @submit.prevent="updateAccountSettings()">
      <div class="form-group">
        <label for="companyName">
          Company Name
        </label>
        <input type="text" class="form-control form-control-sm" id="companyName" v-model="companyName" placeholder="Enter your company name" @change="checkCompanyName()"/>
        <small class="form-text text-muted">
          This is the company that you would like your account to be assigned to. Case and white space sensitive.
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
        <input type="text" class="form-control form-control-sm" id="companyCode" v-model="companyCode" placeholder="Enter your company code" @change="checkCompanyCode()"/>
        <small class="form-text text-muted">
          This is the unique string associated to your company account and to be used in your company subdomain. Examples: mycompany, my-company, my-company-name
        </small>
        <small v-for="(error, index) in companyCodeErrors" :key="index" class="text-danger">
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
        Update Company Settings
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
      <button type="submit" class="btn btn-primary btn-sm" v-bind:disabled="!validPasswordForm" style="margin-top: 1rem;" v-if="false">
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
    responseAlert: Alert;
  }

  get validCompanyInfo (): boolean {
    return this.validCompanyName && this.validCompanyCode
  }

  private companyName: string
  private companyCode: string
  private newPassword: string
  private confirmPassword: string
  private validPassword: boolean
  private validPasswordForm: boolean
  private validCompanyCode: boolean
  private validCompanyName: boolean
  private newPasswordErrors: Array<string>
  private companyNameErrors: Array<string>
  private companyCodeErrors: Array<string>
  private confirmPasswordErrors: Array<string>
  private newsletterConsent: boolean
  private AlertType: any

  constructor () {
    super()
    this.companyName = ''
    this.companyCode = ''
    this.newPassword = ''
    this.confirmPassword = ''
    this.validPassword = false
    this.validPasswordForm = false
    this.validCompanyName = false
    this.validCompanyCode = false
    this.validCompanyCode = false
    this.validCompanyName = false
    this.newPasswordErrors = []
    this.companyNameErrors = []
    this.companyCodeErrors = []
    this.confirmPasswordErrors = []
    this.newsletterConsent = false
    this.AlertType = AlertType
  }

  private mounted () {
    const userProfile = cloneDeep(this.$store.getters.userProfile)
    this.companyName = userProfile.companyName || ''
    this.companyCode = userProfile.companyCode || ''
    this.newsletterConsent = userProfile.newsletterConsent || false
    this.checkCompanyName()
    this.checkCompanyCode()
    this.checkConfirmPassword()
  }

  private checkCompanyName () {
    this.companyNameErrors = []
    this.validCompanyName = false
    if (this.companyName.length > 256) {
      this.companyNameErrors.push('Please choose a company name that is at most 256 characters')
    }
    if (this.companyNameErrors.length === 0) {
      this.validCompanyName = true
    }
  }

  private checkCompanyCode () {
    this.companyCodeErrors = []
    this.validCompanyCode = false
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
    if (this.companyCodeErrors.length === 0) {
      this.validCompanyCode = true
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

  private updateAccountSettings () {
    const promises = []
    // Check for a valid company info form that is not equal to what is currently there. IE someone assigned to a company wants to update their newsletter settings but not change their company info
    if (this.validCompanyInfo && this.$store.getters.userProfile.companyName !== this.companyName && this.$store.getters.userProfile.companyCode !== this.companyCode) {
      promises.push(
        this.$apiService
          .updateCompanyInformation({ company_name: this.companyName, company_code: this.companyCode })
      )
    }
    if (this.$store.getters.userProfile.newsletterConsent !== this.newsletterConsent) {
      promises.push(this.$apiService
        .updateAccountSettings({ newsletter: this.newsletterConsent }))
    }

    if (promises.length === 0) {
      return
    }

    Promise.all(promises)
      .then(() => {
        // TODO: refreshToken returns a promise that should be used to optimize the loading of new tabs
        this.$authService.refreshToken()
        this.$refs.responseAlert.setMessage('Account settings updated successfully')
        this.$refs.responseAlert.setAlertType(AlertType.SUCCESS)
        setTimeout(() => {
          if (this.$refs.responseAlert) {
            this.$refs.responseAlert.resetAlert()
          }
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the account settings')
        console.log(error)
        this.companyName = this.$store.getters.userProfile.companyName
        this.companyCode = this.$store.getters.userProfile.companyCode
        this.newsletterConsent = this.$store.getters.userProfile.newsletterConsent
        this.$refs.responseAlert.setMessage('Failed to update account settings')
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
