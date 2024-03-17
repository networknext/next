<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Relay</p>
      <p class="tight-p test-text"><input id='relay-name-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-show="this.found" class="bottom">

      <div id="left" class="left">

        <div class="d-xxl-none">
          <table id="relay_table" class="table table-striped" style="vertical-align: middle;">
            <tbody>

              <tr>
                <td class="bold">Sessions</td>
                <td> {{ this.data['sessions'] }} </td>
              </tr>

              <tr>
                <td class="bold">Datacenter</td>
                <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Seller</td>
                <td> <router-link :to="'/seller/' + this.data['seller_code']"> {{ this.data['seller_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Version</td>
                <td> {{ this.data['version'] }} </td>
              </tr>

              <tr>
                <td class="bold">Uptime</td>
                <td> {{ this.data['uptime'] }} </td>
              </tr>

            </tbody>
          </table>
        </div>

        <div id="sessions" class="graph"/>
        
        <div id="bandwidth_sent" class="graph"/>

        <div id="bandwidth_received" class="graph"/>

        <div id="packets_sent" class="graph"/>

        <div id="packets_received" class="graph"/>

      </div>

      <div id="right" class="right d-none d-xxl-block">

<!--

        <div class="right-top">

          <div class="map"/>

        </div>
-->

        <div class="right-bottom">
   
          <div class="relay_info">

            <table id="relay_table" class="table table-striped" style="vertical-align: middle;">
              <tbody>

                <tr>
                  <td class="bold">Sessions</td>
                  <td> {{ this.data['sessions'] }} </td>
                </tr>

                <tr>
                  <td class="bold">Datacenter</td>
                  <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Seller</td>
                  <td> <router-link :to="'/seller/' + this.data['seller_code']"> {{ this.data['seller_name'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Version</td>
                  <td> {{ this.data['version'] }} </td>
                </tr>

                <tr>
                  <td class="bold">Uptime</td>
                  <td> {{ this.data['uptime'] }} </td>
                </tr>

              </tbody>
            </table>

          </div>

        </div>

      </div>

    </div>

  </div>

</template>

<script>

import axios from "axios";
import update from '@/update.js'
import uPlot from "uplot";

import {nice_uptime, is_visible, custom_graph} from '@/utils.js'

let sessions_opts = custom_graph({
  title: "Sessions",
  series: [
    { 
      name: 'Sessions',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: '',
    },
  ]
})

let bandwidth_sent_opts = custom_graph({
  title: "Bandwidth Sent",
  series: [
    { 
      name: 'Bandwidth Sent',
      stroke: "orange",
      fill: "rgba(255, 128, 0,0.1)",
      units: 'kbps',
    },
  ]
})

let bandwidth_received_opts = custom_graph({
  title: "Bandwidth Received",
  series: [
    { 
      name: 'Bandwidth Received',
      stroke: "#f5d742",
      fill: "rgba(245, 215, 60, 0.1)",
      units: 'kbps',
    },
  ]
})

let packets_sent_opts = custom_graph({
  title: "Packets Sent",
  series: [
    { 
      name: 'Packets Sent',
      stroke: "rgb(54, 141, 227)",
      fill: "rgba(54, 141, 227,0.1)",
      units: ' per-second',
    },
  ]
})

let packets_received_opts = custom_graph({
  title: "Packets Received",
  series: [
    { 
      name: 'Packets Received',
      stroke: "rgb(115, 158, 201)",
      fill: "rgba(115, 158, 201,0.1)",
      units: ' per-second',
    },
  ]
})

async function getData(page, relay_name) {

  try {

    if (page == null) {
      page = 0
    }

    const url = process.env.VUE_APP_API_URL + '/portal/relay/' + relay_name

    const res = await axios.get(url);

    let data = {}
    
    data['relay_name'] = relay_name

    if (res.data.relay_data !== null) {

      // relay data

      data["sessions"] = res.data.relay_data.num_sessions
      data["seller_name"] = res.data.relay_data.seller_name
      data["seller_code"] = res.data.relay_data.seller_code
      data["datacenter_name"] = res.data.relay_data.datacenter_name
      data["version"] = res.data.relay_data.relay_version                
      data["uptime"] = nice_uptime(res.data.relay_data.uptime)     
      data["latitude"] = res.data.relay_data.latitude              
      data["longitude"] = res.data.relay_data.longitude            

      // session count

      if (res.data.relay_data.session_count_timestamps != null) {
        let sessions_timestamps = []  
        let sessions_values = []
        let i = 0
        while (i < res.data.relay_data.session_count_timestamps.length) {
          sessions_timestamps.push(Math.floor(parseInt(res.data.relay_data.session_count_timestamps[i]) / 1000.0))
          sessions_values.push(parseInt(res.data.relay_data.session_count_values[i]))
          i++
        }
        data.sessions_data = [sessions_timestamps, sessions_values]
      }

      // bandwidth sent kbps

      if (res.data.relay_data.bandwidth_sent_kbps_timestamps != null) {
        let bandwidth_sent_timestamps = []  
        let bandwidth_sent_values = []
        let i = 0
        while (i < res.data.relay_data.bandwidth_sent_kbps_timestamps.length) {
          bandwidth_sent_timestamps.push(Math.floor(parseInt(res.data.relay_data.bandwidth_sent_kbps_timestamps[i]) / 1000.0))
          bandwidth_sent_values.push(parseInt(res.data.relay_data.bandwidth_sent_kbps_values[i]))
          i++
        }
        data.bandwidth_sent_data = [bandwidth_sent_timestamps, bandwidth_sent_values]
      }

      // bandwidth received kbps

      if (res.data.relay_data.bandwidth_received_kbps_timestamps != null) {
        let bandwidth_received_timestamps = []  
        let bandwidth_received_values = []
        let i = 0
        while (i < res.data.relay_data.bandwidth_received_kbps_timestamps.length) {
          bandwidth_received_timestamps.push(Math.floor(parseInt(res.data.relay_data.bandwidth_received_kbps_timestamps[i]) / 1000.0))
          bandwidth_received_values.push(parseInt(res.data.relay_data.bandwidth_received_kbps_values[i]))
          i++
        }
        data.bandwidth_received_data = [bandwidth_received_timestamps, bandwidth_received_values]
      }

      // packets sent per-second

      if (res.data.relay_data.packets_sent_per_second_timestamps != null) {
        let packets_sent_timestamps = []  
        let packets_sent_values = []
        let i = 0
        while (i < res.data.relay_data.packets_sent_per_second_timestamps.length) {
          packets_sent_timestamps.push(Math.floor(parseInt(res.data.relay_data.packets_sent_per_second_timestamps[i]) / 1000.0))
          packets_sent_values.push(parseInt(res.data.relay_data.packets_sent_per_second_values[i]))
          i++
        }
        data.packets_sent_data = [packets_sent_timestamps, packets_sent_values]
      }

      // packets received per-second

      if (res.data.relay_data.packets_received_per_second_timestamps != null) {
        let packets_received_timestamps = []  
        let packets_received_values = []
        let i = 0
        while (i < res.data.relay_data.packets_received_per_second_timestamps.length) {
          packets_received_timestamps.push(Math.floor(parseInt(res.data.relay_data.packets_received_per_second_timestamps[i]) / 1000.0))
          packets_received_values.push(parseInt(res.data.relay_data.packets_received_per_second_values[i]))
          i++
        }
        data.packets_received_data = [packets_received_timestamps, packets_received_values]
      }

      data["found"] = true
    }

    return [data, 0, 1]

  } catch (error) {
    console.log(error);
    let data = {}
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
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let relay_name = values[values.length-1]
    let result = await getData(0, relay_name)
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
  
    this.sessions = new uPlot(sessions_opts, [[],[]], document.getElementById('sessions'))
    this.bandwidth_sent = new uPlot(bandwidth_sent_opts, [[],[]], document.getElementById('bandwidth_sent'))
    this.bandwidth_received = new uPlot(bandwidth_received_opts, [[],[]], document.getElementById('bandwidth_received'))
    this.packets_sent = new uPlot(packets_sent_opts, [[],[]], document.getElementById('packets_sent'))
    this.packets_received = new uPlot(packets_received_opts, [[],[]], document.getElementById('packets_received'))

    this.observer = new ResizeObserver(this.resize)
    this.observer.observe(document.body, {box: 'border-box'})

    document.getElementById("relay-name-input").value = document.getElementById("relay-name-input").defaultValue = this.data['relay_name']
    document.getElementById("relay-name-input").addEventListener('keyup', this.onKeyUp);

    this.$emit('notify-view', 'relay')

    this.updateGraphs()
  },

  beforeUnmount() {
    document.getElementById("relay-name-input").removeEventListener('keyup', this.onKeyUp);
    this.sessions.destroy()
    this.bandwidth_sent.destroy()
    this.bandwidth_received.destroy()
    this.packets_sent.destroy()
    this.packets_received.destroy()
    this.prevWidth = 0
    this.sessions = null
    this.bandwidth_sent = null
    this.bandwidth_received = null
    this.packets_sent = null
    this.packets_received = null
    this.observer = null
  },

  methods: {

    resize() {
      const right_visible = is_visible(document.getElementById('right'))
      const width = document.body.clientWidth;
      if (width !== this.prevWidth) {
        this.prevWidth = width;
        if (this.sessions) {
          let graph_width = width
          if (right_visible) {
            graph_width -= 550
          } else {
            graph_width -= 30
          }
          let graph_height = graph_width * 0.333
          if (graph_height > 450) {
            graph_height = 450
          } else if (graph_height < 250) {
            graph_height = 250
          }
          this.sessions.setSize({width: graph_width, height: graph_height})
          this.bandwidth_sent.setSize({width: graph_width, height: graph_height})
          this.bandwidth_received.setSize({width: graph_width, height: graph_height})
          this.packets_sent.setSize({width: graph_width, height: graph_height})
          this.packets_received.setSize({width: graph_width, height: graph_height})
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
      }    
    },

    async getData(page, relay_name) {
      if (relay_name == null) {
        relay_name = this.$route.params.id
      }
      return getData(page, relay_name)
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
      if (this.sessions != null && this.data.sessions_data != null) {
        this.sessions.setData(this.data.sessions_data, true)
      }
      if (this.bandwidth_sent != null && this.data.bandwidth_sent_data != null) {
        this.bandwidth_sent.setData(this.data.bandwidth_sent_data, true)
      }
      if (this.bandwidth_received != null && this.data.bandwidth_received_data != null) {
        this.bandwidth_received.setData(this.data.bandwidth_received_data, true)
      }
      if (this.packets_sent != null && this.data.packets_sent_data != null) {
        this.packets_sent.setData(this.data.packets_sent_data, true)
      }
      if (this.packets_received != null && this.data.packets_received_data != null) {
        this.packets_received.setData(this.data.packets_received_data, true)
      }
    },

    search() {
      const relay_name = document.getElementById("relay-name-input").value
      this.$router.push('/relay/' + relay_name)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

  },

};

</script>

<style scoped>

.fixed {
  font-family: monospace;
}

.parent {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 15px;
  padding-top: 20px;
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
  gap: 25px;
  padding-top: 5px;
}

.graph {
  width: 100%;
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

.relay_info {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  padding-top: 15px;
}

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
}

.client_relay_info {
  width: 100%;
}

.server_relay_info {
  width: 100%;
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

</style>
