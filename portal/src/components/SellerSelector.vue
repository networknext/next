// -----------------------------------------------------------------------------------------

<template>

  <div class="btn-group">
    <button type="button" class="btn btn-secondary dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">
      All
    </button>
    <ul class="dropdown-menu">
      <li><a class="dropdown-item active" aria-current="true" href="#">All</a></li>
      <li><a class="dropdown-item" href="#">Google</a></li>
      <li><a class="dropdown-item" href="#">Amazon</a></li>
      <li><a class="dropdown-item" href="#">Akamai</a></li>
      <li><a class="dropdown-item" href="#">Vultr</a></li>
    </ul>
  </div>

</template>

// -----------------------------------------------------------------------------------------

<script>

import axios from "axios";

export default {

  name: "App",

  data() {
    return {
      fields: ["Seller Name"],
      data: []
    };
  },

  async created() {
    try {
      const res = await axios.get('http://dev.virtualgo.net/database/sellers');
      let i = 0;
      while (i < res.data.sellers.length) {
        let v = res.data.sellers[i]        
        let row = {"Seller Name":v.name}
        this.data.push(row)
        i++;
      }
    } catch (error) {
      console.log(error);
    }
  },

};

</script>

// -----------------------------------------------------------------------------------------

<style scoped>

button {
  font-weight: bold;
  font-size: 20px;
}

</style>

// -----------------------------------------------------------------------------------------
