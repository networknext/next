// -----------------------------------------------------------------

<template>

  <div v-if="this.updated">

    <header>
      <h1>Relay</h1>
    </header>

    <p class="relay_name">relay name = {{ $route.params.id }}</p>

    <p class="current_sessions = ">sessions = {{ this.current_sessions }}</p>

  </div>

</template>

// -----------------------------------------------------------------

<script>

import axios from "axios";
import update from '@/update.js'

export default {

  name: 'App',

  mixins: [update],

  data() {
    return {
      decimal: 0,
      current_sessions: 0
    }
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/relay/' + this.$route.params.id)
        if (res.data.relay_data !== null) {
          this.current_sessions = res.data.relay_data.num_sessions
          this.updated = true
        }
      } catch (error) {
        console.log(error)
      }
    }

  }

}

</script>

// -----------------------------------------------------------------

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

session {
  font-family: fixed-width;
}

</style>

// -----------------------------------------------------------------
