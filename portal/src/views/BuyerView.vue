<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Buyer</p>
      <p class="tight-p test-text"><input id='buyer-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-show="this.found" class="bottom-buyer">

      <div id="left" class="left">

        <div class="d-xxl-none">
          <table id="buyer_table" class="table table-striped" style="vertical-align: middle;">
            <tbody>

              <tr>
                <td class="bold">Total Sessions</td>
                <td> {{ this.data.total_sessions }} </td>
              </tr>

              <tr>
                <td class="bold">Accelerated Sessions</td>
                <td> {{ this.data.next_sessions }} </td>
              </tr>

              <tr>
                <td class="bold">Accelerated</td>
                <td> {{ this.data.accelerated_percent }}% </td>
              </tr>

              <tr>
                <td class="bold">Servers</td>
                <td> {{ this.data.servers }} </td>
              </tr>

            </tbody>
          </table>
        </div>

        <div id="total_sessions" class="graph"/>
        
        <div id="next_sessions" class="graph"/>
        
        <div id="accelerated_percent" class="graph"/>

        <div id="server_count" class="graph"/>

      </div>

    </div>

  </div>

</template>

<script>

import axios from "axios";
import update from '@/update.js'
import uPlot from "uplot";

import { getAcceleratedPercent, custom_graph } from '@/utils.js'

let total_sessions_opts = custom_graph({
  title: "Total Sessions",
  series: [
    { 
      name: 'Total Sessions',
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      units: '',
    },
  ]
})

let next_sessions_opts = custom_graph({
  title: "Accelerated Sessions",
  series: [
    { 
      name: 'Accelerated Sessions',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: '',
    },
  ]
})

let accelerated_percent_opts = custom_graph({
  title: "Accelerated Percent",
  percent: true,
  series: [
    { 
      name: 'Accelerated',
      stroke: "#8350ba",
      fill: "rgba(131, 80, 186,0.1)",
      units: '%',
    },
  ]
})

let server_count_opts = custom_graph({
  title: "Server Count",
  series: [
    { 
      name: 'Servers',
      stroke: "#faac02",
      fill: "rgba(250, 172, 2,0.075)",
      units: '',
    },
  ]
})

async function getData(page, buyer) {

  try {

    if (page == null) {
      page = 0
    }

    const url = process.env.VUE_APP_API_URL + '/portal/buyer/' + buyer

    const res = await axios.get(url);

    let data = {}

    data['buyer'] = buyer

    if (res.data.buyer_data !== null) {

      // buyer data

      data['live'] = res.data.buyer_data.live
      data['debug'] = res.data.buyer_data.debug
      data['total_sessions'] = res.data.buyer_data.total_sessions.toLocaleString()
      data['next_sessions'] = res.data.buyer_data.next_sessions.toLocaleString()
      data['accelerated_percent'] = getAcceleratedPercent(res.data.buyer_data.next_sessions, res.data.buyer_data.total_sessions)
      data['servers'] = res.data.buyer_data.server_count.toLocaleString()

      // total sessions data

      if (res.data.buyer_data.total_sessions_timestamps != null) {
        let total_sessions_timestamps = []  
        let total_sessions_values = []
        let i = 0
        while (i < res.data.buyer_data.total_sessions_timestamps.length) {
          total_sessions_timestamps.push(Math.floor(parseInt(res.data.buyer_data.total_sessions_timestamps[i]) / 1000))
          total_sessions_values.push(parseInt(res.data.buyer_data.total_sessions_values[i]))
          i++
        }
        data.total_sessions_data = [total_sessions_timestamps, total_sessions_values]
      }

      // next sessions data

      if (res.data.buyer_data.next_sessions_timestamps != null) {
        let next_sessions_timestamps = []  
        let next_sessions_values = []
        let i = 0
        while (i < res.data.buyer_data.next_sessions_timestamps.length) {
          next_sessions_timestamps.push(Math.floor(parseInt(res.data.buyer_data.next_sessions_timestamps[i]) / 1000))
          next_sessions_values.push(parseInt(res.data.buyer_data.next_sessions_values[i]))
          i++
        }
        data.next_sessions_data = [next_sessions_timestamps, next_sessions_values]
      }

      // accelerated percent data

      if (res.data.buyer_data.accelerated_percent_timestamps != null) {
        let accelerated_percent_timestamps = []  
        let accelerated_percent_values = []
        let i = 0
        while (i < res.data.buyer_data.accelerated_percent_timestamps.length) {
          accelerated_percent_timestamps.push(Math.floor(parseInt(res.data.buyer_data.accelerated_percent_timestamps[i]) / 1000))
          accelerated_percent_values.push(parseInt(res.data.buyer_data.accelerated_percent_values[i]))
          i++
        }
        data.accelerated_percent_data = [accelerated_percent_timestamps, accelerated_percent_values]
      }

      // server count data

      if (res.data.buyer_data.server_count_timestamps != null) {
        let server_count_timestamps = []  
        let server_count_values = []
        let i = 0
        while (i < res.data.buyer_data.server_count_timestamps.length) {
          server_count_timestamps.push(Math.floor(parseInt(res.data.buyer_data.server_count_timestamps[i]) / 1000))
          server_count_values.push(parseInt(res.data.buyer_data.server_count_values[i]))
          i++
        }
        data.server_count_data = [server_count_timestamps, server_count_values]
      }

      data['found'] = true
    }
    return [data, 0, 1]
  } catch (error) {
    console.log(error);
    let data = {}
    data['buyer'] = buyer
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
    let buyer = values[values.length-1]
    let result = await getData(0, buyer)
    next(vm => {
      if (result != null && !result.error) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
        vm.$emit('notify-update', vm.page, vm.num_pages)
        vm.found = result[0]['found']
      }
    })
  },

  mounted: function () {
  
    this.total_sessions = new uPlot(total_sessions_opts, [[],[]], document.getElementById('total_sessions'))
    this.next_sessions = new uPlot(next_sessions_opts, [[],[]], document.getElementById('next_sessions'))
    this.accelerated_percent = new uPlot(accelerated_percent_opts, [[],[]], document.getElementById('accelerated_percent'))
    this.server_count = new uPlot(server_count_opts, [[],[]], document.getElementById('server_count'))

    this.observer = new ResizeObserver(this.resize)
    this.observer.observe(document.body, {box: 'border-box'})

    document.getElementById("buyer-input").value = document.getElementById("buyer-input").defaultValue = this.data['buyer']
    document.getElementById("buyer-input").addEventListener('keyup', this.onKeyUp);

    this.$emit('notify-view', 'buyer')
  
    this.updateGraphs()
  },

  beforeUnmount() {
    document.getElementById("buyer-input").removeEventListener('keyup', this.onKeyUp);
    this.total_sessions.destroy()
    this.next_sessions.destroy()
    this.accelerated_percent.destroy()
    this.server_count.destroy()
    this.observer.disconnect()
    this.prevWidth = 0
    this.total_sessions = null
    this.next_sessions = null
    this.accelerated = null
    this.servers = null
    this.observer = null
  },

  methods: {

    resize() {

      const width = document.body.clientWidth;
      if (width !== this.prevWidth) {

        // resize graphs to match page width

        this.prevWidth = width;
        if (this.total_sessions) {
          let graph_width = width;
          graph_width -= 30
          let graph_height = graph_width * 0.333
          if (graph_height > 450) {
            graph_height = 450
          } else if (graph_height < 250) {
            graph_height = 250
          }
          this.total_sessions.setSize({width: graph_width, height: graph_height})
          this.next_sessions.setSize({width: graph_width, height: graph_height})
          this.accelerated_percent.setSize({width: graph_width, height: graph_height})
          this.server_count.setSize({width: graph_width, height: graph_height})
        }
      }    

      // show legends in desktop, hide them in mobile layout

      this.show_legend = width > 1000;
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

    async getData(page, buyer) {
      if (buyer == null) {
        buyer = this.$route.params.id
      }
      let data = getData(page, buyer)
      this.updateGraphs()
      return data
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
      if (this.total_sessions != null && this.data.total_sessions_data != null) {
        this.total_sessions.setData(this.data.total_sessions_data, true)
      }
      if (this.next_sessions != null && this.data.next_sessions_data != null) {
        this.next_sessions.setData(this.data.next_sessions_data, true)
      }
      if (this.accelerated_percent != null && this.data.accelerated_percent_data != null) {
        this.accelerated_percent.setData(this.data.accelerated_percent_data, true)
      }
      if (this.server_count != null && this.data.server_count_data != null) {
        this.server_count.setData(this.data.server_count_data, true)
      }
    },

    search() {
      const buyer = document.getElementById("buyer-input").value
      this.$router.push('/buyer/' + buyer)
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
  padding: 15px;
  padding-top: 20px;
}

.bottom-buyer {
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

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
}

.near_relay_info {
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

.u-title {
  font-family: "Montserrat";
}

</style>
