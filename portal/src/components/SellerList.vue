// -----------------------------------------------------------------------------------------

<template>

  <table v-if="this.updated" id="seller_table" class="table table-striped table-hover">
    <thead>
      <tr>
        <th v-for="field in fields" :key='field'> {{field}} </th>
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

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      fields: ["Seller Name", "Seller Code"],
      data: [],
    };
  },

  methods: {

    async update() {
      try {
        // todo: this should be updated to the portal version of this API
        const res = await axios.get(process.env.VUE_APP_API_URL + '/database/sellers');
        this.items = []
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
        this.data = data.sort()
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

a {
  color: #2c3e50;
  text-decoration: none;
}

tr {
    white-space: nowrap;
}

</style>

// -----------------------------------------------------------------------------------------
