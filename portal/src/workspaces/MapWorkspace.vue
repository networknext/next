<template>
  <SessionMap />
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import SessionMap from '@/components/SessionMap.vue'

/**
 * This component holds the workspace elements related to the map page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    SessionMap
  }
})
export default class MapWorkspace extends Vue {
  private mounted () {
    if (this.$store.getters.isTour && this.$tours.mapTour && !this.$tours.mapTour.isRunning && !this.$store.getters.finishedTours.includes('map')) {
      this.$tours.mapTour.start()
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
