<template>

  <div v-if="this.updated" class="btn-group" role="group" aria-label="Basic example">
    <button type="button" class="btn total">
      <div class="d-xxl-none"> {{ this.total_sessions }} </div>
      <div class="d-none d-xxl-block"> {{ this.total_sessions }} Total Sessions </div>
    </button>

    <button type="button" class="btn next">
      <div class="d-xxl-none"> {{ this.next_sessions }} </div>
      <div class="d-none d-xxl-block"> {{ this.next_sessions }} on Network Next </div>
    </button>
  </div>

</template>

<script>

import axios from "axios";
import update from "@/update.js"

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      total_sessions: 0,
      next_sessions: 0,
    };
  },

  created: function () {
    this.update()
  },

  methods: {

    async update() {
      try {
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/session_counts')
        this.total_sessions = res.data.total_session_count.toLocaleString()
        this.next_sessions = res.data.next_session_count.toLocaleString()
        this.updated = true
      } catch (error) {
        console.log(error);
      }
    }
  },

};

</script>

<style scoped>

button {
  font-weight: bold;
  font-size: 15px;
}

.total, .total:hover, .total:active, .total:visited {
  background-color: #354040;
  color: white;   
}

.next, .next:hover, .next:active, .next:visited {
  background-color: #11AA44;
  color: white;   
}

button {
    white-space: nowrap;
}

</style>
