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
    <div class="spinner-border" role="status" id="sessions-spinner" v-if="false">
      <span class="sr-only">Loading...</span>
    </div>
    <div class="table-responsive table-no-top-line" v-if="true">
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
            <th v-if="false">
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
                {{ session.userHash }}
              </a>
            </td>
            <td>
              {{ session.location.ISP }}
            </td>
            <td>
              <span class="text-dark">
                {{ session.datacenterAlias != '' ? session.datacenterAlias : session.datacenterName }}
              </span>
            </td>
            <td class="text-right ">
              {{ session.directRTT }}
            </td>
            <td class="text-right ">
              {{ session.nextRTT }}
            </td>
            <td class="text-right ">
                <!-- TODO: This should probably be a if/else -->
              <span v-if="session.deltaRTT > 0 && session.onNetworkNext"
                    v-bind:class="{
                      'text-success': session.deltaRTT >= 5,
                      'text-warning': session.deltaRTT >= 2 && session.deltaRTT < 5,
                      'text-danger': session.deltaRTT < 2 && session.deltaRTT > 0
                    }"
              >
              </span>
              <span v-if="session.deltaRTT < 0 || !session.onNetworkNext">
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

  constructor () {
    super()
    this.sessions = []
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
