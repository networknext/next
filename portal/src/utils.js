
import BigNumber from "bignumber.js";

const utils = {
  data() {
    return { dummy: 0 }
  },
  methods: {

    parse_uint64(value) {
      const bignum = new BigNumber(value);
      var hex = bignum.toString(16);
      while (hex.length<16) {
        hex = '0' + hex     // todo: this is super lame, need a better solution
      }
      return hex
    },

    uint64_to_decimal(value) {
      const bignum = new BigNumber(value, 16)
      const decimal = bignum.toString(10)
      return decimal
    },

    nice_uptime(value) {
      if (isNaN(value)) {
        return ''
      }
      if (value > 86400) {
        return Math.floor(value/86400) + "d"
      }
      if (value > 3600) {
        return Math.floor(value/3600) + "h"
      }
      if (value > 60) {
        return Math.floor(value/60) + "m"
      }
      return value + "s"
    }
  }
}

export default utils