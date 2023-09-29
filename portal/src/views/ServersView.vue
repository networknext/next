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
        </tr>
      </tbody>
    </table>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
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

async function getData(page) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/servers/' + page
    const res = await axios.get(url);
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
        "Datacenter Link":datacenterLink,
        "Current Sessions":v.num_sessions.toLocaleString(),
        "SDK Version":v.sdk_version_major + "." + v.sdk_version_minor + "." + v.sdk_version_patch,
        "Uptime": nice_uptime(v.uptime),
      }
      data.push(row)
      i++;
    }
    const outputPage = res.data.output_page
    const numPages = res.data.num_pages
    return [data, outputPage, numPages]
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
    let values = to.path.split("/")
    let page = 0
    if (values.length > 0) {
      let value = values[values.length-1]
      page = parseInt(value)
      if (isNaN(page)) {
        page = 0
      }
    }
    let result = await getData(page)
    next(vm => {
      if (result != null) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
        vm.$emit('update', vm.page, vm.num_pages)
      }
    })
  },

  methods: {

    async getData(page) {
      return getData(page)
    },

    async update() {
      let result = await getData(this.page)
      if (result != null) {
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
      }
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
