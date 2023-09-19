// -----------------------------------------------------------------------------------------

<template>

  <div class="d-md-none">
    <table id="relays_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Relay Name</th>
          <th>Sessions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Relay Link"]'> {{ item["Relay Name"] }} </router-link> </td>
          <td> {{ item["Current Sessions"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-md-block d-xxl-none">
    <table id="relays_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Relay Name</th>
          <th>Current Sessions</th>
          <th>Status</th>
          <th>Uptime</th>
          <th>Datacenter</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Relay Link"]'> {{ item["Relay Name"] }} </router-link> </td>
          <td> {{ item["Current Sessions"] }} </td>
          <td> {{ item["Status"] }} </td>
          <td> {{ item["Uptime"] }} </td>
          <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
        </tr>
      </tbody>
    </table>
  </div>


  <div class="d-none d-xxl-block">
    <table id="relays_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Relay Name</th>
          <th>Current Sessions</th>
          <th>Status</th>
          <th>Uptime</th>
          <th>Relay Version</th>
          <th>Public Address</th>
          <th>Datacenter</th>
          <th>Seller</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Relay Link"]'> {{ item["Relay Name"] }} </router-link> </td>
          <td> {{ item["Current Sessions"] }} </td>
          <td> {{ item["Status"] }} </td>
          <td> {{ item["Uptime"] }} </td>
          <td> {{ item["Relay Version"] }} </td>
          <td> {{ item["Public Address"] }} </td>
          <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
          <td> <router-link :to='item["Seller Link"]'> {{ item["Seller"] }} </router-link> </td>
        </tr>
      </tbody>
    </table>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios"
import utils from "@/utils.js"
import update from "@/update.js"

function nice_uptime(value) {
  if (isNaN(value)) {
    return ''
  }
  if (value > 86400) {
    return Math.floor(value/86400).toLocaleString() + "d"
  }
  if (value > 3600) {
    return Math.floor(value/3600).toLocaleString() + "h"
  }
  if (value > 60) {
    return Math.floor(value/60).toLocaleString() + "m"
  }
  return value + "s"
}

async function getData() {
  try {
    const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/relays/0/100');
    res.data.relays.sort(function(a, b){return b.num_sessions-a.num_sessions});
    let i = 0
    let data = []
    while (i < res.data.relays.length) {
      const v = res.data.relays[i]
      var status = "Online"
      if (v.relay_flags & 1) {
        status = "Shutting Down"
      }
      let row = {
        "Relay Name":v.relay_name,
        "Relay Link":"relay/" + v.relay_name,
        "Public Address":v.relay_address,
        "Datacenter":v.datacenter_name,
        "Datacenter Link":"datacenter/" + v.datacenter_name,
        "Seller":v.seller_name,
        "Seller Link":"seller/" + v.seller_code,
        "Current Sessions":v.num_sessions,
        "Status":status,
        "Uptime":nice_uptime(v.uptime),
        "Relay Version":v.relay_version
      }
      data.push(row)
      i++
    }
    return data
  } catch (error) {
    console.log(error);
    return null
  }
}

export default {

  name: "App",

  mixins: [update, utils],

  data() {
    return {
      data: []
    };
  },

  async beforeRouteEnter (to, from, next) {
    var data = await getData()
    next(vm => {
      vm.data = data
    })
  },

  methods: {

    async update() {
      this.data = await getData()
    }

  }

};

</script>

// -----------------------------------------------------------------------------------------

<style scoped>

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

</style>

// -----------------------------------------------------------------------------------------
