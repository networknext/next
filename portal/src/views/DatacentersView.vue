// -----------------------------------------------------------------------------------------

<template>

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

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios"
import update from "@/update.js"

async function getData() {
  try {
    const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/datacenters');
    res.data.datacenters.sort(function (a, b) {
      if (a.name < b.name) {
        return -1;
      }
      if (a.name > b.name) {
        return 1;
      }
      return 0;
    });
    let i = 0
    let data = []
    while (i < res.data.datacenters.length) {
      let v = res.data.datacenters[i]
      let row = {
        "Datacenter Name":v.name,
        "Datacenter Link":"datacenter/" + v.name,
        "Latitude":v.latitude,
        "Longitude":v.longitude,
        "Seller":v.seller_name,
        "Seller Link":"seller/" + v.seller_code,
        "Native Name":v.native
      }
      data.push(row)
      i++
    }
    return data
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

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

</style>

// -----------------------------------------------------------------------------------------
