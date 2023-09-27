// -----------------------------------------------------------------------------------------

<template>

  <div class="d-md-none">
    <table id="sessions_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Session ID</th>
          <th>Improvement</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td class="fixed"> <router-link :to='"session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
          <td class="green-center" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
          <td class="orange-center" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
          <td class="red-center" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
          <td class="nada-center" v-else> -- </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-md-block d-lg-block d-xl-none">
    <table id="sessions_table" class="table table-striped table-hover">
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
          <td class="right"> {{ item["Direct RTT"] }} ms</td>
          <td class="right"> {{ item["Next RTT"] }} ms</td>
          <td class="green" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
          <td class="orange" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
          <td class="red" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
          <td class="nada" v-else> -- </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-xl-block d-xxl-none">
    <table id="sessions_table" class="table table-striped table-hover">
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
          <td class="right"> {{ item["Direct RTT"] }} ms</td>
          <td class="right"> {{ item["Next RTT"] }} ms</td>
          <td class="green" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
          <td class="orange" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
          <td class="red" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
          <td class="nada" v-else> -- </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-xxl-block">
    <table id="sessions_table" class="table table-striped table-hover">
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
          <td class="right"> {{ item["Direct RTT"] }} ms</td>
          <td class="right"> {{ item["Next RTT"] }} ms</td>
          <td class="green" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
          <td class="orange" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
          <td class="red" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
          <td class="nada" v-else> -- </td>
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

async function getData() {
  try {
    const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/sessions/0/100');
    let i = 0;
    let data = []
    while (i < res.data.sessions.length) {
      const v = res.data.sessions[i]
      const session_id = parse_uint64(v.session_id)
      const user_hash = parse_uint64(v.user_hash)
      const next_rtt = v.next_rtt > 0.0 ? v.next_rtt : ""
      const improvement = v.next_rtt < v.direct_rtt ? v.direct_rtt - v.next_rtt : "--"
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
        "Next RTT":next_rtt,
        "Improvement":improvement,
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

  mixins: [utils,update],

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

.fixed {
  font-family: monospace;
}

.right {
  text-align: right;
}

.green {
  color: #11AA44;
  font-weight: bold;
  text-align: right;
}

.orange {
  color: #F38701;
  font-weight: bold;
  text-align: right;
}

.red {
  color: #E34234;
  font-weight: bold;
  text-align: right;
}

.nada {
  color: #D3D3D3;
  font-weight: bold;
  text-align: right;
}

.green-center {
  color: #11AA44;
  font-weight: bold;
}

.orange-center {
  color: #F38701;
  font-weight: bold;
}

.red-center {
  color: #E34234;
  font-weight: bold;
}

.nada-center {
  color: #D3D3D3;
  font-weight: bold;
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
