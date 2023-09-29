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

  <div class="d-none d-xxl-block">
    <table v-if="this.updated" id="sessions_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Session ID</th>
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
import BigNumber from "bignumber.js";

function parse_uint64(value) {
  const bignum = new BigNumber(value);
  var hex = bignum.toString(16);
  while (hex.length<16) {
    hex = '0' + hex
  }
  return hex
}

async function getData(page, user_hash) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/user_sessions/' + user_hash + '/' + page
    console.log(url)
    const res = await axios.get(url);
    let i = 0
    let data = []
    while (i < res.data.sessions.length) {
      const v = res.data.sessions[i]
      const session_id = parse_uint64(v.session_id)
      const next_rtt = v.next_rtt > 0.0 ? v.next_rtt + " ms" : ""
      const improvement = v.next_rtt != 0 && v.next_rtt < v.direct_rtt ? v.direct_rtt - v.next_rtt : "--"
      let row = {
        "Session ID":session_id,
        "ISP":v.isp,
        "Buyer":v.buyer_name,
        "Buyer Link":"buyer/" + v.buyer_code,
        "Datacenter":v.datacenter_name,
        "Datacenter Link": "datacenter/" + v.datacenter_name,
        "Server Address":v.server_address,
        "Direct RTT":v.direct_rtt + " ms",
        "Next RTT":next_rtt,
        "Improvement":improvement,
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

  mixins: [utils,update],

  data() {
    return {
      data: [],
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    console.log(values)
    let user_hash = 0
    let page = 0
    if (values.length > 0) {
      if (values[values.length-2] != 'user') {
        user_hash = values[values.length-2]
        page = parseInt(values[values.length-1])
      } else {
        user_hash = values[values.length-1]
      }
      if (isNaN(page)) {
        page = 0
      }
      console.log("user hash is " + user_hash)
      console.log("page is " + page)
    }
    let result = await getData(page, user_hash)
    next(vm => {
      if (result != null) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
      }
    })
  },

  methods: {

    async getData(page, user_hash) {
      if (user_hash == null) {
        user_hash = this.$route.params.id
      }
      return getData(page, user_hash)
    },

    async update() {
      let result = await getData(this.page, this.$route.params.id)
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
