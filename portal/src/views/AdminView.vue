<template>

  <div class="parent" id="parent">
  
    <div class="left">

      <div id="total_sessions" class="graph"/>
    
      <div id="accelerated_percent" class="graph"/>

      <div id="session_update" class="graph"/>

    </div>

    <div class="right">

      <div id="next_sessions" class="graph"/>
    
      <div id="server_count" class="graph"/>

    </div>

  </div>

</template>

<script>

import axios from "axios";
import update from '@/update.js'
import uPlot from "uplot";

import { custom_graph } from '@/utils.js'

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
  title: "Next Sessions",
  series: [
    { 
      name: 'Next Sessions',
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      units: '',
    },
  ]
})

let accelerated_percent_opts = custom_graph({
  title: "Accelerated Percent",
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

let session_update_opts = custom_graph({
  title: "Session Updates",
  series: [
    { 
      name: 'Session Updates',
      stroke: "#faac02",
      fill: "rgba(250, 172, 2,0.075)",
      units: ' per-second',
    },
  ]
})

async function getData() {

  try {

    const url = process.env.VUE_APP_API_URL + '/portal/admin_data'

    const res = await axios.get(url);

    let data = {}

    // total sessions data

    let total_sessions_timestamps = []  
    let total_sessions_values = []
    let i = 0
    while (i < res.data.time_series_total_sessions_timestamps.length) {
      total_sessions_timestamps.push(Math.floor(parseInt(res.data.time_series_total_sessions_timestamps[i]) / 1000000000))
      total_sessions_values.push(parseInt(res.data.time_series_total_sessions_values[i]))
      i++
    }
    data.total_sessions_data = [total_sessions_timestamps, total_sessions_values]

    // next sessions data

    let next_sessions_timestamps = []  
    let next_sessions_values = []
    i = 0
    while (i < res.data.time_series_next_sessions_timestamps.length) {
      next_sessions_timestamps.push(Math.floor(parseInt(res.data.time_series_next_sessions_timestamps[i]) / 1000000000))
      next_sessions_values.push(parseInt(res.data.time_series_next_sessions_values[i]))
      i++
    }
    data.next_sessions_data = [next_sessions_timestamps, next_sessions_values]

    // accelerated percent data

    let accelerated_percent_timestamps = []  
    let accelerated_percent_values = []
    i = 0
    while (i < res.data.time_series_accelerated_percent_timestamps.length) {
      accelerated_percent_timestamps.push(Math.floor(parseInt(res.data.time_series_accelerated_percent_timestamps[i]) / 1000000000))
      accelerated_percent_values.push(parseInt(res.data.time_series_accelerated_percent_values[i]))
      i++
    }
    data.accelerated_percent_data = [accelerated_percent_timestamps, accelerated_percent_values]

    // server count data

    let server_count_timestamps = []  
    let server_count_values = []
    i = 0
    while (i < res.data.time_series_server_count_timestamps.length) {
      server_count_timestamps.push(Math.floor(parseInt(res.data.time_series_server_count_timestamps[i]) / 1000000000))
      server_count_values.push(parseInt(res.data.time_series_server_count_values[i]))
      i++
    }
    data.server_count_data = [server_count_timestamps, server_count_values]

    // session update data

    let session_update_timestamps = []  
    let session_update_values = []
    i = 0
    while (i < res.data.counters_session_update_timestamps.length) {
      session_update_timestamps.push(Math.floor(parseInt(res.data.counters_session_update_timestamps[i]) / 1000))
      session_update_values.push(parseInt(res.data.counters_session_update_values[i]))
      i++
    }
    data.session_update_data = [session_update_timestamps, session_update_values]

    data['found'] = true

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
      show_legend: false,
    };
  },

  async beforeRouteEnter (to, from, next) {
    let result = await getData()
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
    this.session_update = new uPlot(session_update_opts, [[],[]], document.getElementById('session_update'))

    this.observer = new ResizeObserver(this.resize)
    this.observer.observe(document.body, {box: 'border-box'})

    this.$emit('notify-view', 'admin')
  
    this.updateGraphs()
  },

  beforeUnmount() {
    this.total_sessions.destroy()
    this.next_sessions.destroy()
    this.accelerated_percent.destroy()
    this.server_count.destroy()
    this.session_update.destroy()
    this.observer.disconnect()
    this.prevWidth = 0
    this.total_sessions = null
    this.next_sessions = null
    this.accelerated_percent = null
    this.server_count = null
    this.session_update = null
    this.observer = null
  },

  methods: {

    resize() {
      const width = document.body.clientWidth;
      if (width !== this.prevWidth) {

        // resize graphs to match page width

        this.prevWidth = width;
        if (this.total_sessions) {
          let graph_width = width / 2
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
          this.session_update.setSize({width: graph_width, height: graph_height})
        }

        // show legends in desktop, hide them in mobile layout

        this.show_legend = width > 1000
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

    async getData() {
      let data = getData()
      this.updateGraphs()           // todo: data is actually out of date here (delayed...)
      return data
    },

    async update() {
      let result = await getData()
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
      if (this.session_update != null && this.data.session_update_data != null) {
        this.session_update.setData(this.data.session_update_data, true)
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
  flex-direction: row;
  gap: 15px;
  padding: 15px;
  padding-top: 35px;
}

.left {
  width: 50%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.right {
  width: 50%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.graph {
  width: 100%;
  height: 100%;
}

</style>
