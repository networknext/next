// -----------------------------------------------------------------------------------------

<template>

  <table id="seller_table" class="table table-striped table-hover">
    <thead>
      <tr>
        <th>Seller Name</th>
        <th>Seller Code</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="item in data" :key='item'>
        <td> <router-link :to='item["Seller Link"]'> {{ item["Seller Name"] }} </router-link> </td>
        <td> <router-link :to='item["Seller Link"]'> {{ item["Seller Code"] }} </router-link> </td>
      </tr>
    </tbody>
  </table>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

async function getData(page) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/sellers/' + page
    const res = await axios.get(url);
    let i = 0;
    let data = []
    while (i < res.data.sellers.length) {
      let v = res.data.sellers[i]
      let row = {
        "Seller Name":v.name,
        "Seller Link":"seller/" + v.code,
        "Seller Code":v.code,
      }
      data.push(row)
      i++;
    }
    data.sort( function(a,b) {
      if (a["Seller Name"] < b["Seller Name"]) {
        return -1
      }
      if (a["Seller Name"] > b["Seller Name"]) {
        return +1
      }
      return 0
    })
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
      }
    })
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
