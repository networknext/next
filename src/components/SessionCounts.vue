<template>
  <h1 class="count-header">
    Map&nbsp;
    <span class="badge badge-dark">
      {{ this.totalSessions }} Total Sessions
    </span>&nbsp;
    <span class="badge badge-success">
      {{ this.totalSessionsReply.onNN }} on Network Next
    </span>
  </h1>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

interface TotalSessionsReply {
  direct: number;
  onNN: number;
}

@Component
export default class SessionCounts extends Vue {
  private totalSessionsReply: TotalSessionsReply

  get totalSessions () {
    return this.totalSessionsReply.direct + this.totalSessionsReply.onNN
  }

  constructor () {
    super()
    this.totalSessionsReply = {
      direct: 0,
      onNN: 0
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .count-header {
    font-size: 2rem;
  }
</style>
