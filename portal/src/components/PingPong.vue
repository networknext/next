<template>
  <div class="session_counts">
    <h1>Ping</h1>
  </div>
  <ul>
    <li v-for="item of items" :key="item.id">
      {{ item }}
    </li>
  </ul>
</template>

<script>
import axios from "axios";
export default {
  name: "App",
  data() {
    return {
      items: [],
    };
  },
  async created() {
    try {
      const res = await axios.get('http://dev.virtualgo.net/ping');
      this.items = []
      this.items.push(res.data);
    } catch (error) {
      console.log(error);
    }
  },
};
</script>

<style scoped>

h1 {
  margin: 75px 0 25px;
  color: #666666;
}

ul {
  list-style-type: none;
  padding: 0;
  width: 60%;
  text-align: center;
  margin: auto;
}

li {
  display: inline-block;
  margin: 0 10px;
}

</style>
