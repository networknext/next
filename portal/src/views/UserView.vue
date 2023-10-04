// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">User</p>
      <p class="tight-p test-text"><input id='user-hash-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-if="this.found" class="bottom">

      <div class="d-xxl-none">
    
        <table id="sessions_table" class="table table-striped" style="table-layout:auto;">
          <thead>
            <tr>
              <th style="width: 15%">Start Time</th>
              <th style="width: 15%">Session ID</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in data.sessions" :key='item'>
              <td> {{ item["Start Time"] }} </td>
              <td class="fixed"> <router-link :to='"/session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="d-none d-xxl-block">
        <table id="sessions_table" class="table table-striped table-hover" style="table-layout:auto;">
          <thead>
            <tr>
              <th style="width: 15%">Start Time</th>
              <th style="width: 15%">Session ID</th>
              <th>ISP</th>
              <th>Connection</th>
              <th>Platform</th>
              <th>Datacenter</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in data.sessions" :key='item'>
              <td> {{ item["Start Time"] }} </td>
              <td class="fixed"> <router-link :to='"/session/" + item["Session ID"]'> {{ item["Session ID"] }} </router-link> </td>
              <td> {{ item["ISP"] }} </td>
              <td> {{ item["Connection"] }} </td>
              <td> {{ item["Platform"] }} </td>
              <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>

  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from '@/update.js'

import {parse_uint64, getPlatformName, getConnectionName} from '@/utils.js'

async function getData(page, user_hash) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/user_sessions/' + user_hash + '/' + page
    const res = await axios.get(url);
    let i = 0
    let data = {}
    data.sessions = []
    while (i < res.data.sessions.length) {
      const v = res.data.sessions[i]
      const session_id = parse_uint64(v.session_id)
      const connection = getConnectionName(v.connection_type)
      const platform = getPlatformName(v.platform_type)
      let start_time = new Date(parseInt(v.start_time)).toLocaleString()
      let row = {
        "Session ID": session_id,
        "Start Time": start_time,
        "ISP": v.isp,
        "Connection": connection,
        "Platform": platform,
        "Datacenter": v.datacenter_name,
        "Datacenter Link": "/datacenter/" + v.datacenter_name,
      }
      data.sessions.push(row)
      data.found = true
      data.user_hash = user_hash
      i++;
    }
    const outputPage = res.data.output_page
    const numPages = res.data.num_pages
    return [data, outputPage, numPages]
  } catch (error) {
    console.log(error);
    let data = {found: false, user_hash: user_hash}
    return [data, 0, 1]
  }
}

export default {

  name: "App",

  mixins: [update],

  mounted: function () {
    document.getElementById("user-hash-input").value = document.getElementById("user-hash-input").defaultValue = this.data['user_hash']
    document.getElementById("user-hash-input").addEventListener('keyup', this.onKeyUp);
    this.$emit('notify-view', 'user')
  },

  data() {
    return {
      data: [],
      found: false,
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
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
    }
    let result = await getData(page, user_hash)
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
        this.found = result[0]['found']
      }
    },

    search() {
      const user_hash = document.getElementById("user-hash-input").value
      this.$router.push('/user/' + user_hash)
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
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 15px;
  padding-top: 25px;
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

.bottom {
  padding: 0px;
}

.align-left {
  text-align: left;
}

</style>

// -----------------------------------------------------------------------------------------
