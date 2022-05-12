<template>
  <div>
    <div
      class="spinner-border"
      role="status"
      id="sessions-spinner"
      v-show="!showSessions"
    >
      <span class="sr-only">Loading...</span>
    </div>
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
          <tr id="data-row" v-for="(session, index) in sessions" :key="index">
            <td>
              {{ convertUTCDateToLocalDate(new Date(session.time_stamp)) }}
            </td>
            <td>
                <router-link :to="`/session-tool/${session.meta.id}`" class="text-dark fixed-width">{{ session.meta.id }}</router-link>
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
      <div v-if="!$flagService.isEnabled(FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT)">
        <div class="float-left" style="padding-bottom: 20px;">
          <button id="more-sessions-button" class="btn btn-primary" @click="reloadSessions()">
            Refresh Sessions
          </button>
        </div>
        <div class="float-right" style="padding-bottom: 20px;">
          <button id="more-sessions-button" class="btn btn-primary" v-if="this.currentPage < MAX_PAGES" @click="fetchMoreSessions()">
            More Sessions
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { NavigationGuardNext, Route } from 'vue-router'
import { MAX_USER_SESSION_PAGES } from './types/Constants'
import { FeatureEnum } from './types/FeatureTypes'

/**
 * This component displays all of the information related to the user
 *  tool page in the Portal and has all the associated logic and api calls
 */

@Component
export default class UserSessions extends Vue {
  private sessions: Array<any>
  private showSessions: boolean
  private searchID: string
  private currentPage: number
  private unwatchFilter: any
  private FeatureEnum: any
  private MAX_PAGES = MAX_USER_SESSION_PAGES

  constructor () {
    super()
    this.searchID = ''
    this.sessions = []
    this.showSessions = false
    this.currentPage = 0
    this.FeatureEnum = FeatureEnum
  }

  private mounted () {
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.reloadSessions()
      }
    )

    this.currentPage = 0
    this.searchID = this.$route.params.pathMatch || ''
    if (this.searchID !== '') {
      this.fetchUserSessions()
    }
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT)) {
      next()
      return
    }
    this.showSessions = false
    this.searchID = to.params.pathMatch || ''
    this.currentPage = 0
    if (this.searchID !== '') {
      this.fetchUserSessions()
    }
    next()
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchMoreSessions () {
    this.fetchUserSessions()
  }

  private reloadSessions () {
    this.currentPage = 0
    this.showSessions = false
    this.sessions = []
    this.fetchUserSessions()
  }

  private fetchUserSessions () {
    if (this.searchID === '') {
      return
    }

    this.$apiService.fetchUserSessions({
      user_id: this.searchID,
      page: this.currentPage,
      timeframe: this.$store.getters.currentFilter.dateRange,
      customer_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.sessions = this.sessions.concat(response.sessions)
        this.currentPage = response.page
      })
      .catch((error: Error) => {
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
