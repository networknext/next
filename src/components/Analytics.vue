<template>
  <div class="card" style="margin-bottom: 250px;" v-if="tabs.length > 0">
    <div class="card-header">
      <ul class="nav nav-tabs card-header-tabs">
        <li class="nav-item" v-for="(tab, tabIndex) in tabs" :key="tabIndex" @click="selectTab(tabIndex, false)">
          <a class="nav-link" :class="{ active: selectedTabIndex === tabIndex }">{{ tab }}</a>
        </li>
      </ul>
    </div>
    <div class="card-body" v-if="!subTabs[tabs[selectedTabIndex]]">
      <div v-for="(dashboard, dashIndex) in tabDashboards" :key="dashIndex">
        <div class="row">
          <LookerEmbed :dashURL="dashboard" dashID="previewDash" />
        </div>
      </div>
    </div>
    <div class="card-body" v-if="subTabs[tabs[selectedTabIndex]]">
      <div class="row border-bottom mb-3">
        <ul class="sub-ul">
          <li class="sub-li" :class="{ 'blue-accent': subTabIndex === selectedSubTabIndex }" v-for="(subTab, subTabIndex) in subTabs[tabs[selectedTabIndex]]" :key="subTabIndex" @click="selectTab(subTabIndex, true)">
            <a class="sub-link">{{ subTab.label }}</a>
          </li>
        </ul>
      </div>
      <div v-for="(dashboard, dashIndex) in tabDashboards" :key="dashIndex">
        <div class="row">
          <LookerEmbed :dashURL="dashboard" dashID="previewDash" />
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
  private selectedSubTabIndex: number
  private selectedTabIndex: number
  private tabs: Array<any>
  private tabDashboards: Array<any>
  private subTabs: any

  constructor () {
    super()
    this.dashboards = {}
    this.selectedSubTabIndex = -1
    this.selectedTabIndex = -1
    this.tabs = []
    this.tabDashboards = []
    this.subTabs = {}
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
        this.tabs = response.labels || []
        this.subTabs = response.sub_categories || {}

        this.selectedTabIndex = 0
        this.selectedSubTabIndex = 0

        const currentTab = this.tabs[this.selectedTabIndex]
        const subTabs = this.subTabs[currentTab]

        if (subTabs) {
          const firstSubTab = subTabs[this.selectedSubTabIndex].label
          this.tabDashboards = this.dashboards[firstSubTab]
        } else {
          this.tabDashboards = this.dashboards[currentTab]
        }
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics dashboard categories')
        console.log(error)
        this.$refs.failureAlert.setMessage('Failed to fetch analytics dashboards. Please refresh the page')
        this.$refs.failureAlert.setAlertType(AlertType.ERROR)
      })
  }

  private selectTab (index: number, isSubTab: boolean) {
    if (index === this.selectedTabIndex) {
      return
    }

    if (isSubTab) {
      this.selectedSubTabIndex = index
    } else {
      this.selectedTabIndex = index
      this.selectedSubTabIndex = 0
    }

    // TODO: This is a bit wasteful because we are making multiple URLs when we only need the one specific to the selected tab. Make a refresh endpoint that will just reload the tab
    this.$apiService.fetchAnalyticsDashboards({
      customer_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.dashboards = response.dashboards || {}
        this.tabs = response.labels || []
        this.subTabs = response.sub_categories || {}

        // If a tab is deleted in the admin tool before this tab switch, the selectedIndex could be greater than the number of new tabs
        if (this.selectedTabIndex > this.tabs.length) {
          this.selectedTabIndex = 0
        }

        const currentTab = this.tabs[this.selectedTabIndex]
        const subTabs = this.subTabs[currentTab]

        if (subTabs) {
          const firstSubTab = subTabs[this.selectedSubTabIndex].label
          this.tabDashboards = this.dashboards[firstSubTab]
        } else {
          this.tabDashboards = this.dashboards[currentTab]
        }
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
  .nav-link {
    cursor: pointer;
  }

  .sub-ul {
    list-style-type: none;
    margin-left: 1rem;
    margin-bottom: 0;
    margin-right: 0;
    margin-top: 0;
    padding: 0;
    overflow: hidden;
  }

  .sub-li {
    float: left;
  }

  .sub-li:hover {
    border-top: 1px solid transparent;
    border-left: 1px solid transparent;
    border-right: 1px solid transparent;
    border-top-left-radius: .25rem;
    border-top-right-radius: .25rem;
    border-color: #e9ecef #e9ecef #dee2e6;
  }

  .sub-link {
    display: block;
    text-align: center;
    padding: 16px;
    padding-top: 4px;
    padding-bottom: 4px;
    text-decoration: none;
    cursor: pointer;
    color: black;
  }

  .blue-accent {
    border-bottom: solid #009FDF;
    margin-bottom: 10px;
  }
</style>
