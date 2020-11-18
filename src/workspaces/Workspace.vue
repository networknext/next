<template>
  <div class="container-fluid below-nav-bar">
    <div class="row">
      <main role="main" class="col-md-12 col-lg-12 px-4">
        <SessionCounts
          v-if="$store.getters.currentPage == 'map' || $store.getters.currentPage == 'sessions'"
        />
        <router-view />
      </main>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import SessionCounts from '@/components/SessionCounts.vue'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'

/**
 * This component is the base component for all other workspace components
 * It also holds the Email Verification alert and Session Count components
 *  so that they are visible across all workspaces if necessary
 */

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
