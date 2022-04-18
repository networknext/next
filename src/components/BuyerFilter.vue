<template>
  <div class="row" id="buyer-filter">
    <div class="col">
      <select class="form-control" @change="updateFilter($event.target.value)">
        <option v-for="option in filterOptions" :key="option.value" :value="option.value" :selected="$store.getters.currentFilter.companyCode === option.value">
          {{ option.name }}
        </option>
      </select>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator'
import { Filter } from './types/FilterTypes'

/**
 * This component is a reusable filter component
 * for filtering things based on a current buyer
 */

@Component
export default class BuyerFilter extends Vue {
  @Prop({ default: true }) readonly includeAll!: boolean
  @Prop({ default: true }) readonly liveOnly!: boolean

  private filterOptions: Array<any>
  private unwatchBuyerList: any

  constructor () {
    super()
    this.filterOptions = []
  }

  private mounted () {
    if (this.$store.getters.isAdmin) {
      this.unwatchBuyerList = this.$store.watch(
        (state: any, getters: any) => {
          return getters.allBuyers
        },
        () => {
          this.setupFilters()
        }
      )
    }

    this.setupFilters()
  }

  private beforeDestroy () {
    if (this.$store.getters.isAdmin) {
      this.unwatchBuyerList()
    }
  }

  private setupFilters () {
    this.filterOptions = []

    if (this.includeAll) {
      this.filterOptions.push({
        name: 'All',
        value: ''
      })
    }

    this.$store.getters.allBuyers.forEach((buyer: any) => {
      if (
        (!this.$store.getters.isAdmin && this.$store.getters.userProfile.companyCode === buyer.company_code) ||
        (this.$store.getters.isAdmin && (buyer.is_live || !this.liveOnly))
      ) {
        this.filterOptions.push({
          name: buyer.company_name,
          value: buyer.company_code
        })
      }
    })

    // If we aren't showing all (Admin filter) then we want to just pick the first option in the list
    if (!this.includeAll && this.$store.getters.currentFilter.companyCode === '') {
      const newFilter: Filter = {
        companyCode: this.filterOptions[0].value,
        dateRange: this.$store.getters.currentFilter.dateRange
      }

      this.$store.dispatch('updateCurrentFilter', newFilter)
    }
  }

  private getBuyerName () {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].company_code === this.$store.getters.userProfile.companyCode) {
        return allBuyers[i].company_name
      }
    }
    return 'Private'
  }

  private updateFilter (companyCode: string) {
    const newFilter: Filter = {
      companyCode: companyCode,
      dateRange: this.$store.getters.currentFilter.dateRange
    }

    this.$store.dispatch('updateCurrentFilter', newFilter)
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
