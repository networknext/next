// ----------------------------------------------------

<template>
  <h1>Map</h1>
</template>

// ----------------------------------------------------

<script>

import update from "@/update.js"

export default {

  name: 'App',

  mixins: [update],

  methods: {

    async update() {
      /*
      try {
        console.log('update map')
        fetch(process.env.VUE_APP_API_URL + '/portal/map_data', {
          headers: {
            'Accept': 'application/octet-stream'
           }
        })
        .then(response => response.text())
        .then(data => console.log("got map data " + data))
        this.updated = true
      } catch (error) {
        console.log(error);
      }
      */
    }

  }

}

</script>

// ----------------------------------------------------

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

</style>

// ----------------------------------------------------
