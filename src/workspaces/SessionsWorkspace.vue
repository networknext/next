<template>
  <div>
    <v-tour name="sessionsTour" :steps="sessionsTourSteps" :options="sessionsTourOptions" :callbacks="sessionsTourCallbacks"></v-tour>
    <div
      class="spinner-border"
      role="status"
      id="sessions-spinner"
      v-show="!showTable"
    >
      <span class="sr-only">Loading...</span>
    </div>
    <div class="table-responsive table-no-top-line" v-show="showTable">
      <table class="table table-sm" :class="{'table-striped': sessions.length > 0, 'table-hover': sessions.length > 0}">
        <thead>
          <tr>
            <th>
              <span>
                <!-- No Title -->
              </span>
            </th>
            <th>
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Unique ID of the session">Session ID</span>
            </th>
            <th v-if="!$store.getters.isAnonymous">
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Hash of the unique user ID">User Hash</span>
            </th>
            <th>
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Internet service provider">
                  ISP
              </span>
            </th>
            <th v-if="$store.getters.isAdmin">
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Customer name">
                Customer
              </span>
            </th>
            <th>
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="The datacenter of the game server this session is connected to">Datacenter</span>
            </th>
            <th class="text-right">
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Round trip time of the session over the public internet">Direct RTT</span>
            </th>
            <th class="text-right">
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Round trip time of the session over Network Next">Next RTT</span>
            </th>
            <th class="text-right">
              <span
                data-toggle="tooltip"
                data-placement="right"
                title="Difference in round trip time between the public internet and Network Next (Direct - Next)">Improvement</span>
            </th>
          </tr>
        </thead>
        <tbody v-if="sessions.length === 0">
          <tr>
            <td colspan="7" class="text-muted">
                There are no top sessions at this time.
            </td>
          </tr>
        </tbody>
        <tbody>
          <tr v-for="(session, index) in sessions" v-bind:key="index">
            <td>
              <font-awesome-icon
                id="status"
                icon="circle"
                class="fa-w-16 fa-fw"
                v-bind:class="{
                  'text-success': session.on_network_next,
                  'text-primary': !session.on_network_next
                }"
              />
            </td>
            <td>
              <router-link
                v-bind:to="`/session-tool/${session.id}`"
                class="text-dark fixed-width"
                v-bind:data-intercom="index"
                v-bind:data-tour="index"
              >{{ session.id }}</router-link>
            </td>
            <td v-if="!$store.getters.isAnonymous">
              <router-link
                v-bind:to="`/user-tool/${session.user_hash}`"
                class="text-dark fixed-width"
              >{{ session.user_hash }}</router-link>
            </td>
            <td>{{ session.location.isp !== "" ? session.location.isp : "Unknown" }}</td>
            <td v-if="$store.getters.isAdmin">{{ getCustomerName(session.customer_id) }}</td>
            <td>
              <span
                class="text-dark"
              >{{ session.datacenter_alias !== "" ? session.datacenter_alias : session.datacenter_name }}</span>
            </td>
            <td
              class="text-right"
            >{{ parseFloat(session.direct_rtt).toFixed(2) == 0 ? "-" : parseFloat(session.direct_rtt).toFixed(2) }}</td>
            <td
              class="text-right"
            >{{ parseFloat(session.next_rtt).toFixed(2) == 0 ? "-" : parseFloat(session.next_rtt).toFixed(2) }}</td>
            <td class="text-right">
              <span
                v-if="session.delta_rtt > 0 && session.on_network_next"
                v-bind:class="{
                  'text-success': session.delta_rtt >= 5,
                  'text-warning': session.delta_rtt >= 2 && session.delta_rtt < 5,
                  'text-danger': session.delta_rtt < 2 && session.delta_rtt > 0
                }"
              >
                <b>{{ parseFloat(session.delta_rtt).toFixed(2) }}</b>
              </span>
              <span v-if="session.delta_rtt < 0 || !session.on_network_next">
                <b>-</b>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import SessionCounts from '@/components/SessionCounts.vue'
import { FeatureEnum } from '@/components/types/FeatureTypes'

/**
 * This component holds the workspace elements related to the top sessions page in the Portal
 */

/**
 * TODO: Cleanup template
 * TODO: Make this a View
 */

const MAX_RETRIES = 4

@Component({
  components: {
    SessionCounts
  }
})
export default class SessionsWorkspace extends Vue {
  private sessions: Array<any>
  private sessionsLoop: any
  private showTable: boolean
  private unwatch: any

  private sessionsTourSteps: Array<any>
  private sessionsTourOptions: any
  private sessionsTourCallbacks: any

  private retryCount: number

  constructor () {
    super()
    this.sessions = []
    this.showTable = false
    this.retryCount = 0

    this.sessionsTourSteps = [
      {
        target: '[data-tour="0"]',
        header: {
          title: 'Top Sessions'
        },
        content: 'Click on this <strong>Session ID</strong> to view more stats (such as latency, packet loss and jitter improvements).',
        params: {
          placement: 'bottom',
          enableScrolling: false
        }
      }
    ]

    this.sessionsTourOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'OK'
      }
    }

    this.sessionsTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_TOURS', 'sessions')

        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          Vue.prototype.$gtag.event('Sessions tour finished', {
            event_category: 'Tours'
          })
        }
      }
    }
  }

  private mounted () {
    this.restartLoop()
    this.unwatch = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        clearInterval(this.sessionsLoop)
        this.restartLoop()
      }
    )
    this.$root.$on('killLoops', this.stopLoop)
  }

  private beforeDestroy (): void {
    clearInterval(this.sessionsLoop)
    this.unwatch()
    this.$root.$off('killLoops')
  }

  private fetchSessions (): void {
    this.$apiService
      .fetchTopSessions({
        company_code: this.$store.getters.currentFilter.companyCode || ''
      })
      .then((response: any) => {
        this.retryCount = 0
        this.sessions = response.sessions || []
        if (this.$store.getters.isTour && this.$tours.sessionsTour && !this.$tours.sessionsTour.isRunning && !this.$store.getters.finishedTours.includes('sessions')) {
          this.$tours.sessionsTour.start()
        }
      })
      .catch((error: any) => {
        this.sessions = []
        console.log('Something went wrong fetching the top sessions list')
        console.log(error)

        this.stopLoop()
        this.retryCount = this.retryCount + 1
        if (this.retryCount < MAX_RETRIES) {
          setTimeout(() => {
            this.restartLoop()
          }, 3000 * this.retryCount)
        }

        if (this.retryCount >= MAX_RETRIES) {
          this.$root.$emit('killLoops')
        }
      })
      .finally(() => {
        if (!this.showTable) {
          this.showTable = true
        }
      })
  }

  // TODO: Move this somewhere with other helper functions
  private getCustomerName (buyerID: string): string {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].id === buyerID) {
        return allBuyers[i].company_name
      }
    }
    return 'Private'
  }

  private restartLoop () {
    this.stopLoop()
    this.fetchSessions()
    this.sessionsLoop = setInterval(() => {
      this.fetchSessions()
    }, 10000)
  }

  private stopLoop () {
    if (this.sessionsLoop) {
      clearInterval(this.sessionsLoop)
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .fixed-width {
    font-family: monospace;
    font-size: 120%;
  }
  div.table-no-top-line th {
    border-top: none !important;
  }
</style>
