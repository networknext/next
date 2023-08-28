<template>
<SellerList/>
</template>

<script>

import SellerList from '@/components/SellerList.vue'

export default {
  name: 'App',
  components: {
    SellerList,
  }
}
</script>

<style lang="scss">

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

</style>
