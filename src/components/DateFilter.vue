<template>
  <div class="px-2">
    <select class="form-control" @change="updateFilter($event.target.value)">
      <option v-for="option in filterOptions" :key="option.value" :value="option.value" :selected="$store.getters.currentFilter.dateRange === option.value">
        {{ option.name }}
      </option>
    </select>
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'

/**
 * This component is a reusable filter component
 * for filtering things based on a current buyer
 */

@Component
export default class DateFilter extends Vue {
  @Prop({ default: false }) readonly includeCustom!: boolean

  private filterOptions: Array<any>
  private startDate: string
  private endDate: string

  constructor () {
    super()
    this.filterOptions = []
    this.startDate = ''
    this.endDate = ''
  }

  private mounted () {
    this.filterOptions = [
      {
        name: 'Current month',
        value: DateFilterType.CURRENT_MONTH
      },
      {
        name: 'Last month',
        value: DateFilterType.LAST_MONTH
      },
      {
        name: 'Last 30 days',
        value: DateFilterType.LAST_30
      },
      {
        name: 'Last 90 days',
        value: DateFilterType.LAST_90
      },
      {
        name: 'Year to date',
        value: DateFilterType.YEAR_TO_DATE
      }
    ]

    if (this.includeCustom) {
      this.filterOptions.push({
        name: 'Custom range',
        value: DateFilterType.CUSTOM
      })
    }
  }

  // TODO: Figure out best way to offer custom date range
  private updateFilter (dateRange: DateFilterType) {
    const newFilter: Filter = {
      companyCode: this.$store.getters.currentFilter.companyCode,
      dateRange: dateRange
    }

    this.$store.commit('UPDATE_CURRENT_FILTER', newFilter)
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
