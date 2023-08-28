// ------------------------------------------------------------------

<template>

  <button type="button" class="btn server">
    {{ this.items[0] }} Servers
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
      const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/server_count');
      this.items = []
      this.items.push(res.data.server_count);
    } catch (error) {
      console.log(error);
    }
  },
};
</script>

// ------------------------------------------------------------------

<style scoped>

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

button {
  font-weight: bold;
  font-size: 20px;
}

.server, .server:hover, .server:active, .server:visited {
  background-color: #11AA44;
  color: white;   
}

</style>

// ------------------------------------------------------------------
