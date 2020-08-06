<template>
  <div :class="[{'alert': true}, className ? className : 'alert-secondary']" role="alert">
    {{ alertMessage }}
    <slot></slot>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'

// TODO: Add helper function that make it easier to set the message and alert type - It is kind of a pain to deal with when there are multiple alerts on the page
@Component
export default class Alert extends Vue {
  @Prop({ required: false, type: String, default: '' }) message!: string
  @Prop({ required: false, type: String, default: '' }) alertType!: string

  get alertMessage (): string {
    return this.message
  }

  get className (): string {
    return this.alertType
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  div {
    text-align: center;
  }
  .alert {
    margin-bottom: 0rem;
  }
</style>
