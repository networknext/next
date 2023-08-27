
const update = {

  data() {
    return {
      timer: null,
      updated: false
    };
  },

  mounted: function () {
    this.timer = setInterval(() => { this.update(); this.updated = true }, 1000)
    this.update()
  },

  beforeUnmount() {
    clearInterval(this.timer)
  }
}

export default update
