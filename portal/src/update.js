
const update = {

  data() {
    return {
      timer: null,
      updated: false,
    };
  },

  mounted: function () {
    this.timer = setInterval(() => { this.update() }, 1000)
  },

  beforeUnmount() {
    clearInterval(this.timer)
  }
}

export default update
