<template>
  <div class="card-body">
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
        <multiselect placeholder="" track-by="name" label="name" v-model="newUserRoles" :options="allRoles" multiple></multiselect>
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
    <div id="account-table-spinner" v-show="!showTable">
      <div class="spinner-border" role="status">
        <span class="sr-only">Loading...</span>
      </div>
    </div>
    <table class="table table-sm mt-4" v-show="showTable">
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
      <tbody>
        <tr v-for="(account, index) in companyUsers" :key="index">
          <Alert :message="messages.editUser" :alertType="alertTypes.newUser" v-if="messages.editUser !== ''"/>
          <td>
            {{ account.email }}
          </td>
          <td>
            <multiselect placeholder="" track-by="name" label="name" v-model="account.roles" :options="allRoles" multiple :disabled="!account.edit"></multiselect>
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
import APIService from '../services/api.service'
import Alert from './Alert.vue'
import { AlertTypes } from './types/AlertTypes'

import _ from 'lodash'

@Component({
  components: {
    Alert,
    Multiselect
  }
})
export default class UserManagement extends Vue {
  private apiService: APIService
  private allRoles: Array<any> = []
  private companyUsers: Array<any> = []
  private companyUsersReadOnly: Array<any> = []

  private newUserRoles: any = []

  private newUserEmails: string

  private showTable: boolean

  private messages: any
  private alertTypes: any

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
    this.newUserEmails = ''
    this.showTable = false
    this.messages = {
      newUsers: '',
      editUser: ''
    }
    this.alertTypes = {
      newUsers: '',
      editUser: ''
    }
  }

  private mounted (): void {
    const promises = [
      this.apiService.fetchAllAccounts({}),
      this.apiService.fetchAllRoles()
    ]
    Promise.all(promises)
      .then((responses: any) => {
        this.allRoles = responses[1].roles
        this.companyUsers = responses[0].accounts
        this.companyUsers.forEach((user: any) => {
          user.edit = false
          user.delete = false
        })
        this.companyUsersReadOnly = _.cloneDeep(this.companyUsers)
        this.showTable = true
      })
  }

  private editUser (account: any, index: number): void {
    this.setAccountState(true, false, account, index)
  }

  private saveUser (account: any, index: number): void {
    if (account.edit) {
      const roles = account.roles
      this.apiService
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
      this.apiService
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
    this.apiService
      .addNewUserAccounts({ emails: emails, roles: roles })
      .then((response: any) => {
        const newAccounts: Array<any> = response.accounts

        newAccounts.forEach((account: any) => {
          account.edit = false
          account.delete = false
        })

        this.companyUsers.concat(newAccounts)
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
