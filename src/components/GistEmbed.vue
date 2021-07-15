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
    const gistContainer = document.getElementById('gist-embed')
    if (gistContainer) {
      gistContainer.insertAdjacentHTML('afterend', this.embedHTML)
      const metaElements = document.getElementsByClassName('gist-meta')
      const dataElements = document.getElementsByClassName('gist-data')

      for (let i = dataElements.length - 1; i >= 0; --i) {
        dataElements[i].classList.replace('gist-data', 'corners')
      }

      for (let i = metaElements.length - 1; i >= 0; --i) {
        metaElements[i].classList.add('hidden')
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .hidden {
    display: none;
  }
  .corners {
    border-radius: 6px;
    word-wrap: normal;
    background-color: var(--color-bg-primary);
    overflow: auto;
  }
</style>
