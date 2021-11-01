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
              id="sdk-button"
              v-on:click="downloadSDK()"
              class="btn btn-primary m-1 btn-width"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              SDK v4.0.16
            </a>
            <a
              href="#"
              id="docs-button"
              v-on:click="downloadUE4()"
              class="btn btn-primary m-1 btn-width"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              UE4 Plugin
            </a>
            <a
              href="#"
              id="docs-button"
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
          title: 'SDK & Documentation'
        },
        content: 'Get our open source SDK and view our latest Documentation here.<br><br>Integration instructions are in the Getting Started section of the Documentation.<br><br>Please contact us in chat (lower right) if you have any questions.',
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
        buttonStop: 'OK'
      }
    }

    this.downloadsTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_SIGN_UP_TOURS', 'downloads')
        if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          this.$gtag.event('Downloads tour finished', {
            event_category: 'Tours'
          })
        }
      }
    }
  }

  private mounted () {
    if (this.$store.getters.isSignUpTour && this.$tours.downloadsTour && !this.$tours.downloadsTour.isRunning && !this.$store.getters.finishedSignUpTours.includes('downloads')) {
      this.$tours.downloadsTour.start()
    }
  }

  private downloadSDK () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event('sdk-download', {
        event_category: 'Important Clicks'
      })
    }
    window.open('https://storage.googleapis.com/portal_sdk_download_storage/next-4.0.16.zip')
    this.$apiService.sendSDKDownloadSlackNotification({ email: this.$store.getters.userProfile.email, customer_name: this.$store.getters.userProfile.companyName, customer_code: this.$store.getters.userProfile.companyCode })
  }

  private downloadDocs () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event('SDK-docs-download', {
        event_category: 'Important Clicks'
      })
    }
    window.open('https://network-next-sdk.readthedocs-hosted.com/en/latest/')
    this.$apiService.sendDocsViewSlackNotification({ email: this.$store.getters.userProfile.email, customer_name: this.$store.getters.userProfile.companyName, customer_code: this.$store.getters.userProfile.companyCode })
  }

  private downloadUE4 () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event('ue4-download', {
        event_category: 'Important Clicks'
      })
    }
    window.open('https://storage.googleapis.com/network-next-ue4/ue4-plugin.zip')
    this.$apiService.sendUE4DownloadNotifications({ email: this.$store.getters.userProfile.email, customer_name: this.$store.getters.userProfile.companyName, customer_code: this.$store.getters.userProfile.companyCode })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .btn-width {
    width: "160px";
  }
  #sdk-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #docs-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #sdk-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
  #docs-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
</style>
