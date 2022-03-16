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
              id="sdk-button"
              @click="downloadSDK()"
              class="btn btn-primary m-1 btn-width white-link"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              SDK v4.20
            </a>
            <a
              id="ue4-button"
              @click="downloadUE4()"
              class="btn btn-primary m-1 btn-width white-link"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              UE4 Plugin
            </a>
            <a
              id="enet-button"
              @click="downloadEnet()"
              class="btn btn-primary m-1 btn-width white-link"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              ENet Support
            </a>
            <a
              id="docs-button"
              @click="downloadDocs()"
              class="btn btn-primary m-1 btn-width white-link"
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
      <hr v-if="false" />
      <div class="card-body" v-if="false">
        <div class="btn-group-vertical btn-group-sm float-right">
          <div style="display: inherit;flex-direction: column;" data-intercom="dataReports" data-tour="dataReports">
            <a
              id="white-paper-button"
              @click="download2022WhitePaper()"
              class="btn btn-primary m-1 btn-width white-link"
            >
              <font-awesome-icon icon="download"
                                  class="fa-w-16 fa-fw"
              />
              Download
            </a>
          </div>
        </div>
        <h5>
          State of the Internet for Real-Time Games 2022 Whitepaper <br />
        </h5>
        <div style="font-size: 90%;">Analysis of over 10 million unique players identifies critical need for session analytics and selective augmented Internet services</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import {
  ENET_DOWNLOAD_EVENT,
  ENET_DOWNLOAD_URL,
  IMPORTANT_CLICKS_CATEGORY,
  SDK_DOCUMENTATION_EVENT,
  SDK_DOCUMENTATION_URL,
  SDK_DOWNLOAD_EVENT,
  SDK_DOWNLOAD_URL,
  UE4_PLUGIN_DOWNLOAD_EVENT,
  UE4_PLUGIN_DOWNLOAD_URL,
  WHITE_PAPER_DOWNLOAD_EVENT,
  WHITE_PAPER_DOWNLOAD_URL
} from '@/components/types/Constants'
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
      },
      {
        target: '[data-tour="dataReports"]',
        header: {
          title: 'Reports'
        },
        content: 'Get access to Network Next curated reports that detail GDRP and CCPA compliant data analyses.',
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
      this.$gtag.event(SDK_DOWNLOAD_EVENT, {
        event_category: IMPORTANT_CLICKS_CATEGORY
      })
    }
    window.open(SDK_DOWNLOAD_URL)
    this.$apiService.sendSDKDownloadSlackNotification({
      email: this.$store.getters.userProfile.email,
      customer_name: this.$store.getters.userProfile.companyName,
      customer_code: this.$store.getters.userProfile.companyCode
    })
  }

  private downloadDocs () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event(SDK_DOCUMENTATION_EVENT, {
        event_category: IMPORTANT_CLICKS_CATEGORY
      })
    }
    window.open(SDK_DOCUMENTATION_URL)
    this.$apiService.sendDocsViewSlackNotification({
      email: this.$store.getters.userProfile.email,
      customer_name: this.$store.getters.userProfile.companyName,
      customer_code: this.$store.getters.userProfile.companyCode
    })
  }

  private downloadUE4 () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event(UE4_PLUGIN_DOWNLOAD_EVENT, {
        event_category: IMPORTANT_CLICKS_CATEGORY
      })
    }
    window.open(UE4_PLUGIN_DOWNLOAD_URL)
    this.$apiService.sendUE4DownloadNotifications({
      email: this.$store.getters.userProfile.email,
      customer_name: this.$store.getters.userProfile.companyName,
      customer_code: this.$store.getters.userProfile.companyCode
    })
  }

  private downloadEnet () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event(ENET_DOWNLOAD_EVENT, {
        event_category: IMPORTANT_CLICKS_CATEGORY
      })
    }
    window.open(ENET_DOWNLOAD_URL)
    this.$apiService.sendENetDownloadNotification({
      email: this.$store.getters.userProfile.email,
      customer_name: this.$store.getters.userProfile.companyName,
      customer_code: this.$store.getters.userProfile.companyCode
    })
  }

  private download2022WhitePaper () {
    if (this.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      this.$gtag.event(WHITE_PAPER_DOWNLOAD_EVENT, {
        event_category: IMPORTANT_CLICKS_CATEGORY
      })
    }
    window.open(WHITE_PAPER_DOWNLOAD_URL)
    this.$apiService.send2022WhitePaperDownloadNotifications({
      email: this.$store.getters.userProfile.email,
      customer_name: this.$store.getters.userProfile.companyName,
      customer_code: this.$store.getters.userProfile.companyCode
    })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .btn-width {
    min-width: 180px;
  }
  #sdk-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #docs-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #white-paper-button {
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
  #enet-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #enet-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
  #white-paper-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #white-paper-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
  #ue4-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #ue4-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
  .white-link {
    color: white;
  }
</style>
