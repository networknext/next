
const update = {

  data() {
    return {
      timer: null,
      updated: false,
      page: 0,
      num_pages: 1,
    };
  },

  mounted: function () {
    this.timer = setInterval(() => { this.update() }, 1000)
    document.addEventListener('keypress', this.onKeyPress);
  },

  beforeUnmount() {
    clearInterval(this.timer)
    document.removeEventListener('keypress', this.onKeyPress);
  },

  methods: {

    onKeyPress(event) {
      if (event.key == '1') {
        this.prevPage()
      }
      if (event.key == '2') {
        this.nextPage()
      }
    },

    async getData() {
      return [null, 0, 1]
    },

    async setPage(page) {
      this.page = page
      let result = await this.getData(this.page)
      this.data = result[0]
      this.page = result[1]
      this.num_pages = result[2]
    },

    async nextPage() {
      console.log("next page " + this.page)
      if (this.page < this.num_pages-1) {
        this.page++
        let result = await this.getData(this.page)
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
      }
    },

    async prevPage() {
      console.log("prev page" + this.page)
      if (this.page > 0) {
        this.page--
        let result = await this.getData(this.page)
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
      }
    }
  }
}

export default update
