<template>
  <div>
    <div class="row" style="text-align: center;" v-show="!showSessions">
      <div class="col"></div>
      <div class="col">
        <div
          class="spinner-border"
          role="status"
          id="customers-spinner"
          style="margin:1rem;"
        >
          <span class="sr-only">Loading...</span>
        </div>
      </div>
      <div class="col"></div>
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
        <tbody v-if="currentPageSessions.length > 0">
          <tr id="data-row" v-for="(session, index) in currentPageSessions" :key="index">
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
              {{ session.meta.datacenter_alias !== "" ? session.meta.datacenter_alias : session.meta.datacenter_name !== "" ? session.meta.datacenter_name : "Unknown" }}
            </td>
            <td>
              {{ session.meta.server_addr }}
            </td>
          </tr>
        </tbody>
        <tbody v-if="currentPageSessions.length === 0">
          <tr>
            <td colspan="7" class="text-muted">
                There are no sessions belonging to this user.
            </td>
          </tr>
        </tbody>
      </table>
      <nav v-if="$flagService.isEnabled(FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT) && currentPageSessions.length > 0">
        <div class="pagination-container">
          <div id="page-counter" class="col-auto">
            Total Pages: {{ numPages }}
          </div>
          <ul class="pagination justify-content-center col">
            <li id="previous-page-button" class="page-item" :class="{ disabled: (currentPage - 1) <= 0 }">
              <a class="page-link" @click.prevent="changePage(currentPage - 1)">Previous</a>
            </li>
            <li id="skip-backward-button" class="page-item" v-if="currentPage - (PAGINATION_RANGE + 1) > 0">
              <a class="page-link" @click.prevent="changePage(currentPage - (PAGINATION_RANGE + 1))">...</a>
            </li>
            <li class="page-item" v-for="index in oldPageNumbers" :key="oldIndexToPageNumber(index)">
              <a class="page-link" @click.prevent="changePage(oldIndexToPageNumber(index))">{{ oldIndexToPageNumber(index) }}</a>
            </li>
            <li class="page-item active">
              <a class="page-link">{{ currentPage }}</a>
            </li>
            <li class="page-item" v-for="index in newPageNumbers" :key="currentPage + index">
              <a class="page-link" @click.prevent="changePage(currentPage + index)">{{ currentPage + index }}</a>
            </li>
            <li id="skip-forward-button" class="page-item" v-if="currentPage + (PAGINATION_RANGE + 1) < numPages">
              <a class="page-link" @click.prevent="changePage(currentPage + (PAGINATION_RANGE + 1))">...</a>
            </li>
            <li id="next-page-button" class="page-item" :class="{ disabled: (currentPage + 1) > numPages }">
              <a class="page-link" @click.prevent="changePage(currentPage + 1)">Next</a>
            </li>
          </ul>
          <div class="col-auto" style="padding-top: .5rem;">Entries per page:</div>
          <div class="col-auto">
            <select id="per-page-dropdown" class="form-control" @change="updateEntriesPerPage($event.target.value)">
              <option v-for="option in entriesPerPageList" :key="option" :value="option" :selected="entriesPerPage === option">
                {{ option }}
              </option>
            </select>
          </div>
        </div>
      </nav>
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
  private PAGINATION_RANGE = 5

  private entriesPerPage: number
  private entriesPerPageList: Array<number>

  get oldPageNumbers () {
    if (this.currentPage - this.PAGINATION_RANGE > 0) {
      return this.PAGINATION_RANGE
    } else {
      return this.currentPage - 1
    }
  }

  get newPageNumbers () {
    return this.currentPage + this.PAGINATION_RANGE <= this.numPages - 1 ? this.PAGINATION_RANGE : this.numPages - this.currentPage
  }

  get currentPageSessions () {
    if (!this.$flagService.isEnabled(FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT)) {
      return this.sessions
    }

    if (this.readOnlySessions.length === 0) {
      return []
    }

    // StartIndex is the currentPage - 1 (pages start at 1 but index starts at 0) * the number of sessions per page
    // If we are on the first page, just use 0 as the start index
    const startIndex = this.currentPage === 1 ? 0 : (this.currentPage - 1) * this.entriesPerPage

    // EndIndex is the startIndex + the number of entries per page
    // If we overflow the number of entries with startIndex + number of entries per page, use (startIndex, this.readOnlySessions.length)
    const endIndex = startIndex + this.entriesPerPage > this.readOnlySessions.length ? this.readOnlySessions.length : startIndex + this.entriesPerPage

    const pageSessions = this.readOnlySessions.slice(startIndex, endIndex)

    // TODO: Change this when looker user tool goes live for everyone...
    return pageSessions
  }

  private sessions: Array<any>
  private readOnlySessions: Array<any>
  private showSessions: boolean
  private searchID: string
  private currentPage: number
  private numPages: number
  private unwatchFilter: any
  private FeatureEnum: any
  private MAX_PAGES = MAX_USER_SESSION_PAGES

  constructor () {
    super()
    this.searchID = ''
    this.sessions = []
    this.readOnlySessions = []
    this.showSessions = false
    this.currentPage = 0
    this.numPages = 0
    this.FeatureEnum = FeatureEnum

    this.entriesPerPageList = [
      10,
      25,
      50,
      100,
      200
    ]

    this.entriesPerPage = this.entriesPerPageList[0]
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
      timeframe: this.$store.getters.currentFilter.dateRange || '',
      customer_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.readOnlySessions = response.sessions || []

        if (this.$flagService.isEnabled(FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT)) {
          this.numPages = Math.ceil(this.readOnlySessions.length / this.entriesPerPage)
          this.currentPage = 1
        } else {
          this.sessions = this.sessions.concat(this.readOnlySessions)
          this.currentPage = response.page
        }
      })
      .catch((error: Error) => {
        this.sessions = []
        this.readOnlySessions = []
        this.numPages = 0

        console.log(`Something went wrong fetching user sessions for: ${this.searchID}`)
        console.log(error)
      })
      .finally(() => {
        if (!this.showSessions) {
          this.showSessions = true
        }
      })
  }

  private changePage (pageNumber: number) {
    if (this.showSessions) {
      this.showSessions = false
    }
    this.currentPage = pageNumber
    setTimeout(() => {
      this.showSessions = true
    }, 1000)
  }

  private updateEntriesPerPage (entries: string) {
    this.showSessions = false
    const entriesPerPage: number = parseInt(entries)
    this.entriesPerPage = entriesPerPage
    this.numPages = Math.ceil(this.readOnlySessions.length / this.entriesPerPage)
    this.changePage(1)
  }

  // Helper function to avoid clogging template with JS
  private oldIndexToPageNumber (index: number) {
    return this.currentPage - this.PAGINATION_RANGE <= 0 ? index : (this.currentPage - this.PAGINATION_RANGE) + (index - 1)
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
  a {
    cursor: pointer;
  }

  #more-sessions-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }

  .pagination-container {
    display: flex;
    flex-wrap: wrap;
  }
</style>
