<template>
  <div class="card-body" id="usageDash-page">
    <div class="row">
      <iframe
        class="col"
        id="usageDash"
        :src="usageDashDashURL"
        style="min-height: 2500px;"
        v-show="usageDashDashURL !== ''"
        frameborder="0"
      >
      </iframe>
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
                JP Morgan Chase<br>
                631 Wilshire Blvd Ste A<br>
                Santa Monica, CA 90401<br>
                (310) 309-3152<br>
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
  private usageDashDashURL: string

  private unwatchFilter: any

  constructor () {
    super()
    this.usageDashDashURL = ''
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
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchUsageSummary () {
    this.$apiService.fetchUsageSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.usageDashDashURL = response.url || ''
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the billing summary dashboard')
        console.log(error)
      })
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
