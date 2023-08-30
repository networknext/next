// -----------------------------------------------------------------------------------------

<template>

  <div class="btn-group">
    <button type="button" class="btn btn-secondary dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">
      All
    </button>
    <ul class="dropdown-menu">
      <li><a class="dropdown-item active" aria-current="true" href="#">All</a></li>
      <li><a class="dropdown-item" href="#">Raspberry</a></li>
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
      fields: ["Buyer Name", "Live", "Debug", "Public Key"],
      data: []
    };
  },

  async created() {
    try {
      // todo: update to a portal specific method, ideally something that returns just the buyer names and codes (if needed)
      const res = await axios.get('http://dev.virtualgo.net/database/buyers');
      let i = 0;
      while (i < res.data.buyers.length) {
        let v = res.data.buyers[i]        
        let row = {"Buyer Name":v.name, "Live":v.live, "Debug":v.debug, "Public Key":v.public_key}
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
