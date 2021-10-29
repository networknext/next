<template>
  <div class="card-body" id="usageDash-page">
    <div v-for="(url, index) in usageDashDashURLs" :key="index" class="row">
      <div class="card" style="margin-bottom: 50px; width: 100%; margin: 0 1rem 2rem;">
        <div class="card-body">
          <iframe
            class="col"
            id="usageDash"
            :src="url"
            :style="{'min-height': index === 0 ? '2300px' : '3400px'}"
            v-show="url !== ''"
            frameborder="0"
          >
          </iframe>
        </div>
      </div>
    </div>
    <div class="row">
      <div class="card" style="margin-bottom: 50px; width: 100%; margin: 0 1rem 2rem;">
        <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
          <div class="mb-2 mb-md-0 flex-grow-1"></div>
          <div class="pr-5">
            Payment Instructions
          </div>
          <div class="mb-2 mb-md-0 flex-grow-1"></div>
          <div class="pr-5">
            <font-awesome-icon
              aria-expanded="false"
              id="status"
              icon="chevron-left"
              class="fa-w-16 fa-fw"
              type="button"
              data-toggle="collapse"
              data-target="#payment-instructions"
            />
            <font-awesome-icon
              aria-expanded="false"
              id="status"
              icon="chevron-down"
              class="fa-w-16 fa-fw"
              type="button"
              data-toggle="collapse"
              data-target="#payment-instructions"
            />
          </div>
        </div>
        <div class="collapse" id="payment-instructions">
          <div class="card-body">
            <div class="row" style="padding: 0 1rem;">
              <h2 class="col">Payment Instructions</h2>
            </div>
            <div class="row" style="padding: 0 1rem 1rem;">
              <h5 class="col">Wire Transfer</h5>
            </div>
            <div class="row" style="padding: 1rem;">
              <div class="col"></div>
              <div class="col">
                Wire to:
                <div class="blue-accent"></div>
              </div>
              <div class="col">
                336 Bon Air Center<br>
                Suite 136<br>
                Greenbrae, CA 94904<br>
                (310) 775-0041<br>
              </div>
              <div class="col"></div>
            </div>
            <div class="row" style="padding: 1rem;">
              <div class="col"></div>
              <div class="col">
                Swift Code:
                <div class="blue-accent"></div>
              </div>
              <div class="col">
                CHASUS33
              </div>
              <div class="col"></div>
            </div>
            <div class="row" style="padding: 1rem;">
              <div class="col"></div>
              <div class="col">
                ABA (Routing Number)
                <div class="blue-accent"></div>
              </div>
              <div class="col">
                322271627
              </div>
              <div class="col"></div>
            </div>
            <div class="row" style="padding: 1rem;">
              <div class="col"></div>
              <div class="col">
                Account Number
                <div class="blue-accent"></div>
              </div>
              <div class="col">
                366991757
              </div>
              <div class="col"></div>
            </div>
            <div class="row" style="padding: 1rem;">
              <div class="col"></div>
              <div class="col">
                Account Name
                <div class="blue-accent"></div>
              </div>
              <div class="col">
                Network Next, Inc.
              </div>
              <div class="col"></div>
            </div>
            <div class="row" style="padding: 1rem;">
              <div class="col"></div>
              <div class="col">
                Address
                <div class="blue-accent"></div>
              </div>
              <div class="col">
                2333 Payne Road,<br>
                Castleton on Hudson,<br>
                NY, 12033
              </div>
              <div class="col"></div>
            </div>
            <div class="row" style="padding: 0 1rem 1rem;">
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Usage extends Vue {
  private usageDashDashURLs: Array<string>

  private unwatchFilter: any

  constructor () {
    super()
    this.usageDashDashURLs = []
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new billing URL
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchUsageSummary()
      }
    )

    this.fetchUsageSummary()

    const usageDashElement = document.getElementById('usageDash')
    if (usageDashElement) {
      usageDashElement.addEventListener('dashboard:run:complete', this.iframeTimeoutHandler)
    }

    // TODO: Add more hooks here to optimize what the user sees while the dashes load...
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchUsageSummary () {
    this.$apiService.fetchUsageSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.usageDashDashURLs = response.urls || []
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the billing summary dashboard')
        console.log(error)
      })
  }

  private iframeTimeoutHandler () {
    // TODO: Look for a status of error or stopped and display a refresh page message....
    console.log('An iframe timed out we should add an alert to refresh the page!')
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .blue-accent {
    border-bottom: solid #009FDF;
    width: 2.2rem;
    padding-bottom: 2px;
  }
  [aria-expanded=true].fa-chevron-left {
    display: none;
  }
  [aria-expanded=false].fa-chevron-down {
    display: none;
  }
</style>
