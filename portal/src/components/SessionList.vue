// -----------------------------------------------------------------------------------------

<template>

  <div class="d-md-none">
    <table v-if="this.updated" id="sessions_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Session ID</th>
          <th class="right">Direct</th>
          <th class="right">Next</th>
          <th class="right">Improvement</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td class="fixed"> <router-link :to='"session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
          <td class="right"> {{ item["Direct RTT"] }} </td>
          <td class="right"> {{ item["Next RTT"] }} </td>
          <td class="right"> {{ item["Improvement"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-md-block d-lg-block d-xl-none">
    <table v-if="this.updated" id="sessions_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Session ID</th>
          <th>ISP</th>
          <th class="right">Direct RTT</th>
          <th class="right">Next RTT</th>
          <th class="right">Improvement</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td class="fixed"> <router-link :to='"session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
          <td> {{ item["ISP"] }} </td>
          <td class="right"> {{ item["Direct RTT"] }} </td>
          <td class="right"> {{ item["Next RTT"] }} </td>
          <td class="right"> {{ item["Improvement"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-xl-block d-xxl-none">
    <table v-if="this.updated" id="sessions_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Session ID</th>
          <th>User Hash</th>
          <th>ISP</th>
          <th class="right">Direct RTT</th>
          <th class="right">Next RTT</th>
          <th class="right">Improvement</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td class="fixed"> <router-link :to='"session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
          <td class="fixed"> <router-link :to='"user/" + item["User Hash"]'> {{ item["User Hash"] }} </router-link> </td>
          <td> {{ item["ISP"] }} </td>
          <td class="right"> {{ item["Direct RTT"] }} </td>
          <td class="right"> {{ item["Next RTT"] }} </td>
          <td class="right"> {{ item["Improvement"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-xxl-block">
    <table v-if="this.updated" id="sessions_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Session ID</th>
          <th>User Hash</th>
          <th>ISP</th>
          <th>Buyer</th>
          <th>Datacenter</th>
          <th>Server Address</th>
          <th class="right">Direct RTT</th>
          <th class="right">Next RTT</th>
          <th class="right">Improvement</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td class="fixed"> <router-link :to='"session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
          <td class="fixed"> <router-link :to='"user/" + item["User Hash"]'> {{ item["User Hash"] }} </router-link> </td>
          <td> {{ item["ISP"] }} </td>
          <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer"] }} </router-link> </td>
          <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
          <td> <router-link :to='"server/" + item["Server Address"]'> {{ item["Server Address"] }} </router-link> </td>
          <td class="right"> {{ item["Direct RTT"] }} </td>
          <td class="right"> {{ item["Next RTT"] }} </td>
          <td class="right"> {{ item["Improvement"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import utils from '@/utils.js'
import update from '@/update.js'

export default {

  name: "App",

  mixins: [utils,update],

  data() {
    return {
      data: [],
    };
  },

  methods: {

    async update() {
      try {
        const res = await axios.get('http://dev.virtualgo.net/portal/sessions/0/1000');
        let i = 0;
        let data = []
        if (res.data.sessions !== null ) {
          while (i < res.data.sessions.length) {
            const v = res.data.sessions[i]
            const session_id = this.parse_uint64(v.session_id)
            const user_hash = this.parse_uint64(v.user_hash)
            let row = {
              "Session ID":session_id, 
              "User Hash":user_hash,
              "ISP":v.isp, 
              "Buyer":v.buyer_name, 
              "Buyer Link":"buyer/" + v.buyer_code, 
              "Datacenter":v.datacenter_name, 
              "Datacenter Link": "datacenter/" + v.datacenter_name,
              "Server Address":v.server_address, 
              "Direct RTT":v.direct_rtt, 
              "Next RTT":v.next_rtt, 
              "Improvement":"--"
            }
            data.push(row)
            i++;
          }
        }
        this.data = data
      } catch (error) {
        console.log(error);
      }
    }

  }
};

</script>

// -----------------------------------------------------------------------------------------

<style scoped>

.fixed {
  font-family: monospace;
}  

.right {
  text-align: right;
}

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

</style>

// -----------------------------------------------------------------------------------------
