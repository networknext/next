<template>
  <div class="card-body">
    <div id="auto-signin">
      <h5 class="card-title">
        Automatic Sign up
      </h5>
      <p class="card-text">
        Save time by allowing users with verified email addresses automatic access to your Network Next account.
      </p>
      <Alert :message="messages.autoDomains" :alertType="alertTypes.autoDomains" v-if="messages.autoDomains !== ''"/>
      <form v-on:submit.prevent="saveAutoSignUp()">
        <div class="form-group">
          <label for="auto-signup-domains">
            Automatic Sign up Domains
          </label>
          <textarea class="form-control form-control-sm" id="auto-signup-domains" v-model="autoSignupDomains"></textarea>
          <small class="form-text text-muted">
            Setting this to a comma seperated list of email domains will allow anyone with that domain to assign themselves to your account using your company code ({{ companyCode }}) in the account settings page.
          </small>
        </div>
        <button type="submit" class="btn btn-primary btn-sm">
          Save Automatic Sign up
        </button>
        <p class="text-muted text-small mt-2"></p>
      </form>
      <hr class="mt-4 mb-4">
  </div>
    <h5 class="card-title">
      Add new users
    </h5>
    <p class="card-text">
      Provide other people with access to your account.
    </p>
    <div id="add-user-spinner" v-show="false">
      <div class="spinner-border" role="status">
        <span class="sr-only">Loading...</span>
      </div>
    </div>
    <Alert :message="messages.newUsers" :alertType="alertTypes.newUsers" v-if="messages.newUsers !== ''"/>
    <form v-show="true" @submit.prevent="addNewUsers()">
      <div class="form-group">
        <label for="customerId">
          Add users by email address
        </label>
        <textarea class="form-control form-control-sm" id="new-user-emails" v-model="newUserEmails"></textarea>
        <small class="form-text text-muted">
          Enter a newline or comma-delimited list of email
          addresses to add users to your account.
        </small>
      </div>
      <div class="form-group">
        <label for="customerId">
          Permission Level
        </label>
        <multiselect placeholder="" track-by="id" label="id" v-model="newUserRoles" :options="allRoles" multiple></multiselect>
        <small class="form-text text-muted">
          The permission level to grant the added user accounts.
        </small>
      </div>
      <button type="submit" class="btn btn-primary btn-sm">
        Add Users
      </button>
      <p class="text-muted text-small mt-2"></p>
    </form>
    <hr class="mt-4 mb-4">
    <h5 class="card-title">
      Manage existing users
    </h5>
    <p class="card-text">
      Manage the list of users that currently have access to your Network Next account.
    </p>
    <Alert :message="messages.editUser" :alertType="alertTypes.editUser" v-if="messages.editUser !== ''"/>
    <table class="table table-sm mt-4">
      <thead class="thead-light">
        <tr>
          <th style="width: 20%;">
            Email Address
          </th>
          <th style="width: 70%;">
            Permissions
          </th>
          <th style="width: 10%;">
            Actions
          </th>
        </tr>
      </thead>
      <tbody v-if="companyUsers.length === 0">
        <tr>
          <td colspan="7" class="text-muted">
              There are no users assigned to your company.
          </td>
        </tr>
      </tbody>
      <tbody v-if="companyUsers.length > 0">
        <tr v-for="(account, index) in companyUsers" :key="index">
          <td>
            {{ account.email }}
          </td>
          <td>
            <multiselect placeholder="" track-by="id" label="id" v-model="account.roles" :options="allRoles" multiple :disabled="!account.edit"></multiselect>
          </td>
          <td class="td-btn" v-show="!account.edit && !account.delete">
            <button
              class="btn btn-xs btn-primary"
              data-toggle="tooltip"
              data-placement="bottom"
              title="Change this user's permissions"
              @click="editUser(account, index)"
            >
              <font-awesome-icon icon="pen"
                                  class="fa-w-16 fa-fw"
              />
            </button>&nbsp;
            <button
              class="btn btn-xs btn-danger"
              data-toggle="tooltip"
              data-placement="bottom"
              title="Remove this user"
              @click="deleteUser(account, index)"
            >
              <font-awesome-icon icon="trash"
                                  class="fa-w-16 fa-fw"
              />
            </button>&nbsp;
          </td>
          <td class="td-btn" v-show="account.edit || account.delete">
            <button
              class="btn btn-xs btn-success"
              data-toggle="tooltip"
              data-placement="bottom"
              title="Save Changes"
              @click="saveUser(account, index)"
            >
              <font-awesome-icon icon="check"
                                  class="fa-w-16 fa-fw"
              />
            </button>&nbsp;
            <button
              class="btn btn-xs btn-secondary"
              data-toggle="tooltip"
              data-placement="bottom"
              title="Cancel Changes"
              @click="cancel(index)"
            >
              <font-awesome-icon icon="times"
                                  class="fa-w-16 fa-fw"
              />
            </button>&nbsp;
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Multiselect from 'vue-multiselect'
import Alert from './Alert.vue'
import { AlertTypes } from './types/AlertTypes'
import _, { cloneDeep } from 'lodash'
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
    Alert,
    Multiselect
  }
})
export default class UserManagement extends Vue {
  private allRoles: Array<any>
  private companyUsers: Array<any>
  private companyUsersReadOnly: Array<any>
  private newUserRoles: Array<any>
  private messages: any
  private alertTypes: any
  private showTable: boolean
  private newUserEmails: string
  private autoSignupDomains: string
  private unwatch: any
  private companyCode: string
  private userProfile: UserProfile

  constructor () {
    super()
    this.newUserEmails = ''
    this.showTable = false
    this.messages = {
      autoDomains: '',
      newUsers: '',
      editUser: ''
    }
    this.alertTypes = {
      autoDomains: '',
      newUsers: '',
      editUser: ''
    }
    this.allRoles = []
    this.newUserRoles = []
    this.companyUsers = []
    this.companyUsersReadOnly = []
    this.companyCode = ''
    this.autoSignupDomains = ''
    this.userProfile = {} as UserProfile
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

  private destory () {
    this.unwatch()
  }

  private checkUserProfile (userProfile: UserProfile) {
    if (!userProfile) {
      return
    }
    this.companyCode = userProfile.companyCode || ''
    this.autoSignupDomains = userProfile.domains.join(', ')
    const promises = [
      // TODO: Figure out how to get rid of this. this.$apiService should be possible...
      // HACK: This is a hack to get tests to work properly
      (this as any).$apiService.fetchAllAccounts(),
      (this as any).$apiService.fetchAllRoles()
    ]
    Promise.all(promises)
      .then((responses: any) => {
        this.allRoles = responses[1].roles
        console.log(this.allRoles)
        this.companyUsers = responses[0].accounts
        this.companyUsers.forEach((user: any) => {
          user.edit = false
          user.delete = false
        })
        this.companyUsersReadOnly = _.cloneDeep(this.companyUsers)
      })
    this.userProfile = cloneDeep(this.$store.getters.userProfile)
  }

  private editUser (account: any, index: number): void {
    this.setAccountState(true, false, account, index)
  }

  private saveAutoSignUp (): void {
    const domains = this.autoSignupDomains
      .split(/(,|\n)/g)
      .map((x) => x.trim())
      .filter((x) => x !== '' && x !== ',');

    // TODO: Figure out how to get rid of this. this.$apiService should be possible...
    // HACK: This is a hack to get tests to work properly
    (this as any).$apiService
      .updateAutoSignupDomains({ domains: domains })
      .then((response: any) => {
        this.userProfile.domains = domains
        this.$store.commit('UPDATE_USER_PROFILE', this.userProfile)
        this.alertTypes.autoDomains = AlertTypes.SUCCESS
        this.messages.autoDomains = 'Successfully update signup domains'
        setTimeout(() => {
          this.messages.autoDomains = ''
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong adding auto signup domains')
        console.log(error)
        this.alertTypes.autoDomains = AlertTypes.ERROR
        this.messages.autoDomains = 'Failed to edit user account'
        setTimeout(() => {
          this.messages.autoDomains = ''
        }, 5000)
      })
  }

  private saveUser (account: any, index: number): void {
    if (account.edit) {
      // HACK because eslint likes to complain...
      // TODO: Figure out how to get rid of this. this.$apiService should be possible...
      // HACK: This is a hack to get tests to work properly
      const vm = (this as any)
      const roles = account.roles
      vm.$apiService
        .updateUserRoles({ user_id: `auth0|${account.user_id}`, roles: roles })
        .then((response: any) => {
          account.roles = response.roles
          this.alertTypes.editUser = AlertTypes.SUCCESS
          this.messages.editUser = 'User account edited successfully'
          setTimeout(() => {
            this.messages.editUser = ''
          }, 5000)
        })
        .catch((error: Error) => {
          console.log('Something went wrong updating the users permissions')
          console.log(error)
          this.alertTypes.editUser = AlertTypes.ERROR
          this.messages.editUser = 'Failed to edit user account'
          setTimeout(() => {
            this.messages.editUser = ''
          }, 5000)
        })
        .finally(() => {
          this.setAccountState(false, false, account, index)
        })
      return
    }
    if (account.delete) {
      // TODO: Figure out how to get rid of this. this.$apiService should be possible...
      // HACK: This is a hack to get tests to work properly
      (this as any).$apiService
        .deleteUserAccount({ user_id: `auth0|${account.user_id}` })
        .then((response: any) => {
          this.companyUsers.splice(index, 1)
          this.alertTypes.editUser = AlertTypes.SUCCESS
          this.messages.editUser = 'User account deleted successfully'
          setTimeout(() => {
            this.messages.editUser = ''
          }, 5000)
        })
        .catch((error: Error) => {
          console.log('Something went wrong updating the users permissions')
          console.log(error)
          this.alertTypes.newUsers = AlertTypes.ERROR
          this.messages.newUsers = 'Failed to delete user account'
          setTimeout(() => {
            this.messages.newUsers = ''
          }, 5000)
        })
    }
  }

  private deleteUser (account: any, index: number): void {
    this.setAccountState(false, true, account, index)
  }

  private cancel (index: number): void {
    const defaultUserAccount = _.cloneDeep(this.companyUsersReadOnly[index])
    this.companyUsers.splice(index, 1, defaultUserAccount)
  }

  private setAccountState (isEdit: boolean, isDelete: boolean, account: any, index: number) {
    account.edit = isEdit
    account.delete = isDelete
    this.companyUsers.splice(index, 1, account)
  }

  private addNewUsers (): void {
    let roles = this.newUserRoles
    const emails = this.newUserEmails
      .split(/(,|\n)/g)
      .map((x) => x.trim())
      .filter((x) => x !== '' && x !== ',')

    if (this.newUserRoles.length === 0) {
      roles = [{
        description: 'Can see current sessions and the map.',
        id: 'rol_ScQpWhLvmTKRlqLU',
        name: 'Viewer'
      }]
    }
    // TODO: Figure out how to get rid of this. this.$apiService should be possible...
    // HACK: This is a hack to get tests to work properly
    (this as any).$apiService
      .addNewUserAccounts({ emails: emails, roles: roles })
      .then((response: any) => {
        const newAccounts: Array<any> = response.accounts

        newAccounts.forEach((account: any) => {
          account.edit = false
          account.delete = false
        })

        this.companyUsers = this.companyUsers.concat(newAccounts)
        this.alertTypes.newUsers = AlertTypes.SUCCESS
        this.messages.newUsers = 'User account(s) added successfully'
        setTimeout(() => {
          this.messages.newUsers = ''
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong creating new users')
        console.log(error)
        this.alertTypes.newUsers = AlertTypes.ERROR
        this.messages.newUsers = 'Failed to add user account(s)'
        setTimeout(() => {
          this.messages.newUsers = ''
        }, 5000)
      })
    this.newUserRoles = []
    this.newUserEmails = ''
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
