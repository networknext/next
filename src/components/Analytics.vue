<template>
  <div class="card-body" id="analytics-page">
    <Alert ref="failureAlert"/>
    <div v-for="(url, index) in analyticsDashURLs" :key="index" class="row">
      <LookerEmbed :dashURL="url" dashID="analyticsDash" />
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

import Alert from '@/components/Alert.vue'
import LookerEmbed from '@/components/LookerEmbed.vue'
import { AlertType } from './types/AlertTypes'

@Component({
  components: {
    Alert,
    LookerEmbed
  }
})
export default class Analytics extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    failureAlert: Alert;
  }

  private analyticsDashURLs: Array<string>

  private unwatchFilter: any

  constructor () {
    super()
    this.analyticsDashURLs = []
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new analytics URL
    if (this.$store.getters.isAdmin) {
      this.unwatchFilter = this.$store.watch(
        (state: any, getters: any) => {
          return getters.currentFilter
        },
        () => {
          this.fetchAnalyticsSummary()
        }
      )
    }

    this.fetchAnalyticsSummary()
  }

  private beforeDestroy () {
    if (this.$store.getters.isAdmin) {
      this.unwatchFilter()
    }
  }

  private fetchAnalyticsSummary () {
    this.$apiService.fetchAnalyticsSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode,
      origin: window.location.origin
    })
      .then((response: any) => {
        this.analyticsDashURLs = response.urls || []
      })
      .catch((error: Error) => {
        console.log('Something went wrong fetching analytics dashboards')
        console.log(error)
        this.$refs.failureAlert.setMessage('Failed to fetch analytics dashboards. Please refresh the page')
        this.$refs.failureAlert.setAlertType(AlertType.ERROR)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
