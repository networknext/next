// ------------------------------------------------------------------

<template>

  <button type="button" class="btn relay">
    {{ this.items[0] }} Relays
  </button>

</template>

// ------------------------------------------------------------------

<script>

import axios from "axios";
export default {
  name: "App",
  data() {
    return {
      items: [],
    };
  },
  async created() {
    try {
      const res = await axios.get(process.env.VUE_APP_API_URL + 'portal/relay_count');
      this.items = []
      this.items.push(res.data.relay_count);
    } catch (error) {
      console.log(error);
    }
  },
};

</script>

// ------------------------------------------------------------------

<style scoped>

button {
  font-weight: bold;
  font-size: 20px;
}

.relay, .relay:hover, .relay:active, .relay:visited {
  background-color: #11AA44;
  color: white;   
}

</style>

// ------------------------------------------------------------------
