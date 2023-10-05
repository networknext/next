<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Buyer</p>
      <p class="tight-p test-text"><input id='buyer-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-show="this.found" class="bottom-buyer">

      <div id="left" class="left">

        <div class="d-xxl-none">
          <table id="buyer_table" class="table table-striped" style="vertical-align: middle;">
            <tbody>

              <tr>
                <td class="bold">Total Sessions</td>
                <td> {{ this.data.total_sessions}} </td>
              </tr>

              <tr>
                <td class="bold">Next Sessions</td>
                <td> {{ this.data.next_sessions}} </td>
              </tr>

              <tr>
                <td class="bold">Accelerated</td>
                <td> {{ this.data.accelerated_percent }}% </td>
              </tr>

              <tr>
                <td class="bold">Servers</td>
                <td> {{ this.data.servers }} </td>
              </tr>

            </tbody>
          </table>
        </div>

        <div id="total_sessions" class="graph"/>
        
        <div id="next_sessions" class="graph"/>
        
        <div id="accelerated" class="graph"/>

        <div id="servers" class="graph"/>

      </div>

      <div id="right" class="right d-none d-xxl-block">

        <div class="buyer_info">

          <table id="buyer_table" class="table table-striped" style="vertical-align: middle;">
            <tbody>

              <tr>
                <td class="bold">Total Sessions</td>
                <td> {{ this.data.total_sessions}} </td>
              </tr>

              <tr>
                <td class="bold">Next Sessions</td>
                <td> {{ this.data.next_sessions}} </td>
              </tr>

              <tr>
                <td class="bold">Accelerated</td>
                <td> {{ this.data.accelerated_percent }}% </td>
              </tr>

              <tr>
                <td class="bold">Servers</td>
                <td> {{ this.data.servers }} </td>
              </tr>

              <tr>
                <td class="bold">Live</td>
                <td> {{ this.data.live }} </td>
              </tr>

              <tr>
                <td class="bold">Debug</td>
                <td> {{ this.data.debug }} </td>
              </tr>

            </tbody>
          </table>

        </div>

      </div>

    </div>

  </div>

</template>

<script>

import axios from "axios";
import update from '@/update.js'
import uPlot from "uplot";

import { is_visible } from '@/utils.js'

const arr = [
  [
    1585724400,
    1586156400,
    1586440800,
    1586959200,
    1587452400,
    1587711600,
    1588143600,
    1588600800,
    1588860000,
    1589353200,
    1589785200,
    1590044400,
    1590588000,
    1591020000,
    1591340400,
    1591772400,
    1592204400,
    1592488800,
    1592920800,
    1593180000,
    1593673200,
    1594191600,
    1594648800,
    1594908000,
    1595340000,
    1595833200,
    1596092400,
    1596549600,
    1596808800,
    1597240800,
    1597734000,
    1597993200,
    1598450400,
    1598882400,
    1599141600,
    1599721200,
    1600153200,
    1600437600,
    1600869600,
    1601301600,
    1601622000,
    1602054000,
    1602511200,
    1602770400,
    1603202400,
    1603695600,
    1603954800,
    1604412000,
    1604671200,
    1605103200,
    1605596400,
    1605855600,
    1606287600,
    1606831200,
    1607090400,
    1607583600,
    1608015600,
    1608274800,
    1608732000,
    1609250400,
    1609830000,
    1610089200,
    1610521200,
    1611064800,
    1611324000,
    1611817200,
    1612249200,
    1612508400,
    1612965600,
    1613484000,
    1613977200,
    1614236400,
    1614668400,
    1614952800,
    1615384800,
    1615878000,
    1616137200,
    1616569200,
    1617026400,
    1617285600,
    1617865200,
    1618297200,
    1618556400,
    1619013600,
    1619449200,
    1619712000,
    1620205200,
    1620637200,
    1620972000,
    1621404000,
    1621836000,
    1622120400,
    1622638800,
    1623132000,
    1623391200,
    1623823200,
    1624280400,
    1624539600,
    1625032800,
    1625248800
  ],
  [
    0,
    1.59,
    10.97,
    10.41,
    10.4,
    12,
    8.34,
    11.16,
    14.47,
    14.65,
    14.61,
    14.98,
    17.08,
    15.94,
    13.88,
    11.07,
    13.41,
    14.3,
    21.64,
    15.8,
    24.42,
    24.63,
    23.65,
    24.4,
    25.03,
    15.07,
    5.21,
    16.4,
    17.51,
    19.66,
    28.19,
    19.21,
    18.51,
    18.47,
    18.09,
    18.83,
    19.24,
    17.51,
    18.35,
    19.15,
    18.61,
    18.72,
    19.76,
    18.76,
    18.66,
    19.45,
    20.37,
    20.98,
    21.09,
    21.66,
    21.86,
    21.93,
    22.45,
    22.34,
    21.33,
    21.21,
    21.08,
    22.18,
    22.19,
    22.88,
    22.81,
    23.31,
    23.72,
    23.47,
    24.47,
    24.38,
    23.25,
    27.07,
    27.55,
    30.03,
    28.1,
    30.6,
    31.18,
    24.95,
    31.62,
    35.54,
    34.65,
    34.45,
    35.1,
    35.65,
    36.38,
    35.87,
    36.49,
    35.65,
    37.81,
    38.15,
    36.13,
    36.46,
    32.81,
    34.92,
    37.28,
    38.2,
    40.38,
    40.08,
    39.98,
    39.35,
    37.98,
    41.13,
    42.74,
    42.177
  ]
];

let total_sessions_opts = {
  title: "Total Sessions",
  width: 0,
  height: 450,
  legend: {
    show: false
  },
  series: [
    {},
    {
      stroke: "green",
      fill: "rgba(100,100,100,0.1)"
    }
  ],
  axes: [
    {},
    {
      side: 1
    }
  ]
};

let next_sessions_opts = {
  title: "Next Sessions",
  width: 0,
  height: 450,
  legend: {
    show: false
  },
  series: [
    {},
    {
      stroke: "green",
      fill: "rgba(100,100,100,0.1)"
    }
  ],
  axes: [
    {},
    {
      side: 1
    }
  ]
};

let accelerated_opts = {
  title: "Accelerated",
  width: 0,
  height: 450,
  legend: {
    show: false
  },
  series: [
    {},
    {
      stroke: "green",
      fill: "rgba(100,100,100,0.1)"
    }
  ],
  axes: [
    {},
    {
      side: 1
    }
  ]
};

let servers_opts = {
  title: "Servers",
  width: 0,
  height: 450,
  legend: {
    show: false
  },
  series: [
    {},
    {
      stroke: "green",
      fill: "rgba(100,100,100,0.1)"
    }
  ],
  axes: [
    {},
    {
      side: 1
    }
  ]
};

const data = arr;

async function getData(page, buyer) {
  try {
    if (page == null) {
      page = 0
    }
    const url = process.env.VUE_APP_API_URL + '/portal/buyer/' + buyer
    const res = await axios.get(url);
    let data = {}
    data['buyer'] = buyer
    if (res.data.buyer_data !== null) {
      data['live'] = res.data.buyer_data.live
      data['debug'] = res.data.buyer_data.debug
      data['total_sessions'] = res.data.buyer_data.total_sessions
      data['next_sessions'] = res.data.buyer_data.next_sessions
      data['accelerated_percent'] = '0.0'
      if (res.data.buyer_data.total_sessions > 0) {
        data['accelerated_percent'] = ((res.data.buyer_data.next_sessions / res.data.buyer_data.total_sessions)*100.0).toFixed(1)
      }
      data["found"] = true
    }
    
    /*
    // todo: this is mocked
    data['total_sessions'] = 1000
    data['next_sessions'] = 500
    data['accelerated'] = 50
    data['servers'] = 8
    data['live'] = true
    data['debug'] = true
    data['found'] = true
    */
    return [data, 0, 1]
  } catch (error) {
    console.log(error);
    let data = {}
    data['buyer'] = buyer
    data['found'] = false
    return [data, 0, 1]
  }
}

export default {

  name: "App",

  mixins: [update],

  data() {
    return {
      data: {},
      found: false,
      observer: null,
      prevWidth: 0,
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let buyer = values[values.length-1]
    let result = await getData(0, buyer)
    next(vm => {
      if (result != null && !result.error) {
        vm.data = result[0]
        vm.page = result[1]
        vm.num_pages = result[2]
        vm.$emit('notify-update', vm.page, vm.num_pages)
        vm.found = result[0]['found']
      }
    })
  },

  mounted: function () {
  
    this.total_sessions = new uPlot(total_sessions_opts, data, document.getElementById('total_sessions'))
    this.next_sessions = new uPlot(next_sessions_opts, data, document.getElementById('next_sessions'))
    this.accelerated = new uPlot(accelerated_opts, data, document.getElementById('accelerated'))
    this.servers = new uPlot(servers_opts, data, document.getElementById('servers'))

    this.observer = new ResizeObserver(this.resize)
    this.observer.observe(document.body, {box: 'border-box'})

    document.getElementById("buyer-input").value = document.getElementById("buyer-input").defaultValue = this.data['buyer']
    document.getElementById("buyer-input").addEventListener('keyup', this.onKeyUp);

    this.$emit('notify-view', 'buyer')
  },

  beforeUnmount() {
    this.total_sessions.destroy()
    this.next_sessions.destroy()
    this.accelerated.destroy()
    this.servers.destroy()
    this.observer.disconnect()
    this.prevWidth = 0
    this.total_sessions = null
    this.next_sessions = null
    this.accelerated = null
    this.servers = null
    this.observer = null
  },

  methods: {

    resize() {
      const width = document.body.clientWidth;
      if (width !== this.prevWidth) {
        this.prevWidth = width;
        if (this.total_sessions) {
          let graph_width = width
          if (is_visible(document.getElementById('right'))) {
            graph_width -= 550
          } else {
            graph_width -= 30
          }
          let graph_height = graph_width * 0.333
          if (graph_height > 450) {
            graph_height = 450
          } else if (graph_height < 250) {
            graph_height = 250
          }
          this.total_sessions.setSize({width: graph_width, height: graph_height})
          this.next_sessions.setSize({width: graph_width, height: graph_height})
          this.accelerated.setSize({width: graph_width, height: graph_height})
          this.servers.setSize({width: graph_width, height: graph_height})
        }
      }    
    },

    async getData(page, buyer) {
      if (buyer == null) {
        buyer = this.$route.params.id
      }
      return getData(page, buyer)
    },

    async update() {
      let result = await getData(this.page, this.$route.params.id)
      if (result != null) {
        this.data = result[0]
        this.page = result[1]
        this.num_pages = result[2]
        this.found = result[0]['found']
        this.$emit('notify-update', this.page, this.num_pages)
      }
    },

    search() {
      const buyer = document.getElementById("buyer-input").value
      this.$router.push('/buyer/' + buyer)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

  },

};

</script>

<style scoped>

.fixed {
  font-family: monospace;
}

.parent {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 15px;
  padding-top: 20px;
}

.bottom-buyer {
  height: 100%;  
  display: flex;
  flex-direction: row;
  padding: 0px;
  justify-content: space-between;
  gap: 15px;
}

.left {
  width: 100%;
  height: 100%;
  padding: 0px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 25px;
  padding-top: 5px;
}

.graph {
  width: 100%;
  height: 100%;
}

.right {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 0px;
  max-width: 500px;
  min-width: 300px;
}

.search {
  width: 100%;
  height: 35px;
  display: flex;
  flex-direction: row;
  align-items: center;
  align-content: center;
  justify-content: space-between;
  gap: 15px;
  font-weight: 1;
  font-size: 18px;
  padding: 0px;
}

.text {
  width: 100%;
  height: 35px;
  font-size: 15px;
  padding-left: 5px;
}

.test-text {
  width: 100%;
}

.right-top {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.right-bottom {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 50px;
}

.map {
  background-color: #555555;
  width: 100%;
  height: 500px;
  flex-shrink: 0;
}

.buyer_info {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
}

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
}

.near_relay_info {
  width: 100%;
}

.header {
  font-weight: bold;
  font-size: 18px;
}

.bold {
  font-weight: bold;
}

button {
  font-size: 15px;
}

.tight-p {
  line-height: 15px;
  margin-bottom: 2px;
}

a {
  color: #2c3e50;
  text-decoration: none;
}

</style>
