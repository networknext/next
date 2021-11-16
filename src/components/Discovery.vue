<template>
  <div class="card-body" id="discoveryDashboards-page">
    <h5 class="card-title">
      Discovery Dashboards
    </h5>
    <p class="card-text">
      Interesting one off dashboards that are created by the Network Next datascience team
    </p>
    <div class="row" v-for="(url, index) in urls" :key="index">
      <div class="card" style="margin-bottom: 50px; width: 100%; margin: 0 1rem 2rem;">
        <div class="card-body">
          <iframe
            class="col"
            id="discoveryDashboard"
            :src="url"
            style="min-height: 1000px;"
            v-show="url !== ''"
            frameborder="0"
          >
          </iframe>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
@Component
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
    window.addEventListener('message', this.resizeIframes)
  }

  private beforeDestroy () {
    this.unwatchFilter()
    window.removeEventListener('message', this.resizeIframes)
  }

  private resizeIframes (event: any) {
    const iframes = document.querySelectorAll('#discoveryDashboard')
    iframes.forEach((frame: any) => {
      if (event.source === frame.contentWindow && event.origin === 'https://networknextexternal.cloud.looker.com' && event.data) {
        const eventData = JSON.parse(event.data)
        if (eventData.type === 'page:properties:changed') {
          frame.height = eventData.height + 50
        }
      }
    })
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
