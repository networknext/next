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
      console.log('update map')
      fetch('http://dev.virtualgo.net/portal/map_data', {
        headers: {
          'Accept': 'application/octet-stream'
         }
      })
      .then(response => response.text())
      .then(data => console.log("got map data " + data))
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
