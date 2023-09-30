
import emitter from "@/mitt.js";

const update = {

  data() {
    return {
      timer: null,
      updated: false,
      page: 0,
      num_pages: 1,
    };
  },

  emits: ['update'],

  mounted: function () {
    this.timer = setInterval(() => { this.update(); this.$emit('update', this.page, this.num_pages) }, 1000)
    document.addEventListener('keydown', this.onKeyPress);
    emitter.on('prev_page', () => this.prevPage() )
    emitter.on('next_page', () => this.nextPage() )
  },

  beforeUnmount() {
    clearInterval(this.timer)
    document.removeEventListener('keydown', this.onKeyPress);
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
      if (result != null) {
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
      }
      this.$emit('update', this.page, this.num_pages)
    },

    async nextPage() {
      if (this.page < this.num_pages-1) {
        this.page++
        let result = await this.getData(this.page, this.param)
        if (result != null) {
          this.data = result[0]
          this.page = result[1]
          this.num_pages = result[2]
        }
      }
      this.$emit('update', this.page, this.num_pages)
    },

    async prevPage() {
      if (this.page > 0) {
        this.page--
        let result = await this.getData(this.page, this.param)
        if (result != null) {
          this.data = result[0]
          this.page = result[1]
          this.num_pages = result[2]
        }
      }
      this.$emit('update', this.page, this.num_pages)
    }
  }
}

export default update

