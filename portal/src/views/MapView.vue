// ----------------------------------------------------

<template>

  <div class="parent">
  
    <div class="map">

    </div>

  </div>

</template>

// ----------------------------------------------------

<script>

import update from "@/update.js"

export default {

  name: 'App',

  mixins: [update],

  mounted: function() {
    this.$emit('notify-view', 'map')
  },

  methods: {

    async update() {
      try {
        fetch(process.env.VUE_APP_API_URL + '/portal/map_data', {
          headers: {
            'Accept': 'application/octet-stream'
           }
        })
        .then(response => response.text())
        .then(data => console.log("got map data (" + data.length + " bytes)"))
        this.updated = true
        this.$emit('notify-update')
      } catch (error) {
        console.log(error);
      }
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

.parent {
  width: 100%;
  height: 100%;
  padding: 0px;
}

.map {
  background-color: #555555;
  width: 100%;
  height: 100%;
}

</style>

// ----------------------------------------------------
