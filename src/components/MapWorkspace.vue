<template>
  <main role="main" class="col-md-12 col-lg-12 px-4">
    <div class="
              d-flex
              justify-content-between
              flex-wrap flex-md-nowrap
              align-items-center
              pt-3
              pb-2
              mb-3
              border-bottom
            "
        id="map-sessions-count"
    >
      <div id="map-spinner" v-if="false">
        <div class="spinner-border" role="status">
          <span class="sr-only">Loading...</span>
        </div>
      </div>
      <h1 class="h2" id="workspace-title" v-if="true">
        Map&nbsp;
        <span class="badge badge-dark">
          MAPSESSIONS Total Sessions
        </span>&nbsp;
        <span class="badge badge-success">
          ONNN on Network Next
        </span>
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1">
        <div class="mr-auto"></div>
        <div class="px-2 hidden">
          <select class="form-control">
            <option>
              World
            </option>
            <option>
              United States
            </option>
          </select>
        </div>
        <div class="px-2 hidden">
          <select class="form-control">
            <option value="everything">
              Everything
            </option>
            <option value="direct">
              Direct Only
            </option>
            <option value="next">
              Network Next Only
            </option>
          </select>
        </div>
        <div class="px-2 hidden">
          <select class="form-control">
            <option value="all">
              All
            </option>
            <option>
              BUYERNAME ADMIN LOOP
            </option>
            <option>
              BUYERNAME SINGLE BUYER
            </option>
          </select>
        </div>
      </div>
    </div>
    <div id="map-container"
        class="map-container-no-offset"
        v-bind:class="{
            'map-container-no-offset': true,
            'map-container-offset': false,
        }"
    ></div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

@Component
export default class MapWorkspace extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
