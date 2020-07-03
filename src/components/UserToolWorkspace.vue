<template>
  <main role="main" class="col-md-12 col-lg-12 px-4">
    <div class="
              d-flex
              justify-content-between
              flex-wrap
              flex-md-nowrap
              align-items-center
              pt-3
              pb-2
              mb-3
              border-bottom
            "
    >
      <h1 class="h2">
        User Tool
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
        <div class="mr-auto"></div>
      </div>
    </div>
    <form class="flow-stats-form">
      <div class="form-group">
        <label for="user-hash-input">
            User Hash
        </label>
        <div class="row">
          <div class="col">
            <input class="form-control"
                    type="text"
                    placeholder="Enter a User Hash to view statistics"
            >
          </div>
          <div class="col-auto">
            <button class="btn btn-primary" type="submit">
              View Sessions
            </button>
          </div>
        </div>
      </div>
    </form>
    <div class="alert alert-info" role="alert" id="user-tool-alert" v-if="false">
        Please enter a User ID or Hash to view their sessions.
    </div>
    <div class="alert alert-danger" role="alert" id="user-tool-danger" v-if="false">
        Failed to fetch user sessions
    </div>
    <div id="user-sessions">
      <div class="table-responsive table-no-top-line">
        <table class="table table-sm" v-if="true">
          <thead>
            <tr>
              <th>
                <span>
                  Session ID
                </span>
              </th>
              <th>
                <span>
                  Platform
                </span>
              </th>
              <th>
                <span>
                  Connection Type
                </span>
              </th>
              <th>
                <span>
                  ISP
                </span>
              </th>
              <th>
                <span>
                  Datacenter
                </span>
              </th>
              <th>
                <span>
                  Server Address
                </span>
              </th>
            </tr>
          </thead>
          <tbody v-if="true">
            <!-- <tr v-for="session in pages.userTool.sessions"> -->
            <tr>
              <td>
                <a class="text-dark fixed-width" href="#">
                    SESSION ID
                </a>
              </td>
              <td>
                SESSION PLATFORM
              </td>
              <td>
                SESSION CONNECTION
              </td>
              <td>
                SESSION LOCATION ISP
              </td>
              <td>
                SESSION DATACENTER
              </td>
              <td>
                SESSION SERVER ADDR
              </td>
            </tr>
          </tbody>
          <tbody v-if="false">
            <tr>
              <td colspan="7" class="text-muted">
                  There are no sessions belonging to this user.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class UserToolWorkspace extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
