<template>
  <div class="card-body" id="billing-page">
    <div class="row">
      <iframe
        class="col"
        id="billingDash"
        :src="billingDashURL"
        style="min-height: 5200px;"
        v-if="billingDashURL !== ''"
        frameborder="0"
      >
      </iframe>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Billing extends Vue {
  private billingDashURL: string

  private unwatchFilter: any

  constructor () {
    super()
    this.billingDashURL = ''
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new billing URL
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
