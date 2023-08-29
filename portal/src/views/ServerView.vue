// --------------------------------------------------------------------

<template>
  <header>
    <h1>Server</h1>
  </header>

  <p class="server_id">server address = {{ $route.params.id }}</p>

  <p class="session_count">sessions = {{ this.sessions }}</p>

</template>

// --------------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

export default {

  name: 'App',

  mixins: [update],

  data() {
    return {
      sessions: 0
    };
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/server/' + this.$route.params.id)
        if (res.data.server_data !== null) {
          this.sessions = res.data.server_data.num_sessions
          this.updated = true
        }
      } catch (error) {
        console.log(error);
      }
    }

  }

}

</script>

// --------------------------------------------------------------------

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

session {
  font-family: fixed-width;
}

</style>

// --------------------------------------------------------------------
