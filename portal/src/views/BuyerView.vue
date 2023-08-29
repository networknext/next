// --------------------------------------------------------------

<template>
  <header>
    <h1>Buyer</h1>
  </header>

  <p class="buyer_code">buyer code = {{ $route.params.id }}</p>

  <p class="buyer_name">buyer name = {{ this.buyer_name }}</p>

</template>

// --------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

export default {

  name: 'App',

  mixins: [update],

  data() {
    return {
      buyer_name: '',
    }
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/buyer/' + this.$route.params.id)
        if (res.data.buyer_data !== null) {
          this.buyer_name = res.data.buyer_data.name
          this.updated = true
        }
      } catch (error) {
        console.log(error)
      }
    }

  }

}

</script>

// --------------------------------------------------------------

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

session {
  font-family: fixed-width;
}

</style>

// --------------------------------------------------------------
