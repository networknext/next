<template>
  <div class="card-body">
    <h5 class="card-title">
      Account Settings
    </h5>
    <p class="card-text">
      Update user account profile.
    </p>
    <Alert :message="message" :alertType="alertType" v-if="message !== ''"/>
    <form @submit.prevent="updateCompanyName()">
      <div class="form-group">
        <label for="companyName">
          Company Name
        </label>
        <input type="text" class="form-control form-control-sm" id="companyName" v-model="companyName" placeholder="Enter your company name"/>
        <small class="form-text text-muted">
          This is the company that you would like your account to be assigned to. Case and white space sensitive.
        </small>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" v-bind:disabled="companyName.length === 0">
        Update Company Name
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
    <form @submit.prevent="updateAccountSettings()">
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
      <div class="form-group" v-if="validPassword">
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
      <div class="form-check">
        <input type="checkbox" class="form-check-input" id="newsletterConsent" v-model="newsletterConsent" @change="checkConfirmPassword()"/>
        <small>
          I would like to recieve the Network Next newsletter
        </small>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" v-bind:disabled="!validForm" style="margin-top: 1rem;">
        Save
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
  </div>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash'
import { Component, Vue } from 'vue-property-decorator'
import Alert from './Alert.vue'
import { AlertTypes } from './types/AlertTypes'
import { UserProfile } from './types/AuthTypes'

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
  private message: any
  private alertType: any
  private companyName: string
  private newPassword: string
  private confirmPassword: string
  private unwatch: any
  private validPassword: boolean
  private validForm: boolean
  private newPasswordErrors: Array<string>
  private confirmPasswordErrors: Array<string>
  private newsletterConsent: boolean

  constructor () {
    super()
    this.message = ''
    this.alertType = AlertTypes.DEFAULT
    this.companyName = ''
    this.newPassword = ''
    this.confirmPassword = ''
    this.validPassword = false
    this.validForm = false
    this.newPasswordErrors = []
    this.confirmPasswordErrors = []
    this.newsletterConsent = false
  }

  private mounted () {
    if (!this.$store.getters.userProfile) {
      this.unwatch = this.$store.watch(
        (_, getters: any) => getters.userProfile,
        (userProfile: any) => {
          this.checkUserProfile(userProfile)
        }
      )
    } else {
      this.checkUserProfile(this.$store.getters.userProfile)
    }
  }

  private checkUserProfile (userProfile: UserProfile) {
    if (this.companyName === '') {
      this.companyName = userProfile.company || ''
    }
    this.newsletterConsent = userProfile.newsletterConsent || false
  }

  private destory () {
    this.unwatch()
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
    this.validForm = false || this.newsletterConsent
    if (this.confirmPassword.length === 0) {
      return
    }
    if (this.newPassword !== this.confirmPassword) {
      this.confirmPasswordErrors.push('Confirmation password does not match')
    }
    if (this.confirmPasswordErrors.length === 0) {
      this.validForm = true
    }
  }

  private updateCompanyName () {
    (this as any).$apiService
      .updateCompanyName({ company_name: this.companyName })
      .then((response: any) => {
        const roles = response.new_roles
        const companyName = response.company_name
        const userProfile: UserProfile = cloneDeep(this.$store.getters.userProfile)
        if (roles.length > 0) {
          userProfile.roles = roles
        }
        userProfile.company = companyName
        this.$store.commit('UPDATE_USER_PROFILE', userProfile)
        this.message = 'Company name updated successfully'
        this.alertType = AlertTypes.SUCCESS
        setTimeout(() => {
          this.message = ''
          this.alertType = AlertTypes.DEFAULT
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the account settings')
        console.log(error)
        this.message = 'Failed to update company name'
        this.alertType = AlertTypes.ERROR
        setTimeout(() => {
          this.message = ''
          this.alertType = AlertTypes.DEFAULT
        }, 5000)
      })
  }

  private updateAccountSettings () {
    (this as any).$apiService
      .updateAccountSettings({ newPassword: this.newPassword, newsletterConsent: this.newsletterConsent })
      .then((response: any) => {
        this.message = 'Account settings updated successfully'
        this.alertType = AlertTypes.SUCCESS
        setTimeout(() => {
          this.message = ''
          this.alertType = AlertTypes.DEFAULT
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the account settings')
        console.log(error)
        this.message = 'Failed to update account settings'
        this.alertType = AlertTypes.ERROR
        setTimeout(() => {
          this.message = ''
          this.alertType = AlertTypes.DEFAULT
        }, 5000)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
