// -------------------------------------------------------------

<template>
  <header>
    <h1>Seller</h1>
  </header>

  <p class="seller_code">seller code = {{ $route.params.id }}</p>

  <p class="seller_name">seller name = {{ this.seller_name }}</p>

</template>

// -------------------------------------------------------------

<script>

import axios from "axios";
import update from "@/update.js"

export default {

  name: 'App',

  mixins: [update],

  data() {
    return {
      seller_name: '',
    }
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/seller/' + this.$route.params.id)
        if (res.data.seller_data != null) {
          this.seller_name = res.data.seller_data.name
          this.updated = true
        }
      } catch (error) {
        console.log(error)
      }
    }

  }


}
</script>

// -------------------------------------------------------------

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

session {
  font-family: fixed-width;
}

</style>

// -------------------------------------------------------------
