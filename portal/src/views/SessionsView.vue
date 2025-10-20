// -----------------------------------------------------------------------------------------

<template>

  <div class="parent">
  
    <div class="d-md-none">
      <table id="sessions_table" class="table table-striped">
        <thead>
          <tr>
            <th>Session ID</th>
            <th>Improvement</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td class="fixed"> <router-link :to='"/session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
            <td class="green-center" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
            <td class="orange-center" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
            <td class="red-center" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
            <td class="nada-center" v-else> -- </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="d-none d-md-block d-lg-none">
      <table id="sessions_table" class="table table-striped table-hover">
        <thead>
          <tr>
            <th>Session ID</th>
            <th>Platform</th>
            <th>Connection</th>
            <th>Country</th>
            <th class="right_align">Improvement</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td class="fixed"> <router-link :to='"/session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
            <td> {{ item["Platform"] }} </td>
            <td> {{ item["Connection"] }} </td>
            <td> {{ item["Country"] }} </td>
            <td class="green" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
            <td class="orange" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
            <td class="red" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
            <td class="nada" v-else> -- </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="d-none d-lg-block d-xl-none d-xxl-none">
      <table id="sessions_table" class="table table-striped table-hover">
        <thead>
          <tr>
            <th>Session ID</th>
            <th>Platform</th>
            <th>Connection</th>
            <th>Country</th>
            <th>Datacenter</th>
            <th class="right_align">Improvement</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td class="fixed"> <router-link :to='"/session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
            <td> {{ item["Platform"] }} </td>
            <td> {{ item["Connection"] }} </td>
            <td> {{ item["Country"] }} </td>
            <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
            <td class="green" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
            <td class="orange" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
            <td class="red" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
            <td class="nada" v-else> -- </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="d-none d-xl-block d-xxl-block">
      <table id="sessions_table" class="table table-striped table-hover">
        <thead>
          <tr>
            <th>Session ID</th>
            <th>Platform</th>
            <th>Connection</th>
            <th>Country</th>
            <th>ISP</th>
            <th>Datacenter</th>
            <th class="right_align">Improvement</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td class="fixed"> <router-link :to='"/session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
            <td> {{ item["Platform"] }} </td>
            <td> {{ item["Connection"] }} </td>
            <td> {{ item["Country"] }} </td>
            <td> {{ item["ISP"] }} </td>
            <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
            <td class="green" v-if="item['Improvement'] != '--' && item['Improvement'] >= 10"> {{ item["Improvement"] }} ms</td>
            <td class="orange" v-else-if="item['Improvement'] != '--' && item['Improvement'] >= 5"> {{ item["Improvement"] }} ms</td>
            <td class="red" v-else-if="item['Improvement'] != '--' && item['Improvement'] > 0"> {{ item["Improvement"] }} ms</td>
            <td class="nada" v-else> -- </td>
          </tr>
        </tbody>
      </table>
    </div>

  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from '@/update.js'

import {parse_uint64, getPlatformName, getConnectionName, getCountryName} from '@/utils.js'

async function getData(page) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/sessions/' + page
    const res = await axios.get(url);
    let i = 0
    let data = []
    while (i < res.data.sessions.length) {
      const v = res.data.sessions[i]
      const session_id = parse_uint64(v.session_id)
      const connection = getConnectionName(v.connection_type)
      const platform = getPlatformName(v.platform_type)
      const country = getCountryName(v.country)
      const latency = (v.next_rtt != 0 && v.next_rtt < v.direct_rtt) ? v.next_rtt : v.direct_rtt
      const improvement = ( v.next_rtt != 0 && v.next_rtt < v.direct_rtt ) ? ( v.direct_rtt - v.next_rtt ) : "--"
      let row = {
        "Session ID":session_id,
        "Country":country,
        "ISP":v.isp,
        "Connection":connection,
        "Platform":platform,
        "Buyer":v.buyer_name,
        "Buyer Link":"/buyer/" + v.buyer_code,
        "Datacenter":v.datacenter_name,
        "Datacenter Link": "/datacenter/" + v.datacenter_name,
        "Latency":latency,
        "Improvement":improvement,
      }
      data.push(row)
      i++;
    }
    const outputPage = res.data.output_page
    const numPages = res.data.num_pages
    return [data, outputPage,numPages]
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
        vm.$emit('notify-update', vm.page, vm.num_pages)
      }
    })
  },

  mounted: function() {
    this.$emit('notify-view', 'sessions')
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

.parent {
  width: 100%;
  height: 100%;
  padding-top: 10px;
}

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

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
}

</style>

// -----------------------------------------------------------------------------------------
