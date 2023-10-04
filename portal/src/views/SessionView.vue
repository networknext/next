<template>

  <div class="parent" id="parent">
  
    <div class="search">

      <p class="tight-p">Session</p>
      <p class="tight-p test-text"><input id='session-id-input' type="text fixed" class="text"></p>
      <p class="tight-p"><button type="button" class="btn btn-secondary" id='search-button' @click="this.search()">Search</button></p>

    </div>

    <div v-show="this.found" class="bottom">

      <div id="left" class="left">

        <div class="d-xxl-none session-info-mobile">
          <table id="session_table" class="table table-striped" style="vertical-align: middle; padding: 15px;">
            <tbody>

              <tr>
                <td class="bold">Datacenter</td>
                <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">Server</td>
                <td> <router-link :to="'/server/' + this.data['server_address']"> {{ this.data['server_address'] }} </router-link> </td>
              </tr>

              <tr>
                <td class="bold">ISP</td>
                <td> {{ this.data['isp'] }} </td>
              </tr>

              <tr>
                <td class="bold">Platform</td>
                <td> {{ this.data['platform'] }} </td>
              </tr>

              <tr>
                <td class="bold">Connection</td>
                <td> {{ this.data['connection'] }} </td>
              </tr>

              <tr>
                <td class="bold">User Hash</td>
                <td class="fixed"> <router-link :to="'/user/' + this.data['user_hash']"> {{ this.data['user_hash'] }} </router-link> </td>
              </tr>

            </tbody>
          </table>
        </div>

        <div id="latency_graph" class="graph"/>
        
        <div id="jitter_graph" class="graph"/>
        
        <div id="packet_loss_graph" class="graph"/>

        <div id="out_of_order_graph" class="graph"/>

        <div id="bandwidth_graph" class="graph"/>

        <div class="d-xxl-none">

          <p class="header" style="padding-top: 15px; padding-bottom: 5px">Route</p>
   
          <table id="route_table" class="table" v-if="this.data['route_relays'] != null && this.data['route_relays'].length > 0">

            <tbody>

              <tr>
                <td class="left_align bold"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr v-for="item in this.data['route_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="right_align"> {{ item.address }} </td>
              </tr>

              <tr>
                <td class="left_align bold"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

          <table id="route_table" class="table" v-else>

            <tbody>

              <tr>
                <td class="left_align bold"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr>
                <td class="left_align bold"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

          <p class="header" style="padding-top: 25px; padding-bottom: 15px">Near Relays</p>
   
          <table class="table">

            <tbody>

              <tr v-for="item in this.data['near_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="left_align"> {{ item.rtt }}ms </td>
                <td class="left_align"> {{ item.jitter }}ms </td>
                <td class="left_align"> {{ item.packet_loss}}% </td>
              </tr>

            </tbody>

          </table>

        </div>

      </div>

      <div id="right" class="right d-none d-xxl-block">

        <div class="right-top">

          <div class="map"/>

        </div>
  
        <div class="right-bottom">
   
          <div class="session_info">

            <table id="session_table" class="table table-striped" style="vertical-align: middle;">
              <tbody>

                <tr>
                  <td class="bold">Datacenter</td>
                  <td> <router-link :to="'/datacenter/' + this.data['datacenter_name']"> {{ this.data['datacenter_name'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Server</td>
                  <td> <router-link :to="'/server/' + this.data['server_address']"> {{ this.data['server_address'] }} </router-link> </td>
                </tr>
                
                <tr>
                  <td class="bold">ISP</td>
                  <td> {{ this.data['isp'] }} </td>
                </tr>

                <tr>
                  <td class="bold">Platform</td>
                  <td> {{ this.data['platform'] }} </td>
                </tr>

                <tr>
                  <td class="bold">Connection</td>
                  <td> {{ this.data['connection'] }} </td>
                </tr>

                <tr>
                  <td class="bold">User Hash</td>
                  <td class="fixed"> <router-link :to="'/user/' + this.data['user_hash']"> {{ this.data['user_hash'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Buyer</td>
                  <td> <router-link :to="'/buyer/' + this.data['buyer_code']"> {{ this.data['buyer_name'] }} </router-link> </td>
                </tr>

                <tr>
                  <td class="bold">Start Time</td>
                  <td> {{ this.data['start_time'] }} </td>
                </tr>

              </tbody>
            </table>

          </div>

        </div>

        <div class="route_info">

          <p class="bold tight-p">Route</p>
   
          <table id="route_table" class="table" v-if="this.data['route_relays'] != null && this.data['route_relays'].length > 0">

            <tbody>

              <tr>
                <td class="left_align bold"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr v-for="item in this.data['route_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="right_align"> {{ item.address }} </td>
              </tr>

              <tr>
                <td class="left_align bold"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

          <table id="route_table" class="table" v-else>

            <tbody>

              <tr>
                <td class="left_align bold"> <router-link :to="'/user/' + this.data['user_hash']"> Client </router-link></td>
                <td class="right_align"> </td>
              </tr>

              <tr>
                <td class="left_align bold"> <router-link :to="'/server/' + this.data['server_address']"> Server </router-link> </td>
                <td class="right_align"> {{ this.data['server_address'] }} </td>
              </tr>

            </tbody>

          </table>

        </div>

        <div class="near_relay_info">

          <p class="bold">Near Relays</p>
   
          <table class="table">

            <tbody>

              <tr v-for="item in this.data['near_relays']" :key="item.id">
                <td class="left_align bold"> <router-link :to="'/relay/' + item.name"> {{ item.name }} </router-link> </td>
                <td class="left_align"> {{ item.rtt }}ms </td>
                <td class="left_align"> {{ item.jitter }}ms </td>
                <td class="left_align"> {{ item.packet_loss}}% </td>
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

import {parse_uint64, is_visible} from '@/utils.js'

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

let latency_opts = {
  title: "Latency",
  titleFont: "Arial",
  width: 0,
  height: 450,
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
    },
    {
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      width: 2,
      label: "Direct",
      points: {
        show: (self, si) => {
          if (is_visible(document.getElementById('right'))) {
            return self.series[si].width < 100
          } else {
            return false
          }
        }
      },
      value: (self, v) => {
        if (v != null) {
          return v + "ms"
        } else if (self._data[1] != null && self._data[1].length > 0) {
          return self._data[1][self._data[1].length-1] + 'ms'
        } else {
          return '--'
        }
      }
    },
    {
      stroke: "#11AA44",
      fill: "rgba(10,100,10,0.1)",
      width: 2,
      label: "Next",
      points: {
        show: (self, si) => {
          if (is_visible(document.getElementById('right'))) {
            return self.series[si].width < 100
          } else {
            return false
          }
        }
      },
      value: (self, v) => {
        if (v != null) {
          return v + "ms"
        } else if (self._data[2] != null && self._data[2].length > 0) {
          return self._data[2][self._data[2].length-1] + 'ms'
        } else {
          return '--'
        }
      }
    },
    {
      stroke: "orange",
      fill: "rgba(100,100,100,0.1)",
      width: 2,
      label: "Predicted",
      points: {
        show: (self, si) => {
          if (is_visible(document.getElementById('right'))) {
            return self.series[si].width < 100
          } else {
            return false
          }
        }
      },
      value: (self, v) => {
        if (v != null) {
          return v + "ms"
        } else if (self._data[3] != null && self._data[3].length > 0) {
          return self._data[3][self._data[3].length-1] + 'ms'
        } else {
          return '--'
        }
      }
    }
  ],
  axes: [
    {
      space: 40,
      incrs: [
         // minute divisors (# of secs)
         10,
         20,
         30,
         // hour divisors
         60,
         60 * 5,
         60 * 10,
         60 * 15,
         60 * 30,
         // day divisors
         3600,
      ],
      values: [
        // tick incr        default           year                             month    day                        hour     min                sec       mode
        [3600 * 24 * 365,   "{YYYY}",         null,                            null,    null,                      null,    null,              null,        1],
        [3600 * 24 * 28,    "{MMM}",          "\n{YYYY}",                      null,    null,                      null,    null,              null,        1],
        [3600 * 24,         "{M}/{D}",        "\n{YYYY}",                      null,    null,                      null,    null,              null,        1],
        [3600,              "{h}{aa}",        "\n{M}/{D}/{YY}",                null,    "\n{M}/{D}",               null,    null,              null,        1],
        [60,                "{h}:{mm}{aa}",   "\n{M}/{D}/{YY}",                null,    "\n{M}/{D}",               null,    null,              null,        1],
        [10,                "",               "{h}:{mm}{aa}\n{M}/{D}/{YY}",     null,    "{h}:{mm}{aa}\n{M}/{D}",  null,    "{h}:{mm}{aa}",    null,        1],
      ],
    },
    {
      side: 1,
    }
  ]
};

let jitter_opts = {
  title: "Jitter",
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

let packet_loss_opts = {
  title: "Packet Loss",
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

let out_of_order_opts = {
  title: "Out of Order",
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

let bandwidth_opts = {
  title: "Bandwidth",
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

async function getData(page, session_id) {
  try {

    if (page == null) {
      page = 0
    }
  
    const url = process.env.VUE_APP_API_URL + '/portal/session/' + session_id
  
    const res = await axios.get(url);
  
    let data = {}
  
    if (res.data.slice_data !== null) {

      // get session data

      let session_data = res.data.session_data

      data['session_id'] = parse_uint64(session_data.session_id)
      data["datacenter_name"] = session_data.datacenter_name
      data["isp"] = session_data.isp
      data["buyer_code"] = session_data.buyer_code
      data["buyer_name"] = session_data.buyer_name
      data["user_hash"] = session_data.user_hash
      data["platform"] = getPlatformName(session_data.platform_type)
      data["connection"] = getConnectionName(session_data.connection_type)
      data["start_time"] = new Date(parseInt(session_data.start_time)).toLocaleString()
      data["server_address"] = session_data.server_address
    
      // route relays

      if (session_data.num_route_relays > 0) {
        let i = 0
        let route_relays = []
        while (i < session_data.num_route_relays) {
          route_relays.push({
            id:        session_data.route_relay_ids[i],
            name:      session_data.route_relay_names[i],
            address:   session_data.route_relay_addresses[i],
          })
          i++
        }
        data['route_relays'] = route_relays
      }

      // near relays
  
      let near_relay_data = res.data.near_relay_data
      if (near_relay_data.length > 0) {
        near_relay_data = near_relay_data[near_relay_data.length-1]
        let i = 0
        let near_relays = []
        while (i < near_relay_data.num_near_relays) {
          if (near_relay_data.near_relay_rtt[i] != 0) {
            near_relays.push({
              id:          near_relay_data.near_relay_id[i],
              name:        near_relay_data.near_relay_name[i],
              rtt:         near_relay_data.near_relay_rtt[i],
              jitter:      near_relay_data.near_relay_jitter[i],
              packet_loss: near_relay_data.near_relay_packet_loss[i],
            })
          }
          i++
        }
        near_relays.sort( function(a,b) {
          if (a.name < b.name) {
            return -1
          }
          if (a.name > b.name) {
            return +1
          }
          return 0
        })
        data['near_relays'] = near_relays
      }

      // latency graph data
  
      let latency_timestamp = []
      let latency_direct = []
      let latency_next = []
      let latency_predicted = []
      let i = 0
      while (i < res.data.slice_data.length) {
        const timestamp = parseInt(res.data.slice_data[i].timestamp)
        latency_timestamp.push(timestamp)
        latency_direct.push(res.data.slice_data[i].direct_rtt)
        latency_next.push(res.data.slice_data[i].next_rtt)
        latency_predicted.push(res.data.slice_data[i].predicted_rtt)
        i++
      }

      data.latency_data = [latency_timestamp, latency_direct, latency_next, latency_predicted]

      // mark data as found

      data["found"] = true
    }

    return [data, 0, 1]

  } catch (error) {
    
    // error

    console.log(error);
    
    let data = {}
    data['session_id'] = session_id
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
      show_legend: false,
    };
  },

  async beforeRouteEnter (to, from, next) {
    let values = to.path.split("/")
    let session_id = values[values.length-1]
    let result = await getData(0, session_id)
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
  
    this.latency = new uPlot(latency_opts, [[],[],[]], document.getElementById('latency_graph'))
    this.jitter = new uPlot(jitter_opts, data, document.getElementById('jitter_graph'))
    this.packet_loss = new uPlot(packet_loss_opts, data, document.getElementById('packet_loss_graph'))
    this.out_of_order = new uPlot(out_of_order_opts, data, document.getElementById('out_of_order_graph'))
    this.bandwidth = new uPlot(bandwidth_opts, data, document.getElementById('bandwidth_graph'))

    this.observer = new ResizeObserver(this.resize)
    this.observer.observe(document.body, {box: 'border-box'})

    document.getElementById("session-id-input").value = document.getElementById("session-id-input").defaultValue = this.data['session_id']
    document.getElementById("session-id-input").addEventListener('keyup', this.onKeyUp);

    this.$emit('notify-view', 'session')

    this.updateGraphs()
  },

  beforeUnmount() {
    this.latency.destroy()
    this.jitter.destroy()
    this.packet_loss.destroy()
    this.out_of_order.destroy()
    this.bandwidth.destroy()
    this.observer.disconnect()
    this.prevWidth = 0
    this.latency = null
    this.jitter = null
    this.packet_loss = null
    this.out_of_order = null
    this.bandwidth = null
    this.observer = null
  },

  methods: {

    resize() {
      const right_visible = is_visible(document.getElementById('right'))
      const width = document.body.clientWidth;
      if (width !== this.prevWidth) {

        // resize the graphs to match the page width

        this.prevWidth = width;
        if (this.latency) {
          let graph_width = width
          if (right_visible) {
            graph_width -= 540
          } else {
            graph_width -= 10
          }
          let graph_height = graph_width * 0.4
          if (graph_height > 500) {
            graph_height = 500
          } else if (graph_height < 250) {
            graph_height = 250
          }
          this.latency.setSize({width: graph_width, height: graph_height})
          this.jitter.setSize({width: graph_width, height: graph_height})
          this.packet_loss.setSize({width: graph_width, height: graph_height})
          this.out_of_order.setSize({width: graph_width, height: graph_height})
          this.bandwidth.setSize({width: graph_width, height: graph_height})
        }
      }    

      // show legends in desktop, hide them in mobile layout

      this.show_legend = right_visible
      var elements = document.getElementsByClassName('u-legend');
      let i = 0;
      while (i < elements.length) {
        if (this.show_legend) {
          elements[i].style.display = 'block';
        } else {
          elements[i].style.display = 'none';
        }
        i++;
      }
    },

    async getData(page, session_id) {
      if (session_id == null) {
        session_id = this.$route.params.id
      }
      let data = getData(page, session_id)
      this.updateGraphs()
      return data
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
      this.updateGraphs()
    },

    updateGraphs() {
      if (this.latency != null && this.data.latency_data != null) {
        this.latency.setData(this.data.latency_data, true)
      }
    },

    search() {
      const session_id = document.getElementById("session-id-input").value
      this.$router.push('/session/' + session_id)
    },

    onKeyUp(event) {
      if (event.key == 'Enter') {
        this.search()
      }
    },

  },

};

</script>

<style>

.fixed {
  font-family: monospace;
}

.parent {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding-top: 20px;
}

.bottom {
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
  gap: v-bind("show_legend ? '35px' : '20px'");
  padding-top: v-bind("show_legend ? '10px' : '0px'");
}

.graph {
  height: 100%;
  overflow: visible;
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
  padding-right: v-bind("show_legend ? '15px' : '0px'");
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
  padding-left: 15px;
  padding-right: 15px;
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

.session_info {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  padding-top: 15px;
}

.route_info {
  width: 100%;
  flex-direction: column;
  justify-content: space-between;
  padding-top: 25px;
}

.near_relay_info {
  width: 100%;
  flex-direction: column;
  justify-content: space-between;
  padding-top: 25px;
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

.u-title {
  font-family: "Montserrat";
}

.session-info-mobile {
  padding-left: 15px;
  padding-right: 15px;
}

.left_align {
  text-align: left;
}

.right_align {
  text-align: right;
}

</style>
