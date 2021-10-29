<template>
  <div class="card-body" id="analytics-page">
    <div v-for="(url, index) in analyticsDashURLs" :key="index" class="row">
      <div class="card-body">
        <iframe
          class="col"
          id="analyticsDash"
          :src="url"
          style="min-height: 1800px;"
          v-if="url !== ''"
          frameborder="0"
        >
        </iframe>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Analytics extends Vue {
  private analyticsDashURLs: Array<string>

  private unwatchFilter: any

  constructor () {
    super()
    this.analyticsDashURLs = []
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new analytics URL
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchAnalyticsSummary()
      }
    )

    this.fetchAnalyticsSummary()

    const usageDashElement = document.getElementById('usageDash')
    if (usageDashElement) {
      usageDashElement.addEventListener('dashboard:run:complete', this.iframeTimeoutHandler)
    }
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchAnalyticsSummary () {
    this.$apiService.fetchAnalyticsSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.analyticsDashURLs = response.urls || []
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics summary dashboard')
        console.log(error)
      })
  }

  private iframeTimeoutHandler () {
    // TODO: Look for a status of error or stopped and display a refresh page message....
    console.log('An iframe timed out we should add an alert to refresh the page!')
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
