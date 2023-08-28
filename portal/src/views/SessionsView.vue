<template>
  <SessionList/>
</template>

<script>

import SessionList from '@/components/SessionList.vue'

export default {
  name: 'App',
  components: {
    SessionList
  }
}
</script>

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

</style>