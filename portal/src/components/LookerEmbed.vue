<template>
  <iframe
    class="col"
    :id="dashID"
    style="min-height: 1000px;"
    :src="dashURL"
    v-if="dashURL !== ''"
    frameborder="0"
  >
  </iframe>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'

@Component
export default class LookerEmbed extends Vue {
  @Prop() readonly dashURL!: string
  @Prop() readonly dashID!: string

  private mounted () {
    window.addEventListener('message', this.resizeIframes)
  }

  private beforeDestroy () {
    window.removeEventListener('message', this.resizeIframes)
  }

  private resizeIframes (event: any) {
    const iframes = document.querySelectorAll(`#${this.dashID}`)
    iframes.forEach((frame: any) => {
      if (event.source === frame.contentWindow && event.origin === 'https://networknextexternal.cloud.looker.com' && event.data) {
        const eventData = JSON.parse(event.data)
        if (eventData.type === 'page:properties:changed') {
          frame.height = eventData.height + 50
        }
      }
    })
  }
}
</script>

<style scoped lang="scss">
</style>
