<template>
  <div class="card-body">
    <h5 class="card-title">
      Analytic Dashboards (Beta)
    </h5>
    <p class="card-text">
      Currated dashboards that provide a look into your data. This feature is in beta and will be receiving continuous updates.
    </p>
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
          <Alert ref="failureAlert"/>
          <LookerEmbed :dashURL="url" dashID="analyticsDash" />
        </div>
      </div>
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

  private unwatchFilter: any
  private dashboards: any
  private domain: string
  private selectedTabIndex: number
  private tabs: Array<string>
  private urls: Array<string>

  constructor () {
    super()
    this.domain = window.location.origin
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
    if (this.$store.getters.isAdmin) {
      this.unwatchFilter()
    }
  }

  private fetchAnalyticsDashboards () {
    const customerCode = this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    this.$apiService.fetchAnalyticsDashboards({
      customer_code: customerCode
    })
      .then((response: any) => {
        this.dashboards = response.dashboards || []
        if (this.dashboards.length === 0) {
          return
        }
        this.tabs = Object.keys(this.dashboards)
        if (customerCode === 'madbyte-games') {
          this.tabs = [
            'Summary',
            'Acceleration Results',
            'AB Test Results',
            'Country Analysis'
          ]
        } else {
          this.tabs.sort((a: any, b: any) => {
            return a < b ? -1 : 1
          })
        }

        this.selectedTabIndex = 0
        this.urls = this.dashboards[this.tabs[0]]
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics dashboard categories')
        console.log(error)
        this.$refs.failureAlert.setMessage('Failed to fetch analytics dashboards. Please refresh the page')
        this.$refs.failureAlert.setAlertType(AlertType.ERROR)
      })
  }

  private selectTab (index: number) {
    if (index === this.selectedTabIndex) {
      return
    }
    this.selectedTabIndex = index

    // TODO: This is a bit wasteful because we are making multiple URLs when we only need the one specific to the selected tab. Make a refresh endpoint that will just reload the tab
    this.$apiService.fetchAnalyticsDashboards({
      customer_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.dashboards = response.dashboards || []
        this.urls = this.dashboards[this.tabs[index]]
      })
      .catch((error: Error) => {
        console.log('There was an issue refreshing the analytics dashboards')
        console.log(error)
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
  ul {
    list-style-type: none;
    margin-left: 1rem;
    margin-bottom: 0;
    margin-right: 0;
    margin-top: 0;
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
