// -----------------------------------------------------------------------------------------

<template>

  <table id="database_header_table" class="table">
    <thead>
      <tr>
        <th v-for="field in fields" :key='field'> {{field}} </th>
      </tr>
    </thead>
    <tbody>
        <tr v-for="item in data" :key='item'>
        <td v-for="field in fields" :key='field'>{{item[field]}}</td>
      </tr>
    </tbody>
  </table> 

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";

export default {

  name: "App",

  data() {
    return {
      fields: ["Creator", "Creation Time", "Relays", "Datacenters", "Buyers", "Sellers"],
      data: []
    };
  },

  async created() {
    try {
      const res = await axios.get(process.env.VUE_APP_API_URL + '/database/header');
      let row = {"Creator":res.data.creator, "Creation Time":res.data.creation_time, "Relays":res.data.num_relays, "Datacenters":res.data.num_datacenters, "Buyers":res.data.num_buyers, "Sellers":res.data.num_sellers}
      this.data.push(row)
    } catch (error) {
      console.log(error);
    }
  },

};

</script>

// -----------------------------------------------------------------------------------------

<style scoped>

</style>

// -----------------------------------------------------------------------------------------
