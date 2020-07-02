<template>
  <main role="main" class="col-md-12 col-lg-12 px-4">
    <div id="sessions">
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
        <h1 class="h2" id="workspace-title" v-if="true">
          Sessions&nbsp;
          <span class="badge badge-dark">
            MAPSESSIONS Total Sessions
          </span>&nbsp;
          <span class="badge badge-success">
            ONNN on Network Next
          </span>
        </h1>
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
                    <!-- <tr v-for="(session, index) in pages.sessions.sessions"> -->
                      <tr>
                        <td>
                            <font-awesome-icon icon="circle"
                                                class="fa-w-16 fa-fw text-success"
                                                v-if="true"
                            />
                            <font-awesome-icon icon="circle"
                                                class="fa-w-16 fa-fw text-primary"
                                                v-if="false"
                            />
                        </td>
                        <td>
                            <a class="text-dark fixed-width" href="#">
                                SESSION ID
                            </a>
                        </td>
                        <td v-if="false">
                            <a class="text-dark fixed-width" href="#">
                                SESSION USER HASH
                            </a>
                        </td>
                        <td>
                            SESSION LOCATION ISP
                        </td>
                        <td>
                            <span class="text-dark">
                                SESSION DATACENTER
                            </span>
                        </td>
                        <td class="text-right ">
                            SESSION DIRECT RTT
                        </td>
                        <td class="text-right ">
                            SESSION NEXT RTT
                        </td>
                        <td class="text-right ">
  <!-- <span v-if="session.delta_rtt > 0 && session.on_network_next"
  v-bind:class="{
      'text-success': session.delta_rtt >= 5,
      'text-warning': session.delta_rtt >= 2 && session.delta_rtt < 5,
      'text-danger': session.delta_rtt < 2 && session.delta_rtt > 0
  }"
  > -->
                            <span>
                                <b>
                                    SESSION DELTA RTT
                                </b>
                            </span>
  <!-- <span v-if="session.delta_rtt < 0 || !session.on_network_next">
  <b>
  -
  </b>
  </span> -->
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

@Component
export default class SessionsWorkspace extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
