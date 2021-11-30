<template>
  <div class="card-body">
    <div>
      <h5 class="card-title">
        Recent Saves
      </h5>
      <p class="card-text">
        Saves that happened recently
      </p>
      <div class="table-responsive table-no-top-line">
        <table class="table table-sm" :class="{'table-striped': saves.length > 0, 'table-hover': saves.length > 0}">
          <thead>
            <tr>
              <th>
                <span
                  data-toggle="tooltip"
                  data-placement="right"
                  title="Unique ID of the session">Session ID</span>
              </th>
              <th>
                <span>RTT</span>
              </th>
              <th>
                <span>Jitter</span>
              </th>
              <th>
                <span>PL</span>
              </th>
              <th>
                <span>Score</span>
              </th>
            </tr>
          </thead>
          <tbody v-if="saves.length === 0">
            <tr>
              <td colspan="7" class="text-muted">
                  There are no saves at this time.
              </td>
            </tr>
          </tbody>
          <tbody>
            <tr v-for="(save, index) in saves" :key="index">
              <td>
                <router-link
                  :to="`/session-tool/${save.id}`"
                  class="text-dark fixed-width"
                  :data-intercom="index"
                  :data-tour="index"
                >{{ save.id }}</router-link>
              </td>
              <td>
                0
              </td>
              <td>
                0
              </td>
              <td>
                0
              </td>
              <td>
                0
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <hr class="mt-4 mb-4">
    </div>
    <div>
      <h5 class="card-title">
        Historic Saves
      </h5>
      <p class="card-text">
        Saves over a given span of time
      </p>
      <hr class="mt-4 mb-4">
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { DateTime } from 'luxon'

/**
 * This component displays all of the necessary information for the user management tab
 *  within the settings page of the Portal and houses all the associated logic and api calls
 */

/**
 * TODO: Clean up template
 * TODO: Pretty sure the card-body can be taken out into a wrapper component - same with route shader and game config...
 */

@Component
export default class Saves extends Vue {
  private dateString: string
  private saves: Array<any>
  private savesDashURL: string

  private unwatchFilter: any

  constructor () {
    super()
    this.dateString = ''
    this.saves = []
    this.savesDashURL = ''
  }

  private mounted () {
    const now = DateTime.now()
    const currentDateString = `${now.year}-${now.month}`

    // Check URL date and set to default if empty
    this.dateString = this.$route.params.pathMatch || ''

    if (this.dateString !== currentDateString) {
      const passedInDate = this.dateString.split('-')
      // check for invalid date
      if (parseInt(passedInDate[0]) > now.year || (parseInt(passedInDate[0]) === now.year && parseInt(passedInDate[1]) > now.month)) {
        this.dateString = ''
      }
    }

    // This is only necessary for admins - when the filter changes, grab the new billing URL
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchCurrentSaves()
      }
    )

    this.fetchCurrentSaves()

    window.addEventListener('message', this.resizeIframes)
  }

  private beforeDestroy () {
    this.unwatchFilter()
    window.removeEventListener('message', this.resizeIframes)
  }

  private resizeIframes (event: any) {
    const iframe = document.getElementById('usageDash') as HTMLIFrameElement
    if (iframe && event.source === iframe.contentWindow && event.origin === 'https://networknextexternal.cloud.looker.com' && event.data) {
      const eventData = JSON.parse(event.data)
      if (eventData.type === 'page:properties:changed') {
        iframe.height = eventData.height + 50
      }
    }
  }

  private fetchCurrentSaves () {
    this.$apiService.fetchCurrentSaves({
      customer_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        console.log(response)
        this.saves = response.saves || []
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching saves for that date range')
        console.log(error)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
