<template>
  <div :class="[{'alert': true}, className ? className : 'alert-secondary']" role="alert">
    {{ alertMessage }}
  </div>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'

@Component
export default class Alert extends Vue {
  @Prop({ required: false, type: String, default: '' }) message!: string
  @Prop({ required: false, type: String, default: '' }) classType!: string

  get alertMessage (): string {
    return this.message
  }

  get className (): string {
    return this.classType
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  div {
    text-align: center;
  }
</style>
