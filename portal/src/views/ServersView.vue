<template>
<ServerList/>
</template>

<script>

import ServerList from '@/components/ServerList.vue'

export default {
  name: 'App',
  components: {
    ServerList
  }
}
</script>

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

</style>