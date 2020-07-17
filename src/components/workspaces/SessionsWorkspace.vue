<template>
  <main role="main" class="col-md-12 col-lg-12 px-4">
    <div class="
            d-flex
            justify-content-between
            flex-wrap flex-md-nowrap
            align-items-center
            pt-3
            pb-2
            mb-3
            border-bottom"
    >
      <SessionCounts />
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1">
        <div class="mr-auto"></div>
        <div class="px-2 hidden">
          <select class="form-control">
            <option value="rtt">
              Sort by RTT
            </option>
          </select>
        </div>
        <div class="px-2 hidden">
          <select class="form-control">
            <option value="everything">
              Everything
            </option>
            <option value="direct">
              Direct Only
            </option>
            <option value="next">
              Network Next Only
            </option>
          </select>
        </div>
        <div class="px-2 hidden">
          <select class="form-control">
            <option value="all">
              All
            </option>
            <option>
              BUYERNAME ADMIN LOOP
            </option>
            <option>
              BUYERNAME SINGLE BUYER
            </option>
          </select>
        </div>
      </div>
    </div>
    <div class="spinner-border" role="status" id="sessions-spinner" v-show="!$store.getters.showTable">
      <span class="sr-only">Loading...</span>
    </div>
    <div class="table-responsive table-no-top-line" v-show="$store.getters.showTable">
      <table class="table table-sm table-striped table-hover">
        <thead>
          <tr>
            <th>
              <span>
                <!-- No Title -->
              </span>
            </th>
            <th>
              <span>
                Session ID
              </span>
            </th>
            <th v-if="!$store.getters.isAnonymous">
              <span>
                User Hash
              </span>
            </th>
            <th>
              <span>
                ISP
              </span>
            </th>
            <th>
              <span>
                Datacenter
              </span>
            </th>
            <th class="text-right">
              <span>
                Direct RTT
              </span>
            </th>
            <th class="text-right">
              <span>
                Next RTT
              </span>
            </th>
            <th class="text-right">
              <span>
                Improvement
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(session, index) in sessions" v-bind:key="index">
            <td>
              <font-awesome-icon icon="circle"
                                 class="fa-w-16 fa-fw"
                                 v-bind:class="{
                                  'text-success': session.onNetworkNext,
                                  'text-primary': !session.onNetworkNext
                                 }"
              />
            </td>
            <td>
              <a class="text-dark fixed-width" href="#">
                {{ session.id }}
              </a>
            </td>
            <td v-if="false">
              <a class="text-dark fixed-width" href="#">
                {{ session.user_hash }}
              </a>
            </td>
            <td>
              {{ session.location.isp != "" ? session.location.isp : "Unknown" }}
            </td>
            <td>
              <span class="text-dark">
                {{ session.datacenter_alias != "" ? session.datacenter_alias : session.datacenter_name }}
              </span>
            </td>
            <td class="text-right">
              {{ parseFloat(session.direct_rtt).toFixed(2) == 0 ? "-" : parseFloat(session.direct_rtt).toFixed(2) }}
            </td>
            <td class="text-right">
              {{ parseFloat(session.next_rtt).toFixed(2) == 0 ? "-" : parseFloat(session.next_rtt).toFixed(2) }}
            </td>
            <td class="text-right">
              <span v-if="session.delta_rtt > 0 && session.on_network_next"
                    v-bind:class="{
                        'text-success': session.delta_rtt >= 5,
                        'text-warning': session.delta_rtt >= 2 && session.delta_rtt < 5,
                        'text-danger': session.delta_rtt < 2 && session.delta_rtt > 0
                    }"
              >
                <b>
                    {{ parseFloat(session.delta_rtt).toFixed(2) }}
                </b>
              </span>
              <span v-if="session.delta_rtt < 0 || !session.on_network_next">
                <b>
                  -
                </b>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import SessionCounts from '@/components/SessionCounts.vue'
import { SessionMeta } from '@/components/types/APITypes'
import APIService from '../../services/api.service'

/**
 * TODO: Move the filter dropdown bar to its own component
 * TODO: Hookup API call
 * TODO: Hookup lifecycle hooks and a spinner if necessary
 * TODO: Hookup looping logic
 * TODO: Cleanup template
 * TODO: Add in Relay interface
 * TODO: Figure out what sessionMeta fields need to be required
 */

@Component({
  components: {
    SessionCounts
  }
})
export default class SessionsWorkspace extends Vue {
  private sessions: Array<SessionMeta>
  private apiService: APIService
  private showTable = false
  private sessionsLoop = -1

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
    this.sessions = []
  }

  private mounted () {
    this.fetchSessions()
    this.sessionsLoop = setInterval(() => {
      this.fetchSessions()
    }, 10000)
  }

  private beforeDestroy () {
    // Stop polling loop
    this.sessions = []
    this.$store.commit('TOGGLE_SESSION_TABLE', false)
    clearInterval(this.sessionsLoop)
  }

  private fetchSessions () {
    this.apiService.call('BuyersService.TopSessions', {})
      .then((response: any) => {
        this.sessions = response.result.sessions
        this.$store.commit('TOGGLE_SESSION_TABLE', true)
      })
      .catch((error: any) => {
        console.log(error)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
