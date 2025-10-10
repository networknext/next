// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Server</p>
      <p class="tight-p test-text"><input id='server-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-if="this.found" class="bottom-server">

      <div id="left" class="left">
      
        <div class="d-md-none">
  
          <table id="session_table" class="table table-striped" style="text-align: center; table-layout:auto;" >
            <tbody>

              <tr>
                <td style="width: 50%" class="bold">Sessions</td>
                <td> {{ this.data.session_count }} </td>
              </tr>

              <tr>
                <td class="bold">Datacenter</td>
                <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Buyer</td>
                <td> <router-link :to="'/buyer/' + this.data['buyer_code']"> {{ this.data['buyer_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Uptime</td>
                <td> {{ this.data.uptime }} </td>
              </tr>

            </tbody>
          </table>
  
        </div>

        <div class="d-md-none">
          <table id="sessions_table" class="table table-striped" style="table-layout:auto;">
            <thead>
              <tr>
                <th style="width: 50%">Session ID</th>
                <th style="width: 50%">Improvement</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in data.server_sessions" :key='item'>
                <td class="fixed"> <router-link :to='"/session/" + item.session_id'> {{ item.session_id }} </router-link> </td>
                <td class="green-center" v-if="item.improvement != '--' && item.improvement >= 10"> {{ item.improvement }} ms</td>
                <td class="orange-center" v-else-if="item.improvement != '--' && item.improvement >= 5"> {{ item.improvement }} ms</td>
                <td class="red-center" v-else-if="item.improvement != '--' && item.improvement > 0"> {{ item.improvement }} ms</td>
                <td class="nada-center" v-else> -- </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="d-none d-md-block d-lg-block d-xl-block d-xxl-none">
          <table id="sessions_table" class="table table-striped table-hover" style="table-layout:auto;">
            <thead>
              <tr>
                <th>Session ID</th>
                <th>ISP</th>
                <th>Country</th>
                <th>Latency</th>
                <th>Improvement</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in data.server_sessions" :key='item'>
                <td class="fixed"> <router-link :to='"/session/" + item.session_id'> {{ item.session_id }} </router-link> </td>
                <td>{{item.country}}</td>
                <td>{{item.isp}}</td>
                <td>{{item.latency}} ms</td>
                <td class="green-center" v-if="item.improvement != '--' && item.improvement >= 10"> {{ item.improvement }} ms</td>
                <td class="orange-center" v-else-if="item.improvement != '--' && item.improvement >= 5"> {{ item.improvement }} ms</td>
                <td class="red-center" v-else-if="item.improvement != '--' && item.improvement > 0"> {{ item.improvement }} ms</td>
                <td class="nada-center" v-else> -- </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="d-none d-xxl-block">
          <table id="sessions_table" class="table table-striped table-hover" style="table-layout:auto;">
            <thead>
              <tr>
                <th>Session ID</th>
                <th>Country</th>
                <th>Connection</th>
                <th>Platform</th>
                <th>ISP</th>
                <th>Latency</th>
                <th>Improvement</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in data.server_sessions" :key='item'>
                <td class="fixed"> <router-link :to='"/session/" + item.session_id'> {{ item.session_id }} </router-link> </td>
                <td>{{item.country}}</td>
                <td>{{item.connection}}</td>
                <td>{{item.platform}}</td>
                <td>{{item.isp}}</td>
                <td>{{item.latency}} ms</td>
                <td class="green-center" v-if="item.improvement != '--' && item.improvement >= 10"> {{ item.improvement }} ms</td>
                <td class="orange-center" v-else-if="item.improvement != '--' && item.improvement >= 5"> {{ item.improvement }} ms</td>
                <td class="red-center" v-else-if="item.improvement != '--' && item.improvement > 0"> {{ item.improvement }} ms</td>
                <td class="nada-center" v-else> -- </td>
              </tr>
            </tbody>
          </table>
        </div>

      </div>

      <div id="right" class="right d-none d-xxl-block">

        <div class="server_info">

          <table id="session_table" class="table table-striped" style="text-align: center;">
            <tbody>

              <tr>
                <td class="bold">Sessions</td>
                <td> {{ this.data.session_count }} </td>
              </tr>

              <tr>
                <td class="bold">Datacenter</td>
                <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Buyer</td>
                <td> <router-link :to="'/buyer/' + this.data['buyer_code']"> {{ this.data['buyer_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Uptime</td>
                <td> {{ this.data.uptime }} </td>
              </tr>

            </tbody>
          </table>

        </div>

      </div>

    </div>

  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from '@/update.js'

import {parse_uint64, nice_uptime, getPlatformName, getConnectionName, getCountryName} from '@/utils.js'

async function getData(page, server) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/server/' + server + '/' + page
    const res = await axios.get(url);
    let i = 0
    let data = {}
    data.server_sessions = []
    if (res.data.server_sessions != null) {
      while (i < res.data.server_sessions.length) {
        const v = res.data.server_sessions[i]
        const session_id = parse_uint64(v.session_id)
        const improvement = v.next_rtt != 0 && v.next_rtt < v.direct_rtt ? v.direct_rtt - v.next_rtt : "--"
        const connection = getConnectionName(v.connection_type)
        const platform = getPlatformName(v.platform_type)
        const country = getCountryName(v.country)
        const latency = (v.next_rtt != 0 && v.next_rtt < v.direct_rtt) ? v.next_rtt : v.direct_rtt
        let start_time = new Date(parseInt(v.start_time*1000)).toLocaleString()
        let row = {
          "session_id":session_id,
          "isp":v.isp,
          "buyer":v.buyer_name,
          "buyer_link":"/buyer/" + v.buyer_code,
          "datacenter":v.datacenter_name,
          "datacenter_link": "/datacenter/" + v.datacenter_name,
          "improvement":improvement,
          "connection":connection,
          "platform":platform,
          "country":country,
          "start_time":start_time,
          "latency":latency,
        }
        data.server_sessions.push(row)
        i++;
      }
    }
    data.found = true
    data.server = server
    const outputPage = res.data.output_page
    const numPages = res.data.num_pages
    data.datacenter_name = res.data.server_data.datacenter_name
    data.buyer_code = res.data.server_data.buyer_code
    data.buyer_name = res.data.server_data.buyer_name
    data.uptime = nice_uptime(res.data.server_data.uptime)
    data.session_count = res.data.server_data.num_sessions
    return [data, outputPage,numPages]
  } catch (error) {
    console.log(error);
    let data = {found: false, server: server}
    return [data, 0, 1]
  }
}

export default {

  name: "App",

  mixins: [update],

  mounted: function () {
    document.getElementById("server-input").value = document.getElementById("server-input").defaultValue = this.data['server']
    document.getElementById("server-input").addEventListener('keyup', this.onKeyUp);
    this.$emit('notify-view', 'server')
  },

  beforeUnmount() {
    document.getElementById("server-input").removeEventListener('keyup', this.onKeyUp);
  },

  data() {
    return {
      data: [],
      found: false,
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let server = ''
    let page = 0
    if (values.length > 0) {
      if (values[values.length-2] != 'server') {
        server = values[values.length-2]
        page = parseInt(values[values.length-1])
      } else {
        server = values[values.length-1]
      }
      if (isNaN(page)) {
        page = 0
      }
    }
    let result = await getData(page, server)
    next(vm => {
      if (result != null) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
        vm.$emit('notify-update', vm.page, vm.num_pages)
        vm.found = result[0]['found']
      }
    })
  },

  methods: {

    async getData(page, server) {
      if (server == null) {
        server = this.$route.params.id
      }
      return getData(page, server)
    },

    async update() {
      let result = await getData(this.page, this.$route.params.id)
      if (result != null) {
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
        this.found = result[0]['found']
      }
    },

    search() {
      const server = document.getElementById("server-input").value
      this.$router.push('/server/' + server)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

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

.parent {
  width: 100%;
  height: 100%;
  padding: 15px;
  padding-top: 20px;
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

.bottom-server {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: row;
  gap: 25px;
  padding-top: 10px;
}

.left {
  width: 100%;
  height: 100%;
  padding: 0px;
  display: flex;
  flex-direction: column;
  gap: 25px;
  margin-top: -10px;
}

.right {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 0px;
  max-width: 500px;
  min-width: 300px;
}

.right-top {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.right-bottom {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 50px;
}

.map {
  background-color: #555555;
  width: 100%;
  height: 500px;
  flex-shrink: 0;
}

.server_info {
  width: 100%;
  padding-top: 15px;
}

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
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

</style>

// -----------------------------------------------------------------------------------------
