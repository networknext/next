<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Session</p>
      <p class="tight-p test-text"><input id='session-id-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-show="this.found" class="bottom">

      <div id="left" class="left">

        <div class="d-xxl-none session-info-mobile">
          <table id="session_table" class="table table-striped" style="vertical-align: middle; padding: 15px;">
            <tbody>

              <tr>
                <td class="bold">Datacenter</td>
                <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Server</td>
                <td> <router-link :to="'/server/' + this.data['server_address']"> {{ this.data['server_address'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">ISP</td>
                <td> {{ this.data['isp'] }} </td>
              </tr>

              <tr>
                <td class="bold">Platform</td>
                <td> {{ this.data['platform'] }} </td>
              </tr>

              <tr>
                <td class="bold">Connection</td>
                <td> {{ this.data['connection'] }} </td>
              </tr>

              <tr>
                <td class="bold">User Hash</td>
                <td class="fixed"> <router-link :to="'/user/' + this.data['user_hash']"> {{ this.data['user_hash'] }} </router-link> </td>
              </tr>

            </tbody>
          </table>
        </div>

        <div id="latency_graph" class="graph"/>
        
        <div id="jitter_graph" class="graph"/>
        
        <div id="packet_loss_graph" class="graph"/>

        <div id="out_of_order_graph" class="graph"/>

        <div id="bandwidth_up_graph" class="graph"/>

        <div id="bandwidth_down_graph" class="graph"/>

        <div class="d-xxl-none">

          <p class="header" style="padding-top: 15px; padding-bottom: 5px">Route</p>
   
          <table id="route_table" class="table" v-if="this.data['route_relays'] != null && this.data['route_relays'].length > 0">

            <tbody>

              <tr>
                <td class="left_align"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr v-for="item in this.data['route_relays']" :key="item.id">
                <td class="left_align"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="right_align"> {{ item.address }} </td>
              </tr>

              <tr>
                <td class="left_align"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

          <table id="route_table" class="table" v-else>

            <tbody>

              <tr>
                <td class="left_align"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr>
                <td class="left_align"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

          <div v-if="this.data['client_relays'] != null && this.data['client_relays'].length > 0">

            <p class="header" style="padding-top: 25px; padding-bottom: 15px">Client Relays</p>
     
            <table class="table">

              <tbody>

                <tr v-for="item in this.data['client_relays']" :key="item.id">
                  <td class="left_align"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                  <td class="left_align"> {{ item.rtt }}ms </td>
                  <td class="left_align"> {{ item.jitter }}ms </td>
                  <td class="left_align"> {{ item.packet_loss}}% </td>
                </tr>

              </tbody>

            </table>

          </div>

          <div v-if="this.data['server_relays'] != null && this.data['server_relays'].length > 0">

            <p class="header" style="padding-top: 25px; padding-bottom: 15px">Server Relays</p>
     
            <table class="table">

              <tbody>

                <tr v-for="item in this.data['server_relays']" :key="item.id">
                  <td class="left_align"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                  <td class="left_align"> {{ item.rtt }}ms </td>
                  <td class="left_align"> {{ item.jitter }}ms </td>
                  <td class="left_align"> {{ item.packet_loss}}% </td>
                </tr>

              </tbody>

            </table>

          </div>

        </div>

      </div>

      <div id="right" class="right d-none d-xxl-block">

<!--
        <div class="right-top">

          <div class="map"/>

        </div>
-->
  
        <div class="right-bottom">
   
          <div class="session_info">

            <table id="session_table" class="table table-striped" style="vertical-align: middle;">
              <tbody>

                <tr>
                  <td class="bold">Datacenter</td>
                  <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Server</td>
                  <td> <router-link :to="'/server/' + this.data['server_address']"> {{ this.data['server_address'] }} </router-link> </td>
                </tr>
                
                <tr>
                  <td class="bold">ISP</td>
                  <td> {{ this.data['isp'] }} </td>
                </tr>

                <tr>
                  <td class="bold">Platform</td>
                  <td> {{ this.data['platform'] }} </td>
                </tr>

                <tr>
                  <td class="bold">Connection</td>
                  <td> {{ this.data['connection'] }} </td>
                </tr>

                <tr>
                  <td class="bold">User Hash</td>
                  <td class="fixed"> <router-link :to="'/user/' + this.data['user_hash']"> {{ this.data['user_hash'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Buyer</td>
                  <td> <router-link :to="'/buyer/' + this.data['buyer_code']"> {{ this.data['buyer_name'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Start Time</td>
                  <td> {{ this.data['start_time'] }} </td>
                </tr>

              </tbody>
            </table>

          </div>

        </div>

        <div class="route_info">

          <p class="bold tight-p">Route</p>
   
          <table id="route_table" class="table" v-if="this.data['route_relays'] != null && this.data['route_relays'].length > 0">

            <tbody>

              <tr>
                <td class="left_align bold"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr v-for="item in this.data['route_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="right_align"> {{ item.address }} </td>
              </tr>

              <tr>
                <td class="left_align bold"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

          <table id="route_table" class="table" v-else>

            <tbody>

              <tr>
                <td class="left_align bold"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr>
                <td class="left_align bold"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

        </div>

        <div v-if="this.data['client_relays'] != null && this.data['client_relays'].length > 0" class="client_relay_info">

          <p class="bold">Client Relays</p>
   
          <table class="table">

            <tbody>

              <tr v-for="item in this.data['client_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="left_align"> {{ item.rtt }}ms </td>
                <td class="left_align"> {{ item.jitter }}ms </td>
                <td class="left_align"> {{ item.packet_loss}}% </td>
              </tr>

            </tbody>

          </table>

        </div>

        <div v-if="this.data['server_relays'] != null && this.data['server_relays'].length > 0" class="server_relay_info">

          <p class="bold">Server Relays</p>
   
          <table class="table">

            <tbody>

              <tr v-for="item in this.data['server_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="left_align"> {{ item.rtt }}ms </td>
                <td class="left_align"> {{ item.jitter }}ms </td>
                <td class="left_align"> {{ item.packet_loss}}% </td>
              </tr>

            </tbody>

          </table>

        </div>

      </div>

    </div>

  </div>

</template>

<script>

import axios from "axios";
import update from '@/update.js'
import uPlot from "uplot";

import {parse_uint64, is_visible, custom_graph} from '@/utils.js'

let latency_opts = custom_graph({
  title: "Latency",
  series: [
    { 
      name: 'Direct',
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      units: 'ms',
    },
    {
      name: 'Next',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: 'ms',
    },
    {
      name: 'Predicted',
      stroke: "orange",
      fill: "rgba(0,0,0,0)",
      units: "ms",
    },
  ]
})

let jitter_opts = custom_graph({
  title: "Jitter",
  series: [
    { 
      name: 'Direct',
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      units: 'ms',
    },
    {
      name: 'Next',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: 'ms',
    },
    {
      name: 'Real',
      stroke: "purple",
      fill: "rgba(10,10,10,0.035)",
      units: "ms",
    },
  ]
})

let packet_loss_opts = custom_graph({
  title: "Packet Loss",
  percent: true,
  series: [
    {
      name: 'Real',
      stroke: "rgb(200,10,10)",
      fill: "rgba(10,10,10,0.035)",
      units: "%",
    },
  ]
})

let out_of_order_opts = custom_graph({
  title: "Out of Order",
  percent: true,
  series: [
    {
      name: 'Real',
      stroke: "#ffcc00",
      fill: "rgba(10,10,10,0.035)",
      units: "%",
    },
  ]
})

let bandwidth_up_opts = custom_graph({
  title: "Bandwidth Up",
  series: [
    { 
      name: 'Direct',
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      units: 'kbps',
    },
    {
      name: 'Next',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: 'kbps',
    },
  ]
})

let bandwidth_down_opts = custom_graph({
  title: "Bandwidth Down",
  series: [
    { 
      name: 'Direct',
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      units: 'kbps',
    },
    {
      name: 'Next',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: 'kbps',
    },
  ]
})

function getPlatformName(platformId) {
  switch(platformId) {
  case 1: return "Windows"
  case 2: return "Mac"
  case 3: return "Linux"
  case 4: return "Nintendo Switch"
  case 5: return "PS4"
  case 6: return "iOS"
  case 7: return "Xbox One"
  case 8: return "Xbox Series X"
  case 9: return "PS5"
  default:
    return "Unknown"
  }
}

function getConnectionName(connectionType) {
  switch(connectionType) {
  case 1: return "Wired"
  case 2: return "Wi-Fi"
  case 3: return "Cellular"
  default:
    return "Unknown"
  }
}

async function getData(page, session_id) {
  try {

    if (page == null) {
      page = 0
    }
  
    const url = process.env.VUE_APP_API_URL + '/portal/session/' + session_id
  
    const res = await axios.get(url);
  
    let data = {}
  
    if (res.data.slice_data !== null) {

      // get session data

      let session_data = res.data.session_data

      data['session_id'] = parse_uint64(session_data.session_id)
      data["datacenter_name"] = session_data.datacenter_name
      data["isp"] = session_data.isp
      data["buyer_code"] = session_data.buyer_code
      data["buyer_name"] = session_data.buyer_name
      data["user_hash"] = parse_uint64(session_data.user_hash)
      data["platform"] = getPlatformName(session_data.platform_type)
      data["connection"] = getConnectionName(session_data.connection_type)
      data["start_time"] = new Date(parseInt(session_data.start_time)*1000).toLocaleString()
      data["server_address"] = session_data.server_address
    
      // route relays

      if (session_data.num_route_relays > 0) {
        let i = 0
        let route_relays = []
        while (i < session_data.num_route_relays) {
          route_relays.push({
            id:        session_data.route_relay_ids[i],
            name:      session_data.route_relay_names[i],
            address:   session_data.route_relay_addresses[i],
          })
          i++
        }
        data['route_relays'] = route_relays
      }

      // client relays
  
      let client_relay_data = res.data.client_relay_data
      if (client_relay_data.length > 0) {
        client_relay_data = client_relay_data[client_relay_data.length-1]
        let i = 0
        let client_relays = []
        while (i < client_relay_data.num_client_relays) {
          if (client_relay_data.client_relay_rtt[i] != 0) {
            client_relays.push({
              id:          client_relay_data.client_relay_id[i],
              name:        client_relay_data.client_relay_name[i],
              rtt:         client_relay_data.client_relay_rtt[i],
              jitter:      client_relay_data.client_relay_jitter[i],
              packet_loss: client_relay_data.client_relay_packet_loss[i],
            })
          }
          i++
        }
        client_relays.sort( function(a,b) {
          if (a.name < b.name) {
            return -1
          }
          if (a.name > b.name) {
            return +1
          }
          return 0
        })
        data['client_relays'] = client_relays
      }

      // server relays
  
      let server_relay_data = res.data.server_relay_data
      if (server_relay_data.length > 0) {
        server_relay_data = server_relay_data[server_relay_data.length-1]
        let i = 0
        let server_relays = []
        while (i < server_relay_data.num_server_relays) {
          if (server_relay_data.server_relay_rtt[i] != 0) {
            server_relays.push({
              id:          server_relay_data.server_relay_id[i],
              name:        server_relay_data.server_relay_name[i],
              rtt:         server_relay_data.server_relay_rtt[i],
              jitter:      server_relay_data.server_relay_jitter[i],
              packet_loss: server_relay_data.server_relay_packet_loss[i],
            })
          }
          i++
        }
        server_relays.sort( function(a,b) {
          if (a.name < b.name) {
            return -1
          }
          if (a.name > b.name) {
            return +1
          }
          return 0
        })
        data['server_relays'] = server_relays
      }

      // timestamps (same for all graphs...)
  
      let graph_timestamps = []
      let i = 0
      while (i < res.data.slice_data.length) {
        const timestamp = parseInt(res.data.slice_data[i].timestamp)
        graph_timestamps.push(timestamp)
        i++
      }

      // latency graph data
  
      let latency_direct = []
      let latency_next = []
      let latency_predicted = []
      i = 0
      while (i < res.data.slice_data.length) {
        latency_direct.push(res.data.slice_data[i].direct_rtt)
        latency_next.push(res.data.slice_data[i].next_rtt)
        latency_predicted.push(res.data.slice_data[i].predicted_rtt)
        i++
      }

      data.latency_data = [graph_timestamps, latency_direct, latency_next, latency_predicted]

      // jitter graph data
  
      let jitter_direct = []
      let jitter_next = []
      let jitter_real = []
      i = 0
      while (i < res.data.slice_data.length) {
        jitter_direct.push(res.data.slice_data[i].direct_jitter)
        jitter_next.push(res.data.slice_data[i].next_jitter)
        jitter_real.push(res.data.slice_data[i].real_jitter)
        i++
      }

      data.jitter_data = [graph_timestamps, jitter_direct, jitter_next, jitter_real]

      // packet loss graph data
  
      let packet_loss_real = []
      i = 0
      while (i < res.data.slice_data.length) {
        packet_loss_real.push(res.data.slice_data[i].real_packet_loss)
        i++
      }

      data.packet_loss_data = [graph_timestamps, packet_loss_real]

      // out of order graph data
  
      let out_of_order_real = []
      i = 0
      while (i < res.data.slice_data.length) {
        out_of_order_real.push(res.data.slice_data[i].real_out_of_order)
        i++
      }

      data.out_of_order_data = [graph_timestamps, out_of_order_real]

      // bandwidth up graph data
  
      let bandwidth_up_direct = []
      let bandwidth_up_next = []
      i = 0
      while (i < res.data.slice_data.length) {
        bandwidth_up_direct.push(res.data.slice_data[i].direct_kbps_up)
        bandwidth_up_next.push(res.data.slice_data[i].next_kbps_up)
        i++
      }

      data.bandwidth_up_data = [graph_timestamps, bandwidth_up_direct, bandwidth_up_next]

      // bandwidth down graph data
  
      let bandwidth_down_direct = []
      let bandwidth_down_next = []
      i = 0
      while (i < res.data.slice_data.length) {
        bandwidth_down_direct.push(res.data.slice_data[i].direct_kbps_down)
        bandwidth_down_next.push(res.data.slice_data[i].next_kbps_down)
        i++
      }

      data.bandwidth_down_data = [graph_timestamps, bandwidth_down_direct, bandwidth_down_next]

      // mark data as found

      data["found"] = true
    }

    return [data, 0, 1]

  } catch (error) {
    
    // error

    console.log(error);
    
    let data = {}
    data['session_id'] = session_id
    data['found'] = false
    
    return [data, 0, 1]

  }
}

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      data: {},
      found: false,
      observer: null,
      prevWidth: 0,
      show_legend: false,
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let session_id = values[values.length-1]
    let result = await getData(0, session_id)
    next(vm => {
      if (result != null && !result.error) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
        vm.$emit('notify-update', vm.page, vm.num_pages)
        vm.found = result[0]['found']
        vm.updateGraphs()
      }
    })
  },

  mounted: function () {
  
    this.latency = new uPlot(latency_opts, [[],[],[]], document.getElementById('latency_graph'))
    this.jitter = new uPlot(jitter_opts, [[],[],[]], document.getElementById('jitter_graph'))
    this.packet_loss = new uPlot(packet_loss_opts, [[],[],[]], document.getElementById('packet_loss_graph'))
    this.out_of_order = new uPlot(out_of_order_opts, [[]], document.getElementById('out_of_order_graph'))
    this.bandwidth_up = new uPlot(bandwidth_up_opts, [[],[]], document.getElementById('bandwidth_up_graph'))
    this.bandwidth_down = new uPlot(bandwidth_down_opts, [[],[]], document.getElementById('bandwidth_down_graph'))

    this.observer = new ResizeObserver(this.resize)
    this.observer.observe(document.body, {box: 'border-box'})

    document.getElementById("session-id-input").value = document.getElementById("session-id-input").defaultValue = this.data['session_id']
    document.getElementById("session-id-input").addEventListener('keyup', this.onKeyUp);

    this.$emit('notify-view', 'session')

    this.updateGraphs()
  },

  beforeUnmount() {
    document.getElementById("session-id-input").removeEventListener('keyup', this.onKeyUp);
    this.latency.destroy()
    this.jitter.destroy()
    this.packet_loss.destroy()
    this.out_of_order.destroy()
    this.bandwidth_up.destroy()
    this.bandwidth_down.destroy()
    this.observer.disconnect()
    this.prevWidth = 0
    this.latency = null
    this.jitter = null
    this.packet_loss = null
    this.out_of_order = null
    this.bandwidth_up = null
    this.bandwidth_down = null
    this.observer = null
  },

  methods: {

    resize() {

      const right_visible = is_visible(document.getElementById('right'))
      const width = document.body.clientWidth;
      if (width !== this.prevWidth) {

        // resize the graphs to match the page width

        this.prevWidth = width;
        if (this.latency) {
          let graph_width = width
          if (right_visible) {
            graph_width -= 540
          } else {
            graph_width -= 25
          }
          let graph_height = graph_width * 0.4
          if (graph_height > 500) {
            graph_height = 500
          } else if (graph_height < 250) {
            graph_height = 250
          }
          this.latency.setSize({width: graph_width, height: graph_height})
          this.jitter.setSize({width: graph_width, height: graph_height})
          this.packet_loss.setSize({width: graph_width, height: graph_height})
          this.out_of_order.setSize({width: graph_width, height: graph_height})
          this.bandwidth_up.setSize({width: graph_width, height: graph_height})
          this.bandwidth_down.setSize({width: graph_width, height: graph_height})
        }
      }    

      // show legends in desktop, hide them in mobile layout

      this.show_legend = right_visible
      var elements = document.getElementsByClassName('u-legend');
      let i = 0;
      while (i < elements.length) {
        if (this.show_legend) {
          elements[i].style.display = 'block';
        } else {
          elements[i].style.display = 'none';
        }
        i++;
      }
    },

    async getData(page, session_id) {
      if (session_id == null) {
        session_id = this.$route.params.id
      }
      return getData(page, session_id)
    },

    async update() {
      let result = await getData(this.page, this.$route.params.id)
      if (result != null) {
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
        this.found = result[0]['found']
        this.$emit('notify-update', this.page, this.num_pages)
        this.updateGraphs()
      }
    },

    updateGraphs() {
      if (this.latency != null && this.data.latency_data != null) {
        this.latency.setData(this.data.latency_data, true)
      }
      if (this.jitter != null && this.data.jitter_data != null) {
        this.jitter.setData(this.data.jitter_data, true)
      }
      if (this.packet_loss != null && this.data.packet_loss_data != null) {
        this.packet_loss.setData(this.data.packet_loss_data, true)
      }
      if (this.out_of_order != null && this.data.out_of_order_data != null) {
        this.out_of_order.setData(this.data.out_of_order_data, true)
      }
      if (this.bandwidth_up != null && this.data.bandwidth_up_data != null) {
        this.bandwidth_up.setData(this.data.bandwidth_up_data, true)
      }
      if (this.bandwidth_down != null && this.data.bandwidth_down_data != null) {
        this.bandwidth_down.setData(this.data.bandwidth_down_data, true)
      }
    },

    search() {
      const session_id = document.getElementById("session-id-input").value
      this.$router.push('/session/' + session_id)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

  },

};

</script>

<style>

.fixed {
  font-family: monospace;
}

.parent {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding-top: 20px;
  padding-bottom: 50px;
}

.bottom {
  height: 100%;  
  display: flex;
  flex-direction: row;
  padding: 0px;
  justify-content: space-between;
  gap: 15px;
}

.left {
  width: 100%;
  height: 100%;
  padding: 0px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: v-bind("show_legend ? '35px' : '20px'");
  padding-top: v-bind("show_legend ? '10px' : '0px'");
}

.graph {
  height: 100%;
}

.right {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 0px;
  max-width: 500px;
  min-width: 300px;
  padding-right: v-bind("show_legend ? '15px' : '0px'");
}

.search {
  width: 100%;
  height: 35px;
  display: flex;
  flex-direction: row;
  align-items: center;
  align-content: center;
  justify-content: space-between;
  gap: 15px;
  font-weight: 1;
  font-size: 18px;
  padding: 0px;
  padding-left: 15px;
  padding-right: 15px;
}

.text {
  width: 100%;
  height: 35px;
  font-size: 15px;
  padding-left: 5px;
}

.test-text {
  width: 100%;
}

.right-top {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.right-bottom {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 50px;
}

.map {
  background-color: #555555;
  width: 100%;
  height: 500px;
  flex-shrink: 0;
}

.session_info {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  padding-top: 15px;
}

.route_info {
  width: 100%;
  flex-direction: column;
  justify-content: space-between;
  padding-top: 25px;
}

.client_relay_info {
  width: 100%;
  flex-direction: column;
  justify-content: space-between;
  padding-top: 25px;
}

.server_relay_info {
  width: 100%;
  flex-direction: column;
  justify-content: space-between;
  padding-top: 25px;
}

.header {
  font-weight: bold;
  font-size: 18px;
}

.bold {
  font-weight: bold;
}

button {
  font-size: 15px;
}

.tight-p {
  line-height: 15px;
  margin-bottom: 2px;
}

a {
  color: #2c3e50;
  text-decoration: none;
}

.u-title {
  font-family: "Montserrat";
}

.session-info-mobile {
  padding-left: 15px;
  padding-right: 15px;
}

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
}

</style>
