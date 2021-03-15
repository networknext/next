<template>
  <div>
    <v-tour name="downloadsTour" :steps="downloadsTourSteps" :options="downloadsTourOptions" :callbacks="downloadsTourCallbacks"></v-tour>
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
          <div style="display: inherit;flex-direction: column;" data-intercom="sdkDocumentation" data-tour="sdkDocumentation">
            <a
              href="#"
              v-on:click="downloadSDK()"
              class="btn btn-primary m-1 btn-width"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              SDK v4.0.10
            </a>
            <a
              href="#"
              v-on:click="downloadDocs()"
              class="btn btn-primary m-1 btn-width"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              Documentation
            </a>
          </div>
      </div>
      <h5 class="card-title">
          Network Next SDK
      </h5>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { FeatureEnum } from '@/components/types/FeatureTypes'
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component
export default class DownloadsWorkspace extends Vue {
  private downloadsTourSteps: Array<any>
  private downloadsTourOptions: any
  private downloadsTourCallbacks: any

  constructor () {
    super()

    this.downloadsTourSteps = [
      {
        target: '[data-tour="sdkDocumentation"]',
        header: {
          title: 'SDK and Documentation'
        },
        content: 'Get our open source SDK and view our latest Documentation here.\nIntegration instructions are in the Getting Started section of the Documentation.\nPlease contact us in chat (lower right) if you have any questions.',
        params: {
          placement: 'left'
        }
      }
    ]

    this.downloadsTourOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'DONE'
      }
    }

    this.downloadsTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_SIGN_UP_TOURS', 'downloads')
      },
      onSkip: () => {
        this.$store.commit('UPDATE_FINISHED_SIGN_UP_TOURS', 'downloads')
      }
    }
  }

  private mounted () {
    if (this.$store.getters.isSignUpTour && this.$route.name === 'downloads' && this.$tours.downloadsTour && !this.$tours.downloadsTour.isRunning) {
      this.$tours.downloadsTour.start()
    }
  }

  private downloadSDK () {
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event('sdk-download', {
        event_category: 'Important Clicks'
      })
    }
    window.open('https://storage.googleapis.com/portal_sdk_download_storage/next-4.0.10.zip')
  }

  private downloadDocs () {
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event('SDK-docs-download', {
        event_category: 'Important Clicks'
      })
    }
    window.open('https://network-next-sdk.readthedocs-hosted.com/en/latest/')
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .btn-width {
    width: "160px";
  }
</style>
