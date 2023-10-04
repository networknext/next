// -----------------------------------------------------------------------------------------

<template>

  <div class="parent" id="parent">
  
    <div class="d-xl-none">
      <table id="datacenter_table" class="table table-striped table-hover">
        <thead>
          <tr>
            <th>Datacenter Name</th>
            <th>Latitude</th>
            <th>Longitude</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter Name"] }} </router-link> </td>
            <td> {{ item["Latitude"] }} </td>
            <td> {{ item["Longitude"] }} </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="d-none d-xl-block">
      <table id="datacenter_table" class="table table-striped table-hover">
        <thead>
          <tr>
            <th>Datacenter Name</th>
            <th>Latitude</th>
            <th>Longitude</th>
            <th>Seller</th>
            <th>Native Name</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key='item'>
            <td> <router-link :to='item["Datacenter Link"]'> {{ item["Datacenter Name"] }} </router-link> </td>
            <td> {{ item["Latitude"] }} </td>
            <td> {{ item["Longitude"] }} </td>
            <td> <router-link :to='item["Seller Link"]'> {{ item["Seller"] }} </router-link> </td>
            <td> {{ item["Native Name"] }} </td>
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

async function getData(page) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/datacenters/' + page
    const res = await axios.get(url);
    let i = 0
    let data = []
    while (i < res.data.datacenters.length) {
      let v = res.data.datacenters[i]
      let row = {
        "Datacenter Name":v.name,
        "Datacenter Link":"/datacenter/" + v.name,
        "Latitude":v.latitude.toFixed(2),
        "Longitude":v.longitude.toFixed(2),
        "Seller":v.seller_name,
        "Seller Link":"/seller/" + v.seller_code,
        "Native Name":v.native
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
    this.$emit('notify-view', 'datacenters')
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

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

</style>

// -----------------------------------------------------------------------------------------
