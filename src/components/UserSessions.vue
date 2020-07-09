<template>
  <div class="table-responsive table-no-top-line" v-if="showSessions">
    <table class="table table-sm">
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
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * TODO: Cleanup template
 * TODO: Figure out what sessionMeta fields need to be required
 * TODO: Hookup API call
 * TODO: Hookup loop logic
 */

@Component
export default class UserSessions extends Vue {
  // TODO: Refactor out the alert/error into its own component.
  private searchID = ''
  private showSessions = false

  private created () {
    console.log('Creating User Sessions')
    if (this.searchID === '' && this.$route.params.pathMatch) {
      this.searchID = this.$route.params.pathMath
    }
    this.fetchUserSessions()
  }

  private fetchUserSessions () {
    // API Call to fetch the details associated to ID
    this.showSessions = true
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
