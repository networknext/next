
import emitter from "@/mitt.js";

const update = {

  data() {
    return {
      timer: null,
      updated: false,
      page: 0,
      num_pages: 1,
      left_down: false,
      right_down: false,
    };
  },

  emits: ['update', 'loaded', 'view'],

  mounted: function () {
    this.timer = setInterval(() => { this.update(); this.$emit('update', this.page, this.num_pages) }, 1000)
    document.addEventListener('keydown', this.onKeyDown);
    document.addEventListener('keyup', this.onKeyUp);
    emitter.on('prev_page', () => this.prevPage() )
    emitter.on('next_page', () => this.nextPage() )
    this.$emit('loaded')
  },

  beforeUnmount() {
    clearInterval(this.timer)
    document.removeEventListener('keydown', this.onKeyDown);
    document.removeEventListener('keyup', this.onKeyUp);
  },

  methods: {

    onKeyDown(event) {
      if (event.key == '1') {
        if (!this.left_down) {
          this.prevPage()
          this.left_down = true
        }
      }
      if (event.key == '2') {
        if (!this.right_down) {
          this.nextPage()
          this.right_down = true
        }
      }
    },

    onKeyUp(event) {
      if (event.key == '1') {
        this.left_down = false
      }
      if (event.key == '2') {
        this.right_down = false
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
