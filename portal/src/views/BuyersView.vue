// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">

    <div class="d-xl-none">
      <table id="buyer_table" class="table table-striped">
        <thead>
          <tr>
            <th>Buyer Name</th>
            <th>Accelerated Sessions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer Name"] }} </router-link> </td>
            <td> {{ item["Next Sessions"] }} / {{ item["Total Sessions"] }} ({{ item["Accelerated Percent"] }}%) </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="d-none d-xl-block">
      <table id="buyer_table" class="table table-striped">
        <thead>
          <tr>
            <th>Buyer Name</th>
            <th>Buyer Code</th>
            <th>Live</th>
            <th>Debug</th>
            <th>Public Key</th>
            <th>Server Count</th>
            <th>Accelerated Sessions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer Name"] }} </router-link> </td>
            <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer Code"] }} </router-link> </td>
            <td> {{ item["Live"] }} </td>
            <td> {{ item["Debug"] }} </td>
            <td> {{ item["Public Key"] }} </td>
            <td> {{ item["Server Count"] }} </td>
            <td> {{ item["Next Sessions"] }} / {{ item["Total Sessions"] }} ({{ item["Accelerated Percent"] }}%) </td>
          </tr>
        </tbody>
      </table>
    </div>

  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

import { getAcceleratedPercent } from '@/utils.js'

async function getData(page) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/buyers/' + page
    const res = await axios.get(url);
    let i = 0;
    let data = []
    while (i < res.data.buyers.length) {
      let v = res.data.buyers[i]
      let row = {
        "Buyer Name":v.name,
        "Buyer Code":v.code,
        "Buyer Link":"/buyer/" + v.code,
        "Live":v.live,
        "Debug":v.debug,
        "Public Key":v.public_key,
        "Total Sessions":v.total_sessions,
        "Next Sessions":v.next_sessions,
        "Accelerated Percent":getAcceleratedPercent(v.next_sessions, v.total_sessions),
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
    this.$emit('notify-view', 'buyers')
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

tbody>tr>:nth-child(4){
  font-family: monospace;
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
