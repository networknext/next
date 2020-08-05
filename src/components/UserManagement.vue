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
        <multiselect track-by="name" label="name" v-model="newUserRoles" :options="allRoles" multiple></multiselect>
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
          <td>
            {{ account.email }}
          </td>
          <td>
            <multiselect track-by="name" label="name" v-model="selectedRoles[account.user_id]" :options="allRoles" multiple :disabled="!account.edit"></multiselect>
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
              @click="cancel(account, index)"
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

@Component({
  components: {
    Multiselect
  }
})
export default class UserManagement extends Vue {
  // TODO: Fix weird issue with dropdown library change events (select/delete) handler
  private apiService: APIService
  private allRoles: Array<any> = []
  private companyUsers: Array<any> = []

  private selectedRoles: any = {}
  private newUserRoles: any = []

  private newUserEmails = ''

  private showTable = false

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
  }

  private mounted () {
    const promises = [
      this.apiService.fetchAllAccounts({}),
      this.apiService.fetchAllRoles()
    ]
    Promise.all(promises)
      .then((responses: any) => {
        const companyUsers: Array<any> = responses[0].result.accounts
        const allRoles = responses[1].result.roles

        this.allRoles = allRoles
        this.companyUsers = companyUsers
        this.companyUsers.forEach((user: any) => {
          user.edit = false
          user.delete = false
        })

        this.companyUsers.forEach((user: any) => {
          this.selectedRoles[user.user_id] = user.roles
        })
        this.showTable = true
      })
  }

  private editUser (account: any, index: number) {
    setTimeout(() => {
      account.delete = false
      account.edit = true
      this.companyUsers.splice(index, 1, account)
    })
  }

  private saveUser (account: any, index: number) {
    if (account.edit) {
      const roles = this.selectedRoles[account.user_id]
      this.apiService
        .updateUserRoles({ user_id: `auth0|${account.user_id}`, roles: roles })
        .then((response: any) => {
          account.roles = response.roles
        })
        .catch((error: Error) => {
          console.log('Something went wrong updating the users permissions')
          console.log(error)
        })
        .finally(() => {
          this.cancel(account, index)
        })
      return
    }
    if (account.delete) {
      this.apiService
        .deleteUserAccount({ user_id: `auth0|${account.user_id}` })
        .then((response: any) => {
          this.companyUsers.splice(index, 1)
          this.selectedRoles[account.user_id] = null
        })
        .catch((error: Error) => {
          console.log('Something went wrong updating the users permissions')
          console.log(error)
        })
    }
  }

  private deleteUser (account: any, index: number) {
    account.delete = true
    account.edit = false
    this.companyUsers.splice(index, 1, account)
  }

  private cancel (account: any, index: number) {
    account.delete = false
    account.edit = false
    this.companyUsers.splice(index, 1, account)
  }

  private addNewUsers () {
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
          this.selectedRoles[account.user_id] = account.roles
        })

        this.companyUsers.concat(newAccounts)
      })
      .catch((error: Error) => {
        console.log('Something went wrong creating new users')
        console.log(error)
      })
    this.newUserRoles = []
    this.newUserEmails = ''
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
