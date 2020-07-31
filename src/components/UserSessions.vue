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
      <tbody v-if="sessions.length > 0">
        <tr v-for="(session, index) in sessions" v-bind:key="index">
          <td>
              <router-link v-bind:to="`/session-tool/${session.id}`" class="text-dark fixed-width">{{ session.id }}</router-link>
          </td>
          <td>
            {{ session.platform }}
          </td>
          <td>
            {{ session.connection }}
          </td>
          <td>
            {{ session.location.isp }}
          </td>
          <td>
            {{ session.datacenter }}
          </td>
          <td>
            {{ session.server_addr }}
          </td>
        </tr>
      </tbody>
      <tbody v-if="sessions.length === 0">
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
import { Component, Vue, Prop } from 'vue-property-decorator'
import APIService from '../services/api.service'

/**
 * TODO: Cleanup template
 * TODO: Figure out what sessionMeta fields need to be required
 * TODO: Hookup API call
 * TODO: Hookup loop logic
 */

@Component
export default class UserSessions extends Vue {
  @Prop({ required: false, type: String, default: '' }) searchID!: string

  // TODO: Refactor out the alert/error into its own component.
  private apiService: APIService
  private showSessions = false
  private sessions = []

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
  }

  private mounted () {
    this.fetchUserSessions()
  }

  private fetchUserSessions () {
    if (this.searchID === '') {
      return
    }

    this.apiService.fetchUserSessions({ user_hash: this.searchID })
      .then((response: any) => {
        this.sessions = response.result.sessions || []
        this.showSessions = true
      })
      .catch((error: Error) => {
        console.log(error)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
