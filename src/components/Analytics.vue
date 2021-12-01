<template>
  <div class="card-body" id="analytics-page">
    <div v-for="(url, index) in analyticsDashURLs" :key="index" class="row">
      <LookerEmbed :dashURL="url" dashID="analyticsDash" />
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

import LookerEmbed from '@/components/LookerEmbed.vue'

@Component({
  components: {
    LookerEmbed
  }
})
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
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchAnalyticsSummary () {
    this.$apiService.fetchAnalyticsSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode,
      origin: window.location.origin
    })
      .then((response: any) => {
        this.analyticsDashURLs = response.urls || []
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
