<template>
  <div class="card-body">
    <h3 v-if="tabs.length === 0">No dashboards available</h3>
    <div class="card" style="margin-bottom: 250px;" v-if="tabs.length > 0">
      <div class="card-body">
        <div class="row">
          <ul>
            <li v-for="(tab, tabIndex) in tabs" :key="tabIndex" @click="selectTab(tabIndex)">
              <a :class="{ active: tabIndex === selectedTabIndex }">{{ tab }}</a>
              <div class="blue-accent"></div>
            </li>
          </ul>
        </div>
        <div class="row" v-for="(url, urlIndex) in urls" :key="urlIndex">
          <iframe
            class="col"
            :id="`discoveryDashboard-${index}`"
            :src="url"
            style="min-height: 2500px;"
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
export default class Analytics extends Vue {
  private dashboards: any
  private selectedTabIndex: number
  private tabs: Array<string>
  private urls: Array<string>

  private unwatchFilter: any

  constructor () {
    super()
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
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchAnalyticsDashboards () {
    this.$apiService.fetchAnalyticsDashboards({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        console.log(response)
        this.dashboards = response.dashboards || []
        this.tabs = Object.keys(this.dashboards)
        this.selectTab(0)
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics dashboard categories')
        console.log(error)
      })
  }

  private selectTab (index: number) {
    this.selectedTabIndex = index
    this.urls = this.dashboards[this.tabs[index]]
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .blue-accent {
    border-bottom: solid #009FDF;
    width: 2.2rem;
    padding-bottom: 2px;
  }
</style>
