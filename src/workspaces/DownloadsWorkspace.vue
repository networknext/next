<template>
  <div>
    <div class="
              d-flex
              justify-content-between
              flex-wrap
              flex-md-nowrap
              align-items-center
              pt-3
              pb-2
              mb-3
              border-bottom
            "
    >
      <h1 class="h2">
        Downloads
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
        <div class="mr-auto"></div>
      </div>
    </div>
    <div class="card mb-2">
      <div class="card-body">
        <div class="btn-group-vertical btn-group-sm float-right">
          <a
            href="#"
            onclick="window.open('https://storage.googleapis.com/portal_sdk_download_storage/next-4.0.1.zip')"
            class="btn btn-primary m-1 btn-width"
            title="SDK v4.0.1"
          >
            <font-awesome-icon icon="download"
                                class="fa-w-16 fa-fw"
            />
            SDK v4.0.1
          </a>
          <a
            href="#"
            onclick="window.open('https://network-next-sdk.readthedocs-hosted.com/en/latest/')"
            class="btn btn-primary m-1 btn-width"
            title="Documentation"
          >
            <font-awesome-icon icon="download"
                                class="fa-w-16 fa-fw"
            />
            Documentation
          </a>
      </div>
      <h5 class="card-title">
          Network Next SDK
      </h5>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component
export default class DownloadsWorkspace extends Vue {
  // Empty for now
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .btn-width {
    width: "160px";
  }
</style>
