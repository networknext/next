// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Datacenter</p>
      <p class="tight-p test-text"><input id='datacenter-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-if="this.found" class="bottom">

      <div v-if="this.data.relays.length > 0" class="sessions">

        <div class="d-md-none">
          <table id="relays_table" class="table table-striped table-hover">
            <thead>
              <tr>
                <th>Relay Name</th>
                <th>Sessions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in this.data.relays" :key='item'>
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
              <tr v-for="item in this.data.relays" :key='item'>
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
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in this.data.relays" :key='item'>
                <td> <router-link :to='item["Relay Link"]'> {{ item["Relay Name"] }} </router-link> </td>
                <td> {{ item["Current Sessions"] }} </td>
                <td> {{ item["Status"] }} </td>
                <td> {{ item["Uptime"] }} </td>
                <td> {{ item["Relay Version"] }} </td>
                <td> {{ item["Public Address"] }} </td>
                <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
              </tr>
            </tbody>
          </table>
        </div>

      </div>

      <div v-else>

        <p class="no-relays">There are no relays in {{this.data.datacenter}}</p>

      </div>

    </div>

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

async function getData(page, datacenter) {
  try {
    if (page == null) {
      page = 0
    }

    let data = {}
    data.relays = []
    data.datacenter = datacenter
    const url = process.env.VUE_APP_API_URL + '/portal/datacenter/' + datacenter + '/' + page
    const res = await axios.get(url);
    let i = 0
    while (i < res.data.relays.length) {
      const v = res.data.relays[i]
      var status = "Online"
      if (v.relay_flags & 1) {
        status = "Shutting Down"
      }
      let row = {
        "Relay Name":v.relay_name,
        "Relay Link":"/relay/" + v.relay_name,
        "Public Address":v.relay_address,
        "Datacenter":v.datacenter_name,
        "Datacenter Link":"/datacenter/" + v.datacenter_name,
        "Seller":v.seller_name,
        "Seller Link":"/seller/" + v.seller_code,
        "Current Sessions":v.num_sessions,
        "Status":status,
        "Uptime":nice_uptime(v.uptime),
        "Relay Version":v.relay_version
      }
      data.relays.push(row)
      i++
    }
    const outputPage = res.data.output_page
    const numPages = res.data.num_pages
    data.found = true
    return [data, outputPage,numPages]
  } catch (error) {
    console.log(error);
    let data = {}
    data.datacenter = datacenter
    data.found = false
    return [data, 0, 1]
  }
}

export default {

  name: "App",

  mixins: [update, utils],

  data() {
    return {
      data: [],
      found: false,
    };
  },

  mounted: function () {
    console.log('mounted')
    document.getElementById("datacenter-input").value = document.getElementById("datacenter-input").defaultValue = this.data['datacenter']
    document.getElementById("datacenter-input").addEventListener('keyup', this.onKeyUp);
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let datacenter = 0
    let page = 0
    if (values.length > 0) {
      if (values[values.length-2] != 'datacenter') {
        datacenter = values[values.length-2]
        page = parseInt(values[values.length-1])
      } else {
        datacenter = values[values.length-1]
      }
      if (isNaN(page)) {
        page = 0
      }
    }
    let result = await getData(page, datacenter)
    next(vm => {
      if (result != null) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
        vm.$emit('update', vm.page, vm.num_pages)
        vm.found = result[0]['found']
        console.log('found = ' + vm.found)
      }
    })
  },

  methods: {

    search() {
      const datacenter = document.getElementById("datacenter-input").value
      this.$router.push('/datacenter/' + datacenter)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

    async getData(page, datacenter) {
      if (datacenter == null) {
        datacenter = this.$route.params.id
      }
      return getData(page, datacenter)
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

.parent {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 15px;
}

.bottom {
  width: 100%;
  height: 100%;  
  padding: 0px;
}

.search {
  width: 100%;
  height: 35px;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
  font-weight: 1;
  font-size: 18px;
  padding: 0px;
}

.text {
  width: 100%;
  height: 35px;
  font-size: 15px;
  padding: 5px;
}

.test-text {
  width: 10px;
  flex-grow: 1;
}

.no-relays {
  font-size: 18px;
  padding: 0px;
}

</style>

// -----------------------------------------------------------------------------------------
