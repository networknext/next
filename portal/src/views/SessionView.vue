<template>

  <div v-if="this.updated">

    <header>
      <h1>Session</h1>
    </header>

    <p class="session">session id = {{ $route.params.id }}</p>

    <p class="user">user hash = {{ this.user_hash }}</p>

    <p class="slices">slices = {{ this.num_slices }}</p>

    <p class="near_relays">near relays = {{ this.num_near_relays }}</p>

    <p class="buyer_code">buyer code = {{ this.buyer_code }}</p>

    <p class="datacenter_name">datacenter name = {{ this.datacenter_name }}</p>

  </div>

</template>

<script>

import axios from "axios";
import utils from '@/utils.js'
import update from '@/update.js'

export default {

  name: "App",

  mixins: [utils, update],

  methods: {

    async update() {
      try {
        const session_id = this.$route.params.id
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/session/' + session_id)
        if (res.data.slice_data !== null) {
          this.user_hash = this.parse_uint64(res.data.session_data.user_hash)
          this.buyer_code = res.data.session_data.buyer_code
          this.datacenter_name = res.data.session_data.datacenter_name
          this.num_slices = res.data.slice_data.length
          this.num_near_relays = res.data.near_relay_data[0].num_near_relays
        }
      } catch (error) {
        console.log(error)
      }
    }
  },

  data() {
    return {
      fields: ["Session ID", "ISP", "Buyer", "Datacenter", "Server Address", "Direct RTT", "Next RTT", "Improvement"],
      data: [],
      user_hash: 0,
      num_slices: 0,
      num_near_relays: 0,
      buyer_code: '',
      relay_name: ''
    };
  },

};

</script>

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

session {
  font-family: fixed-width;
}

</style>