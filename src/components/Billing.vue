<template>
  <div class="card-body" id="billing-page">
    <h5 class="card-title">Billing Dashboard</h5>
    <p class="card-text">
      One stop shop for billing information
    </p>
    <iframe
      id="billingDash"
      :src="billingDashURL"
      v-if="showSummary"
      width="1200"
      height="1600"
      frameborder="0"
    >
    </iframe>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Billing extends Vue {
  private billingDashURL: string
  private showSummary: boolean

  private unwatchFilter: any

  private startDate: string
  private endDate: string

  constructor () {
    super()
    this.billingDashURL = ''
    this.showSummary = false

    this.startDate = ''
    this.endDate = ''
  }

  private mounted () {
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchBillingSummary()
      }
    )

    this.fetchBillingSummary()
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchBillingSummary () {
    this.$apiService.fetchBillingSummary({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.billingDashURL = response.url || ''
        if (this.billingDashURL !== '') {
          this.showSummary = true
        }
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
</style>
