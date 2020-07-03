<template>
  <div class="container-fluid below-nav-bar">
    <div class="row">
        <map-workspace></map-workspace>
        <sessions-workspace v-if="false"></sessions-workspace>
        <session-tool-workspace v-if="false"></session-tool-workspace>
        <user-tool-workspace v-if="false"></user-tool-workspace>
        <downloads-workspace v-if="false"></downloads-workspace>
        <settings-workspace v-if="false"></settings-workspace>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import MapWorkspace from './MapWorkspace.vue'
import SessionsWorkspace from './SessionsWorkspace.vue'
import SessionToolWorkspace from './SessionToolWorkspace.vue'
import SettingsWorkspace from './SettingsWorkspace.vue'

@Component({
  components: {
    MapWorkspace,
    SessionsWorkspace,
    SessionToolWorkspace,
    SettingsWorkspace
  }
})
export default class Workspace extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .below-nav-bar {
    padding-top: 48px;
  }
</style>
