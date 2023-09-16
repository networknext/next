// -----------------------------------------------------------------------------------------

<template>

  <div class="d-md-none">
    <table id="servers_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Server Address</th>
          <th>Current Sessions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Server Link"]'> {{ item["Server Address"] }} </router-link> </td>
          <td> {{ item["Current Sessions"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-md-block">
    <table id="servers_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Server Address</th>
          <th>Current Sessions</th>
          <th>Uptime</th>
          <th>SDK Version</th>
          <th>Buyer</th>
          <th>Datacenter</th>
          <th>Datacenter Id</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Server Link"]'> {{ item["Server Address"] }} </router-link> </td>
          <td> {{ item["Current Sessions"] }} </td>
          <td> {{ item["Uptime"] }} </td>
          <td> {{ item["SDK Version"] }} </td>
          <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer"] }} </router-link> </td>
          <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
          <td> {{ item["Datacenter Id"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"
import BigNumber from "bignumber.js";

function parse_uint64(value) {
  const bignum = new BigNumber(value);
  var hex = bignum.toString(16);
  while (hex.length<16) {
    hex = '0' + hex
  }
  return hex
}

function nice_uptime(value) {
  if (isNaN(value)) {
    return ''
  }
  if (value > 86400) {
    return Math.floor(value/86400) + "d"
  }
  if (value > 3600) {
    return Math.floor(value/3600) + "h"
  }
  if (value > 60) {
    return Math.floor(value/60) + "m"
  }
  return value + "s"
}

async function getData() {
  try {
    const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/servers/0/100');
    let i = 0;
    let data = []
    while (i < res.data.servers.length) {
      let v = res.data.servers[i]
      const datacenterLink = v.datacenter_name != "" ? "datacenter/" + v.datacenter_name : ""
      let row = {
        "Server Address":v.server_address,
        "Server Link":"server/" + v.server_address,
        "Buyer":"Raspberry",
        "Buyer Link":"buyer/" + v.buyer_code,
        "Datacenter":v.datacenter_name,
        "Datacenter Id": parse_uint64(v.datacenter_id),
        "Datacenter Link":datacenterLink,
        "Current Sessions":v.num_sessions,
        "SDK Version":v.sdk_version_major + "." + v.sdk_version_minor + "." + v.sdk_version_patch,
        "Uptime": nice_uptime(v.uptime),
      }
      data.push(row)
      i++;
    }
    return data
  } catch (error) {
    console.log(error);
    return null
  }
}

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      data: [],
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
