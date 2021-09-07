<template>
  <div class="card-body" id="analytics-page">
    <h5 class="card-title">Analytics Dashboard</h5>
    <p class="card-text">
      One stop shop for analytics information
    </p>
    <iframe
      id="analyticsDash"
      :src="analyticsDashURL"
      v-if="showSummary"
      frameborder="0"
    >
    </iframe>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Analytics extends Vue {
  private analyticsDashURL: string
  private showSummary: boolean

  private unwatchFilter: any

  private startDate: string
  private endDate: string

  constructor () {
    super()
    this.analyticsDashURL = ''
    this.showSummary = false

    this.startDate = ''
    this.endDate = ''
  }

  private mounted () {
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchAnalyticsSummary()
      }
    )

    this.fetchAnalyticsSummary()
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchAnalyticsSummary () {
    this.$apiService.fetchAnalyticsSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.analyticsDashURL = response.url || ''
        if (this.analyticsDashURL !== '') {
          this.showSummary = true
        }
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics summary dashboard')
        console.log(error)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
