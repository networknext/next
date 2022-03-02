<template>
  <div>
    <div>
      <div class="card-title" style="text-align: center; padding-top: 1rem;">
        Coming Soon!
      </div>
    </div>
    <div v-if="false" style="padding-top: 1rem;"> <!-- TODO: Bring this back eventually -->
      <div
        class="spinner-border"
        role="status"
        id="saves-spinner"
        v-show="!showSaves"
      >
        <span class="sr-only">Loading...</span>
      </div>
      <div v-if="showSaves">
        <h5 class="card-title looker-padding">
          Recent Saves
        </h5>
        <p class="card-text looker-padding">
          Saves that have happened in the last week
        </p>
        <div class="table-responsive table-no-top-line looker-padding">
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
                  <span>Save Score</span>
                </th>
                <th>
                  <span>Average Direct RTT</span>
                </th>
                <th>
                  <span>Average Next RTT</span>
                </th>
                <th>
                  <span>Average Direct Packet Loss</span>
                </th>
                <th>
                  <span>Average Next Packet Loss</span>
                </th>
                <th>
                  <span>Duration (Hours)</span>
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
                  {{ save.save_score }}
                </td>
                <td>
                  {{ save.average_direct_rtt }}
                </td>
                <td>
                  {{ save.average_next_rtt }}
                </td>
                <td>
                  {{ save.average_direct_pl }}
                </td>
                <td>
                  {{ save.average_next_pl }}
                </td>
                <td>
                  {{ save.duration }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="row" v-if="savesDashURL !== ''">
          <LookerEmbed dashID="savesDash" :dashURL="savesDashURL" />
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

import LookerEmbed from '@/components/LookerEmbed.vue'

/**
 * This component displays all of the necessary information for the user management tab
 *  within the settings page of the Portal and houses all the associated logic and api calls
 */

/**
 * TODO: Clean up template
 * TODO: Pretty sure the card-body can be taken out into a wrapper component - same with route shader and game config...
 */

@Component({
  components: {
    LookerEmbed
  }
})
export default class Saves extends Vue {
  private saves: Array<any>
  private savesDashURL: string
  private showSaves: boolean

  private unwatchFilter: any

  constructor () {
    super()
    this.saves = []
    this.savesDashURL = ''
    this.showSaves = false
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new billing URL
    if (this.$store.getters.isAdmin) {
      this.unwatchFilter = this.$store.watch(
        (state: any, getters: any) => {
          return getters.currentFilter
        },
        () => {
          this.fetchCurrentSavesData()
        }
      )

      this.fetchCurrentSavesData()
    }
  }

  private beforeDestroy () {
    if (this.$store.getters.isAdmin) {
      this.unwatchFilter()
    }
  }

  private fetchCurrentSavesData () {
    const customerCode = this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    const promises = [
      this.$apiService.fetchSavesDashboard({
        customer_code: customerCode,
        origin: window.location.origin
      }),
      this.$apiService.fetchCurrentSaves({
        customer_code: customerCode
      })
    ]

    Promise.all(promises)
      .then((responses: any) => {
        this.savesDashURL = responses[0].url || ''
        this.saves = responses[1].saves || []
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching saves for that date range')
        console.log(error)
      })
      .finally(() => {
        this.showSaves = true
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  .looker-padding {
    padding-right: 3.75rem;
    padding-left: 3.75rem;
  }
</style>
