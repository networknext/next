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
    <div class="alert alert-success"
          role="alert"
          id="session-tool-alert"
          v-show="false">
      NEWUSER SUCCESS
    </div>
    <div class="alert alert-danger"
          role="alert"
          id="session-tool-alert"
          v-show="false">
      NEWUSER FAILURE
    </div>
    <form v-show="true">
      <div class="form-group">
        <label for="customerId">
          Add users by email address
        </label>
        <textarea class="form-control form-control-sm" id="new-user-emails"></textarea>
        <small class="form-text text-muted">
          Enter a newline or comma-delimited list of email
          addresses to add users to your account.
        </small>
      </div>
      <div class="form-group">
        <label for="customerId">
          Permission Level
        </label>
        <multiselect track-by="name" label="name" v-model="value" :options="options" multiple></multiselect>
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
    <div id="account-table-spinner" v-show="false">
      <div class="spinner-border" role="status">
        <span class="sr-only">Loading...</span>
      </div>
    </div>
    <table class="table table-sm mt-4" v-show="true">
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
          <div
            class="alert alert-success"
            role="alert"
            v-show="false"
          >
            UPDATEUSER SUCCESS
          </div>
          <div
              class="alert alert-danger"
              role="alert"
              v-show="false"
          >
            UPDATEUSER FAILURE
          </div>
          <td>
            EMAIL
          </td>
          <td>
            <multiselect track-by="name" label="name" v-model="value" :options="options" multiple></multiselect>
          </td>
          <td class="td-btn" v-show="true">
            <button
              class="btn btn-xs btn-primary"
              data-toggle="tooltip"
              data-placement="bottom"
              title="Change this user's permissions"
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
          <td class="td-btn" v-show="false">
            <button
              class="btn btn-xs btn-success"
              data-toggle="tooltip"
              data-placement="bottom"
              title="Save Changes"
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
import { UserAccount, Role } from './types/APITypes'

@Component({
  components: {
    Multiselect
  }
})
export default class UserManagement extends Vue {
  private allRoles: Array<Role> = []
  private companyUsers: Array<UserAccount> = []

  private value = [
  ]

  private options = [
    { name: 'Vue.js', code: 'vu' },
    { name: 'Javascript', code: 'js' },
    { name: 'Open Source', code: 'os' }
  ]

  private created () {
    console.log('User Management Created')

    // TODO: API call to get all role options
    this.allRoles = [
      {
        id: '1234',
        name: 'Admin',
        description: 'With great power comes great responsibility'
      }
    ]
    console.log(this.allRoles)
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
