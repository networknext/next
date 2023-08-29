// -----------------------------------------------------------------------------------------

<template>

  <div class="d-xl-none">
    <table v-if="this.updated" id="buyer_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Buyer Name</th>
          <th>Live</th>
          <th>Debug</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer Name"] }} </router-link> </td>
          <td> {{ item["Live"] }} </td>
          <td> {{ item["Debug"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="d-none d-xl-block">
    <table v-if="this.updated" id="buyer_table" class="table table-striped table-hover">
      <thead>
        <tr>
          <th>Buyer Name</th>
          <th>Buyer Code</th>
          <th>Live</th>
          <th>Debug</th>
          <th>Public Key</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in data" :key='item'>
          <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer Name"] }} </router-link> </td>
          <td> <router-link :to='item["Buyer Link"]'> {{ item["Buyer Code"] }} </router-link> </td>
          <td> {{ item["Live"] }} </td>
          <td> {{ item["Debug"] }} </td>
          <td> {{ item["Public Key"] }} </td>
        </tr>
      </tbody>
    </table>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      data: [],
    };
  },

  methods: {

    async update() {
      try {
        // todo: this should become the portal version of this query
        const res = await axios.get(process.env.VUE_APP_API_URL + '/database/buyers');
        let i = 0;
        let data = []
        while (i < res.data.buyers.length) {
          let v = res.data.buyers[i]        
          let row = {
            "Buyer Name":v.name, 
            "Buyer Code":v.code, 
            "Buyer Link":"buyer/" + v.code,
            "Live":v.live, 
            "Debug":v.debug, 
            "Public Key":v.public_key
          }
          data.push(row)
          i++;
        }
        data.sort( function(a,b) {
        if (a["Buyer"] < b["Buyer"]) {
          return -1
        }
        if (a["Buyer Name"] > b["Buyer Name"]) {
          return +1
        }
        return 0
        })
        this.data = data
        this.updated = true
      } catch (error) {
        console.log(error);
      }
    }

  }

};

</script>

// -----------------------------------------------------------------------------------------

<style scoped>

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
