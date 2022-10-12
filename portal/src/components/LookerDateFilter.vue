<template>
  <div id="date-filter" class="px-2">
    <select class="form-control" @change="updateFilter($event.target.value)">
      <option v-for="option in filterOptions" :key="option.value" :value="option.value" :selected="$store.getters.currentFilter.dateRange === option.value">
        {{ option.name }}
      </option>
    </select>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { DateFilterType, Filter, LookerDateFilterOption } from '@/components/types/FilterTypes'

/**
 * This component is a reusable filter component
 * for filtering things based on a current buyer
 */

@Component
export default class LookerDateFilter extends Vue {
  private filterOptions: Array<LookerDateFilterOption>

  constructor () {
    super()
    this.filterOptions = []
  }

  private mounted () {
    this.filterOptions = [
      {
        name: 'Last 7 Days',
        value: DateFilterType.LAST_7
      }
    ]

    if (this.$store.getters.isOwner || this.$store.getters.isExplorer) {
      this.filterOptions = this.filterOptions.concat([
        {
          name: 'Last 14 Days',
          value: DateFilterType.LAST_14
        },
        {
          name: 'Last 30 Days',
          value: DateFilterType.LAST_30
        }
      ])

      if (this.$store.getters.isExplorer && this.$store.getters.hasAnalytics) {
        this.filterOptions = this.filterOptions.concat([
          {
            name: 'Last 60 Days',
            value: DateFilterType.LAST_60
          },
          {
            name: 'Last 90 Days',
            value: DateFilterType.LAST_90
          }
        ])
      }
    }
  }

  // TODO: Figure out best way to offer custom date range
  private updateFilter (dateRange: DateFilterType) {
    const newFilter: Filter = {
      companyCode: this.$store.getters.currentFilter.companyCode,
      dateRange: dateRange
    }

    this.$store.dispatch('updateCurrentFilter', newFilter)
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
