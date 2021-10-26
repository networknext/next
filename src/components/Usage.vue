<template>
  <div class="card-body" id="usageDash-page">
    <div class="row">
      <iframe
        class="col"
        id="usageDash"
        :src="usageDashDashURL"
        style="min-height: 5200px;"
        v-if="usageDashDashURL !== ''"
        frameborder="0"
      >
      </iframe>
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
</style>
