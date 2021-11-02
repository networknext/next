<template>
  <div class="card-body" v-if="categories.length > 0">
    <div class="card" style="margin-bottom: 250px;">
      <div class="card-header">
        <ul class="nav nav-tabs card-header-tabs">
          <li class="nav-item" v-for="(category, index) in categories" :key="index" @click="selectCategory(index)">
            <a class="nav-link" :class="{ active: index === selectedCategoryIndex }">{{ category.name }}</a>
          </li>
        </ul>
      </div>
      <div class="card-body" id="analytics-page">
        <div class="row">
          <div class="card" style="margin-bottom: 50px; width: 100%; margin: 0 1rem 2rem;">
            <div class="card-body">
              <iframe
                class="col"
                id="analyticsDash"
                :src="categories[selectedCategoryIndex].url"
                style="min-height: 3400px;"
                v-show="categories[selectedCategoryIndex].url !== ''"
                frameborder="0"
              >
              </iframe>
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
export default class Analytics extends Vue {
  private categories: Array<any>
  private selectedCategoryIndex: number

  private unwatchFilter: any

  constructor () {
    super()
    this.categories = []
    this.selectedCategoryIndex = 0
  }

  private mounted () {
    // This is only necessary for admins - when the filter changes, grab the new analytics URL
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.fetchAnalyticsCategories()
      }
    )

    this.fetchAnalyticsCategories()
  }

  private beforeDestroy () {
    this.unwatchFilter()
  }

  private fetchAnalyticsCategories () {
    this.$apiService.fetchAnalyticsCategories({
      company_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.categories = response.categories || []
      })
      .catch((error: Error) => {
        console.log('There was an issue fetching the analytics dashboard categories')
        console.log(error)
      })
  }

  private selectCategory (index: number) {
    this.selectedCategoryIndex = index
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
