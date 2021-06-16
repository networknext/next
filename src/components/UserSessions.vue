<template>
  <div class="table-responsive table-no-top-line" v-if="showSessions">
    <table class="table table-sm">
      <thead>
        <tr>
          <th>
            <span>
              Date
            </span>
          </th>
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
        <tr id="data-row" v-for="(session, index) in sessions" v-bind:key="index">
          <td>
            {{ convertUTCDateToLocalDate(new Date(session.time_stamp)) }}
          </td>
          <td>
              <router-link v-bind:to="`/session-tool/${session.meta.id}`" class="text-dark fixed-width">{{ session.meta.id }}</router-link>
          </td>
          <td>
            {{ session.meta.platform }}
          </td>
          <td>
            {{ session.meta.connection === "wifi" ? "Wi-Fi" : session.meta.connection.charAt(0).toUpperCase() + session.meta.connection.slice(1) }}
          </td>
          <td>
            {{ session.meta.location.isp || "Unknown"}}
          </td>
          <td>
            {{ session.meta.datacenter_alias !== "" ? session.meta.datacenter_alias : session.meta.datacenter_name }}
          </td>
          <td>
            {{ session.meta.server_addr }}
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
    <div class="float-right">
      <button id="more-sessions-button" class="btn btn-primary" v-if="currentPage < MAX_PAGES" @click="fetchMoreSessions()">
        More Sessions
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { NavigationGuardNext, Route } from 'vue-router'

const MAX_PAGES = 10

/**
 * This component displays all of the information related to the user
 *  tool page in the Portal and has all the associated logic and api calls
 */

@Component
export default class UserSessions extends Vue {
  private sessions: Array<any>
  private timeStamps: Array<any>
  private sessionLoop: any
  private showSessions: boolean
  private searchID: string
  private currentPage: number
  private savedPage: number

  constructor () {
    super()
    this.searchID = ''
    this.sessions = []
    this.timeStamps = []
    this.showSessions = false
    this.sessionLoop = null
    this.currentPage = 0
    this.savedPage = 0
  }

  private mounted () {
    this.currentPage = 0
    this.searchID = this.$route.params.pathMatch || ''
    if (this.searchID !== '') {
      this.restartLoop()
    }
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    if (this.sessionLoop) {
      clearInterval(this.sessionLoop)
    }
    this.showSessions = false
    this.searchID = to.params.pathMatch || ''
    this.currentPage = 0
    if (this.searchID !== '') {
      this.fetchUserSessions()
      this.sessionLoop = setInterval(() => {
        this.fetchUserSessions()
      }, 10000)
    }
    next()
  }

  private beforeDestroy () {
    clearInterval(this.sessionLoop)
  }

  private fetchMoreSessions () {
    const nextPage = this.currentPage + 1

    if (nextPage >= MAX_PAGES) {
      // This shouldn't happen
      return
    }
    this.currentPage = nextPage
    this.savedPage = this.currentPage
    this.restartLoop()
  }

  private fetchUserSessions () {
    if (this.searchID === '') {
      return
    }

    this.$apiService.fetchUserSessions({ user_id: this.searchID, page: this.savedPage })
      .then((response: any) => {
        this.sessions = response.sessions || []
        this.showSessions = true
        this.currentPage = response.page
      })
      .catch((error: Error) => {
        if (this.sessionLoop) {
          clearInterval(this.sessionLoop)
        }
        if (this.sessions.length === 0) {
          console.log(`Something went wrong fetching sessions details for: ${this.searchID}`)
          console.log(error)
        }
      })
      .finally(() => {
        if (!this.showSessions) {
          this.showSessions = true
        }
      })
  }

  private restartLoop () {
    if (this.sessionLoop) {
      clearInterval(this.sessionLoop)
    }
    this.fetchUserSessions()
    this.sessionLoop = setInterval(() => {
      this.fetchUserSessions()
    }, 10000)
  }

  private convertUTCDateToLocalDate (date: Date) {
    const newDate = new Date(date.getTime() + date.getTimezoneOffset() * 60 * 1000)

    newDate.setMinutes(date.getMinutes() - date.getTimezoneOffset())

    return newDate.toLocaleString().replace(',', '')
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  #more-sessions-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
</style>
