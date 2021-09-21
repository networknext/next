<template>
  <div class="card-body" id="analytics-page">
    <div class="row">
      <iframe
        class="col"
        id="analyticsDash"
        :src="analyticsDashURL"
        style="min-height: 1800px;"
        v-if="analyticsDashURL !== ''"
        frameborder="0"
      >
      </iframe>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Analytics extends Vue {
  private analyticsDashURL: string

  private unwatchFilter: any

  constructor () {
    super()
    this.analyticsDashURL = ''
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
