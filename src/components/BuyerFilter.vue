<template>
  <div class="px-2" v-if="$store.getters.isBuyer || $store.getters.isAdmin">
    <select class="form-control" @change="updateFilter($event.target.value)">
      <option v-for="option in filterOptions" :key="option.value" :value="option.value" :selected="$store.getters.currentFilter.companyCode === option.value">
        {{ option.name }}
      </option>
    </select>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component is a reusable filter component
 * for filtering things based on a current buyer
 */

@Component
export default class BuyerFilter extends Vue {
  private filterOptions: Array<any>

  constructor () {
    super()
    this.filterOptions = []
  }

  private mounted () {
    this.filterOptions.push({
      name: 'All',
      value: ''
    })

    this.$store.getters.allBuyers.forEach((buyer: any) => {
      if (!this.$store.getters.isAdmin || (this.$store.getters.isAdmin && buyer.is_live)) {
        this.filterOptions.push({
          name: buyer.company_name,
          value: buyer.company_code
        })
      }
    })
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
    this.$store.commit('UPDATE_CURRENT_FILTER', { companyCode: companyCode })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
