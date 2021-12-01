<template>
  <div class="card-body" id="discoveryDashboards-page">
    <div v-if="urls.length > 0">
      <h5 class="card-title">
        Discovery Dashboards
      </h5>
      <p class="card-text">
        Interesting one off dashboards that are created by the Network Next datascience team
      </p>
      <div class="row" v-for="(url, index) in urls" :key="index">
        <div class="card" style="margin-bottom: 50px; width: 100%; margin: 0 1rem 2rem;">
          <div class="card-body">
            <LookerEmbed :dashURL="url" dashID="discoveryDashboard" />
          </div>
        </div>
      </div>
      <hr class="mt-4 mb-4">
    </div>
    <Saves />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

import LookerEmbed from '@/components/LookerEmbed.vue'
import Saves from '@/components/Saves.vue'

@Component({
  components: {
    LookerEmbed,
    Saves
  }
})
export default class Disccovery extends Vue {
  private urls: Array<string>
  private unwatchFilter: any

  constructor () {
    super()
    this.urls = []
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new billing URL
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchDiscoveryDashboards()
      }
    )
    this.fetchDiscoveryDashboards()
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchDiscoveryDashboards () {
    this.$apiService.fetchDiscoveryDashboards({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode,
      origin: window.location.origin
    })
      .then((response: any) => {
        this.urls = response.urls || []
        console.log(response)
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the discover dashboards')
        console.log(error)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
