<template>
  <SessionMap/>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import SessionMap from '@/components/SessionMap.vue'

/**
 * TODO: Move the filter dropdown bar to its own component
 * TODO: Cleanup template
 */

@Component({
  components: {
    SessionMap
  }
})
export default class MapWorkspace extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
