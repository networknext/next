// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">

    <div class="d-md-none">
      <table id="relays_table" class="table table-striped">
        <thead>
          <tr>
            <th>Relay Name</th>
            <th>Sessions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
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
          <tr v-for="item in data" :key='item'>
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
            <th>Datacenter</th>
            <th>Seller</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td> <router-link :to='item["Relay Link"]'> {{ item["Relay Name"] }} </router-link> </td>
            <td> {{ item["Current Sessions"] }} </td>
            <td> {{ item["Status"] }} </td>
            <td> {{ item["Uptime"] }} </td>
            <td> {{ item["Relay Version"] }} </td>
            <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter"] }} </router-link> </td>
            <td> <router-link :to='item["Seller Link"]'> {{ item["Seller"] }} </router-link> </td>
          </tr>
        </tbody>
      </table>
    </div>

  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios"
import update from "@/update.js"

import {nice_uptime} from '@/utils.js'

async function getData(page) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/relays/' + page
    const res = await axios.get(url);
    let i = 0
    let data = []
    while (i < res.data.relays.length) {
      const v = res.data.relays[i]
      var status = "Online"
      if (v.relay_flags & 1) {
        status = "Shutting Down"
      }
      let row = {
        "Relay Name":v.relay_name,
        "Relay Link":"/relay/" + v.relay_name,
        "Datacenter":v.datacenter_name,
        "Datacenter Link":"/datacenter/" + v.datacenter_name,
        "Seller":v.seller_name,
        "Seller Link":"/seller/" + v.seller_code,
        "Current Sessions":v.num_sessions,
        "Status":status,
        "Uptime":nice_uptime(v.uptime),
        "Relay Version":v.relay_version
      }
      data.push(row)
      i++
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
    this.$emit('notify-view', 'relays')
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

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

</style>

// -----------------------------------------------------------------------------------------
