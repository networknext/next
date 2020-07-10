<template>
  <div class="card-body">
    <div id="auto-signin" class="hidden">
      <h5 class="card-title">
        Automatic sign-in
      </h5>
      <p class="card-text">
        Save time by allowing users with verified email
        addresses automatic access to your Network Next account.
      </p>
      <div id="auto-sign-in-spinner" v-show="false">
        <div class="spinner-border" role="status">
          <span class="sr-only">Loading...</span>
        </div>
      </div>
      <form v-show="true">
        <div class="form-group">
          <label for="customerId">
            Automatic Sign-in Domain
          </label>
          <input type="text"
                  class="form-control form-control-sm"
                  value="networknext.com"
                  id="auto-sign-in-domain"
          >
          <small class="form-text text-muted">
            If you set this to a domain, such as "example.com",
            then anyone with a verified email address with
            this domain will automatically
            join your account when logging in.
            You can set this so that you don't need to
            manually onboard everyone in your organization.
          </small>
        </div>
        <div class="form-group">
          <label for="customerId">
            Permission Level
          </label>
          <select
            class="form-control"
            id="auto-signin-permissions"
            multiple
          >
          </select>
          <small class="form-text text-muted">
            The permission level to grant accounts that
            join your account via automatic sign-in.
          </small>
        </div>
        <button type="submit" class="btn btn-primary btn-sm">
          Save Automatic Sign-in
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
        <!-- Replace these with a new dropdown component -->
        <select
          class="form-control"
          id="add-user-permissions"
          placeholder="Select..."
          multiple
        >
        </select>
        <!-- Replace these with a new dropdown component -->
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
        <!-- <tr v-for="(account, index) in pages.settings.accounts"> -->
        <tr>
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
          <!-- Replace these with a new dropdown component -->
            <!-- <select
              class="form-control"
              :id="`edit-user-permissions-${account.user_id}`"
              multiple
            >
            </select> -->
          <!-- Replace these with a new dropdown component -->
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

@Component
export default class UserManagement extends Vue {
  private switchToGameConfiguration () {
    // Empty for now
    this.$router.push('/user-tool/game-config')
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
