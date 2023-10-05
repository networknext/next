// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Seller</p>
      <p class="tight-p test-text"><input id='seller-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-if="this.found" class="bottom-seller">

      <div class="d-md-none">
        <table id="relays_table" class="table table-striped">
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

  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios"
import update from "@/update.js"

import { nice_uptime } from '@/utils.js'

async function getData(page, seller) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/seller/' + seller + '/' + page
    const res = await axios.get(url);
    let i = 0
    let data = {}
    data.relays = []
    data.seller = seller
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
    data.seller = seller
    data.found = false
    return [data, 0, 1]
  }
}

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      data: [],
      found: false,
    };
  },

  mounted: function () {
    document.getElementById("seller-input").value = document.getElementById("seller-input").defaultValue = this.data['seller']
    document.getElementById("seller-input").addEventListener('keyup', this.onKeyUp);
    this.$emit('notify-view', 'seller')
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let seller = 0
    let page = 0
    if (values.length > 0) {
      if (values[values.length-2] != 'seller') {
        seller = values[values.length-2]
        page = parseInt(values[values.length-1])
      } else {
        seller = values[values.length-1]
      }
      if (isNaN(page)) {
        page = 0
      }
    }
    let result = await getData(page, seller)
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

    search() {
      const seller = document.getElementById("seller-input").value
      this.$router.push('/seller/' + seller)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

    async getData(page, seller) {
      if (seller == null) {
        seller = this.$route.params.id
      }
      return getData(page, seller)
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

.parent {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 15px;
  padding-top: 20px;
}

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

.bottom-seller {
  width: 100%;
  height: 100%;  
  padding: 0px;
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

</style>

// -----------------------------------------------------------------------------------------
