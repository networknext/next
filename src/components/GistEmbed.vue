<template>
  <div id="gist-embed">
    <link v-if="cssURL !== ''" rel="stylesheet" :href="cssURL">
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator'

@Component
export default class GistEmbed extends Vue {
  @Prop() readonly cssURL!: string
  @Prop() readonly embedHTML!: string

  private mounted () {
    console.log(document.getElementById('gist-embed'))

    const gistContainer = document.getElementById('gist-embed')

    if (gistContainer) {
      gistContainer.insertAdjacentHTML('afterend', this.embedHTML)
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
