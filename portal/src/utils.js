
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
  case 1: return "PC"
  case 2: return "Mac"
  case 3: return "Linux"
  case 4: return "Switch"
  case 5: return "PS4"
  case 6: return "iOS"
  case 7: return "Xbox One"
  case 8: return "Series X"
  case 9: return "PS5"
  default:
    return "Unknown"
  }
}

function getConnectionName(connectionType) {
  switch(connectionType) {
  case 1: return "Wired"
  case 2: return "Wi-Fi"
  case 3: return "Cell"
  default:
    return "Unknown"
  }
}

function getCountryName(countryCode) {
  if (countryCode=="" || countryCode==null) {
    return "Unknown"
  }
  try {
    const regionNames = new Intl.DisplayNames(['en'], {type: 'region'});
    return regionNames.of(countryCode);
  } catch (error) {
    return countryCode
  }
}

function getAcceleratedPercent(nextSessions, totalSessions) {
  let acceleratedPercent = 0.0
  if (totalSessions > 0) {
    acceleratedPercent = ((nextSessions / totalSessions)*100.0).toFixed(1)
  }
  return acceleratedPercent
}

function custom_graph(config) {

  let percent = config.percent

  let opts = {
    title: config.title,
    width: 0,
    height: 0,
    legend: {
      show: true,
    },
    cursor: {
      drag: {
        x: false,
        y: false,
      }
    },
    series: [
      {
        value: (self, v) => {
          if (v != null) {
            return new Date(v*1000).toLocaleString()
          } else if (self._data[0] != null && self._data[0].length > 0) {
            return new Date((self._data[0][self._data[0].length-1])*1000).toLocaleString()
          } else {
            return '--'
          }
        }
      }
    ],
    axes: [
      {
        space: 60,
        incrs: [
           // minute divisors (# of secs)
           5,
           10,
           20,
           30,
           // hour divisors
           60,
           60 * 10,
           60 * 20,
           60 * 30,
           // day divisors
           60 * 60,
        ],
        values: [
          // tick incr        default           year                           month    day                      hour     min                sec       mode
          [3600 * 24 * 365,   "{YYYY}",         null,                          null,    null,                    null,    null,              null,        1],
          [3600 * 24 * 28,    "{MMM}",          "\n{YYYY}",                    null,    null,                    null,    null,              null,        1],
          [3600 * 24,         "{M}/{D}",        "\n{YYYY}",                    null,    null,                    null,    null,              null,        1],
          [3600,              "{h}{aa}",        "\n{M}/{D}/{YY}",              null,    "\n{M}/{D}",             null,    null,              null,        1],
          [60,                "{h}:{mm}{aa}",   "\n{M}/{D}/{YY}",              null,    "\n{M}/{D}",             null,    null,              null,        1],
          [10,                "",               "{h}:{mm}{aa}\n{M}/{D}/{YY}",  null,    "{h}:{mm}{aa}",          null,    "{h}:{mm}{aa}",    null,        1],
        ],
      },
      {
        side: 1,
      }
    ],
    scales: {
      y: {
        range: (u, dataMin, dataMax) => {
          if (percent) {
            return [0, 100];
          } else {
            return [0, Math.max(dataMax*1.1, 1)]
          }
        }
      },
    },
  };

  let i = 0
  while (i < config.series.length) {
    let units = config.series[i].units
    let index = i + 1
    opts.series.push({
      stroke: config.series[i].stroke,
      fill: config.series[i].fill,
      width: 2,
      label: config.series[i].name,
      points: {
        show: (self, si) => {
          let element = document.getElementById('right')
          if (element != null && is_visible(element)) {
            return self.data[si].length < 50
          } else {
            return false
          }
        }
      },
      value: (self, v) => {
        if (v != null) {
          return v.toLocaleString() + units
        } else if (self._data[index] != null && self._data[index].length > 0) {
          return (self._data[index][self._data[index].length-1]).toLocaleString() + units
        } else {
          return '--'
        }
      }
    })
    i++
  }
  return opts
}

export {parse_uint64, uint64_to_decimal, nice_uptime, is_visible, getPlatformName, getConnectionName, getCountryName, getAcceleratedPercent, custom_graph};
