// -----------------------------------------------------------------------------------------

<template>

  <div class="d-md-none">
    <table v-if="this.updated" id="servers_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Server Address</th>
          <th>Datacenter</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Server Link"]'> {{ item["Server Address"] }} </router-link> </td>
          <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-md-block">
    <table v-if="this.updated" id="servers_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Server Address</th>
          <th>Buyer</th>
          <th>Datacenter</th>
          <th>Current Sessions</th>
          <th>Uptime</th>
          <th>SDK Version</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Server Link"]'> {{ item["Server Address"] }} </router-link> </td>
          <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer"] }} </router-link> </td>
          <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
          <td> {{ item["Current Sessions"] }} </td>
          <td> {{ item["Uptime"] }} </td>
          <td> {{ item["SDK Version"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"
import utils from "@/utils.js"

export default {

  name: "App",

  mixins: [update, utils],

  data() {
    return {
      data: []
    };
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/servers/0/10');
        let i = 0;
        let data = []
        while (i < res.data.servers.length) {
          let v = res.data.servers[i]        
          let row = {
            "Server Address":v.server_address, 
            "Server Link":"server/" + v.server_address,
            "Buyer":"Raspberry", 
            "Buyer Link":"buyer/" + v.buyer_code,
            "Datacenter":v.datacenter_name, 
            "Datacenter Link":"datacenter/" + v.datacenter_name,
            "Current Sessions":v.num_sessions, 
            "SDK Version":v.sdk_version_major + "." + v.sdk_version_minor + "." + v.sdk_version_patch,
            "Uptime": this.nice_uptime(v.uptime),
          }
          data.push(row)
          i++;
        }
        this.data = data
        this.updated = true
      } catch (error) {
        console.log(error);
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
