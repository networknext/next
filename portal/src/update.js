
const update = {

  data() {
    return {
      timer: null,
      updated: false,
    };
  },

  created: function() {
    this.update()
  },

  mounted: function () {
    this.timer = setInterval(() => { this.update() }, 10000)
  },

  beforeUnmount() {
    clearInterval(this.timer)
  }
}

export default update
