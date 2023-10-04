
import BigNumber from "bignumber.js";

function parse_uint64(value) {
  const bignum = new BigNumber(value);
  var hex = bignum.toString(16);
  while (hex.length<16) {
    hex = '0' + hex
  }
  return hex
}

function uint64_to_decimal(value) {
  const bignum = new BigNumber(value, 16)
  const decimal = bignum.toString(10)
  return decimal
}

function nice_uptime(value) {
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

function is_visible(element) {
  var style = window.getComputedStyle(element);
  return (style.display !== 'none')
}

function getPlatformName(platformId) {
  switch(platformId) {
  case 1: return "Windows"
  case 2: return "Mac"
  case 3: return "Linux"
  case 4: return "Nintendo Switch"
  case 5: return "PS4"
  case 6: return "iOS"
  case 7: return "Xbox One"
  case 8: return "Xbox Series X"
  case 9: return "PS5"
  default:
    return "Unknown"
  }
}

function getConnectionName(connectionType) {
  switch(connectionType) {
  case 1: return "Wired"
  case 2: return "Wi-Fi"
  case 3: return "Cellular"
  default:
    return "Unknown"
  }
}

export {parse_uint64, uint64_to_decimal, nice_uptime, is_visible, getPlatformName, getConnectionName};