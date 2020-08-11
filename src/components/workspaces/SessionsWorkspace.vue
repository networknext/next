<template>
  <div>
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
            <th v-if="$store.getters.isAdmin">
                Customer
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
                                  'text-success': session.on_network_next,
                                  'text-primary': !session.on_network_next
                                 }"
              />
            </td>
            <td>
              <router-link v-bind:to="`/session-tool/${session.id}`" class="text-dark fixed-width">{{ session.id }}</router-link>
            </td>
            <td v-if="!$store.getters.isAnonymous">
              <router-link v-bind:to="`/user-tool/${session.user_hash}`" class="text-dark fixed-width">{{ session.user_hash }}</router-link>
            </td>
            <td>
              {{ session.location.isp != "" ? session.location.isp : "Unknown" }}
            </td>
            <td v-if="$store.getters.isAdmin">
                {{ getCustomerName(session.customer_id) }}
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
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import SessionCounts from '@/components/SessionCounts.vue'
import { SessionMeta } from '@/components/types/APITypes'
import APIService from '../../services/api.service'
import { Route, NavigationGuardNext } from 'vue-router'

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
  private sessions: Array<any>
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
    this.apiService.fetchTopSessions({ buyer_id: '' })
      .then((response: any) => {
        this.sessions = response.sessions
        this.$store.commit('TOGGLE_SESSION_TABLE', true)
      })
      .catch((error: any) => {
        console.log(error)
      })
  }

  // TODO: Move this somewhere with other helper functions
  private getCustomerName (buyerId: string) {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].id === buyerId) {
        return allBuyers[i].name
      }
    }
    return 'Private'
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    console.log('Before Route Update')
    console.log(to)
    console.log(from)
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
