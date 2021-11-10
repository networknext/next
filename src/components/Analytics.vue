<template>
  <div class="card-body">
    <h3 v-if="tabs.length === 0">No dashboards available</h3>
    <div class="card" v-if="tabs.length > 0">
      <div class="card-body">
        <div class="row border-bottom mb-3">
          <ul>
            <li :class="{ 'blue-accent': tabIndex === selectedTabIndex }" v-for="(tab, tabIndex) in tabs" :key="tabIndex" @click="selectTab(tabIndex)">
              <a>{{ tab }}</a>
            </li>
          </ul>
        </div>
        <div class="row" v-for="(url, urlIndex) in urls" :key="urlIndex">
          <iframe
            class="col"
            id="analyticsDashboard"
            :src="url"
            v-if="url !== ''"
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
export default class Analytics extends Vue {
  private dashboards: any
  private domain: string
  private heights: Array<number>
  private selectedTabIndex: number
  private tabs: Array<string>
  private urls: Array<string>

  private unwatchFilter: any

  constructor () {
    super()
    this.domain = window.location.origin
    this.heights = []
    this.selectedTabIndex = -1
    this.tabs = []
    this.urls = []
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new analytics URL
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchAnalyticsDashboards()
      }
    )

    this.fetchAnalyticsDashboards()

    window.addEventListener('message', this.resizeIframes)
  }

  private beforeDestroy () {
    this.unwatchFilter()
    window.removeEventListener('message', this.resizeIframes)
  }

  private resizeIframes (event: any) {
    const iframes = document.querySelectorAll('#analyticsDashboard')
    iframes.forEach((frame: any, index: number) => {
      if (event.source === frame.contentWindow && event.origin === 'https://networknextexternal.cloud.looker.com' && event.data) {
        const eventData = JSON.parse(event.data)
        if (eventData.type === 'page:properties:changed') {
          frame.height = eventData.height + 50
        }
      }
    })
  }

  private fetchAnalyticsDashboards () {
    this.$apiService.fetchAnalyticsDashboards({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.dashboards = response.dashboards || []
        this.tabs = Object.keys(this.dashboards)
        this.selectedTabIndex = 0
        this.urls = this.dashboards[this.tabs[0]]
        this.heights = new Array(this.urls.length)
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics dashboard categories')
        console.log(error)
      })
  }

  private selectTab (index: number) {
    if (index === this.selectedTabIndex) {
      return
    }
    this.selectedTabIndex = index
    this.urls = this.dashboards[this.tabs[index]]
    this.heights = new Array(this.urls.length)

    // TODO: This is a bit wasteful because we are making multiple URLs when we only need the one specific to the selected tab. Make a refresh endpoint that will just reload the tab
    this.$apiService.fetchAnalyticsDashboards({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.dashboards = response.dashboards || []
      })
      .catch((error: Error) => {
        console.log('There was an issue refreshing the analytics dashboards')
        console.log(error)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  ul {
    list-style-type: none;
    margin: 0;
    padding: 0;
    overflow: hidden;
  }

  li {
    float: left;
  }

  li:hover {
    border-top: 1px solid transparent;
    border-left: 1px solid transparent;
    border-right: 1px solid transparent;
    border-top-left-radius: .25rem;
    border-top-right-radius: .25rem;
    border-color: #e9ecef #e9ecef #dee2e6;
  }

  li a {
    display: block;
    text-align: center;
    padding: 16px;
    padding-top: 4px;
    padding-bottom: 4px;
    text-decoration: none;
    cursor: pointer;
  }

  .blue-accent {
    border-bottom: solid #009FDF;
    margin-bottom: 10px;
  }
</style>
