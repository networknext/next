<template>
<DatacenterList/>
</template>

<script>

import DatacenterList from '@/components/DatacenterList.vue'

export default {
  name: 'App',
  components: {
    DatacenterList
  }
}
</script>

<style lang="scss">

nav a {
  font-weight: bold;
  color: #2c3e50;
}

header {
  margin: 75px 0 25px;
  color: #666666;
}

</style>