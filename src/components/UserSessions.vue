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
import { AlertTypes } from './types/AlertTypes'

/**
 * This component displays all of the information related to the user
 *  tool page in the Portal and has all the associated logic and api calls
 */

@Component
export default class UserSessions extends Vue {
  private sessions: Array<any>
  private sessionLoop: any
  private showSessions: boolean
  private searchID: string
  private message: string
  private alertType: string
  private vueInstance: any

  constructor () {
    super()
    this.searchID = ''
    this.sessions = []
    this.showSessions = false
    this.sessionLoop = null
    this.message = 'Failed to fetch user sessions'
    this.alertType = AlertTypes.ERROR
    this.vueInstance = Vue
  }

  private mounted () {
    this.searchID = this.$route.params.pathMatch || ''
    if (this.searchID !== '') {
      this.fetchUserSessions()
      this.sessionLoop = setInterval(() => {
        this.fetchUserSessions()
      }, 10000)
    }
  }

  private fetchUserSessions () {
    if (this.searchID === '') {
      return
    }

    this.vueInstance.fetchUserSessions({ user_hash: this.searchID })
      .then((response: any) => {
        this.sessions = response.sessions || []
        this.showSessions = true
      })
      .catch((error: Error) => {
        if (this.sessionLoop) {
          clearInterval(this.sessionLoop)
        }
        if (this.sessions.length === 0) {
          this.message = 'Failed to fetch session details'
          console.log(`Something went wrong fetching sessions details for: ${this.searchID}`)
          console.log(error)
        }
        console.log(error)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
