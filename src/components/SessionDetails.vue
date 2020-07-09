<template>
  <div v-if="showDetails">
    <div class="row">
      <div class="col-12 col-lg-8">
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Latency
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: rgb(0, 109, 44);">
                — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                — Direct
              </span>
              <span></span>
            </div>
          </div>
          <div class="card-body">
            <div id="latency-chart-1"></div>
          </div>
        </div>
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Jitter
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: rgb(0, 109, 44);">
                — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                — Direct
              </span>
              <span></span>
            </div>
          </div>
          <div class="card-body">
            <div id="jitter-chart-1"></div>
          </div>
        </div>
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Packet Loss
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: rgb(0, 109, 44);">
                  — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                  — Direct
              </span>
              <span></span>
            </div>
          </div>
          <div class="card-body">
            <div id="packet-loss-chart-1"></div>
          </div>
        </div>
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Bandwidth
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: orange;">
                — Up
              </span>
              <span style="color: blue;">
                — Down
              </span>
              <span></span>
            </div>
            <div class="float-right">
              <span id="traffic-up-legend"></span>
              <span id="traffic-down-legend"></span>
            </div>
          </div>
          <div class="card-body">
            <div id="bandwidth-chart-1"></div>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="card">
          <div class="card-img-top">
            <div id="session-tool-map"
                  style="
                      width: 100%;
                      height: 40vh;
                      margin: 0px;
                      padding: 0px;
                      position: relative;
                  "
            ></div>
          </div>
          <div class="card-body">
            <div class="card-text">
              <dl>
                <dt>
                  ISP
                </dt>
                <dd>
                  <em>
                    META LOCATION ISP
                  </em>
                </dd>
                <div v-if="false">
                  <dt>
                    User Hash
                  </dt>
                  <dd>
                    <a class="text-dark">
                      META USER HASH
                    </a>
                  </dd>
                </div>
                <dt>
                    User IP Address
                </dt>
                <dd>
                    META CLIENT ADDR
                </dd>
                <dt>
                    Platform
                </dt>
                <dd>
                    META PLATFORM
                </dd>
                <dt v-if="false">
                    Customer
                </dt>
                <dd v-if="false">
                    BUYER NAME
                </dd>
                <dt>
                  SDK Version
                </dt>
                <dd>
                  META SDK
                </dd>
                <dt>
                  Connection Type
                </dt>
                <dd>
                  META CONNECTION
                </dd>
                <dt v-if="false">
                    Nearby Relays
                </dt>
                <dd v-if="false">
                    No Nearby Relays
                </dd>
                <table class="table table-sm mt-1" v-if="false">
                  <thead>
                    <tr>
                      <th style="width: 50%;">
                        Name
                      </th>
                      <th style="width: 16.66%;">
                        RTT
                      </th>
                      <th style="width: 16.66%;">
                        Jitter
                      </th>
                      <th style="width: 16.66%;">
                        Packet Loss
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                      <!-- <tr v-for="relay in pages.sessionTool.meta.nearby_relays"> -->
                      <tr>
                        <td>
                          <a class="text-dark">NAME</a>&nbsp;
                        </td>
                        <td>
                          CLIENT STATS RTT
                        </td>
                        <td>
                          CLIENT STATS JITTER
                        </td>
                        <td>
                          CLIENT STATS PACKETLOSS%
                        </td>
                      </tr>
                  </tbody>
                </table>
                <dt  v-if="false">
                    Route
                </dt>
                <table class="table table-sm mt-1" v-if="false">
                  <thead>
                    <tr>
                      <th style="width: 50%;">
                        Name
                      </th>
                      <th style="width: 50%;">
                        Role
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                        <td>
                            META CLIENT ADDR
                        </td>
                        <td>
                            <em>
                                User (Player)
                            </em>
                        </td>
                    </tr>
<!-- <tr v-for="(hop, index) in pages.sessionTool.meta.hops" scope="row"> -->
                    <tr>
                      <td>
                          NAME
                      </td>
                      <td>
                          Hop INDEX + 1
                      </td>
                    </tr>
                    <tr>
                      <td>
                        META SERVER ADDR
                      </td>
                      <td>
                          <em>Destination Server</em>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * TODO: Cleanup template
 * TODO: Figure out what sessionMeta fields need to be required
 * TODO: Hookup API call
 */

@Component
export default class SessionDetails extends Vue {
  // TODO: Refactor out the alert/error into its own component.
  private showDetails = false

  private created () {
    console.log('Creating session details')
    this.fetchSessionDetails()
    /* setInterval(() => {
      this.fetchSessionDetails()
    }, 10000) */
  }

  private fetchSessionDetails () {
    // API Call to fetch the details associated to ID
    this.showDetails = true
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
