<template>
  <div class="container-fluid below-nav-bar">
    <SessionCounts v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'"/>
    <div class="row">
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import MapWorkspace from './MapWorkspace.vue'
import SessionCounts from '../SessionCounts.vue'
import SessionsWorkspace from './SessionsWorkspace.vue'
import SessionToolWorkspace from './SessionToolWorkspace.vue'
import SettingsWorkspace from './SettingsWorkspace.vue'

@Component({
  components: {
    MapWorkspace,
    SessionCounts,
    SessionsWorkspace,
    SessionToolWorkspace,
    SettingsWorkspace
  }
})
export default class Workspace extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style scoped lang="scss">
</style>
