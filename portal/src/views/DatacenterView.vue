// ---------------------------------------------------------------

<template>
  <header>
    <h1>Datacenter</h1>
  </header>

  <p class="datacenter_name">datacenter name = {{ $route.params.id }}</p>

  <p class="latlong">location = {{ this.latitude + "," + this.longitude }}</p>

</template>

// ---------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

export default {

  name: 'App',

  mixins: [update],

  data() {
    return {
      latitude: 0,
      longitude: 0,
    }
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/datacenter/' + this.$route.params.id)
        if (res.data.session_data !== null) {
          this.latitude = res.data.datacenter_data.latitude
          this.longitude = res.data.datacenter_data.longitude
          this.updated = true
        }
      } catch (error) {
        console.log(error)
      }
    }

  }

}

</script>

// ---------------------------------------------------------------

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

session {
  font-family: fixed-width;
}

</style>

// ---------------------------------------------------------------
