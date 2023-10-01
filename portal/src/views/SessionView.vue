-left<template>

  <div class="parent">
  
    <div class="bottom">

      <div class="padding"/>

      <div class="left">

        <div id="latency" class="graph"/>
        
        <div id="jitter" class="graph"/>
        
        <div id="packet_loss" class="graph"/>

        <div id="bandwidth" class="graph"/>

      </div>

      <div class="padding"/>

      <div class="right">

        <div class="right-top">

          <div class="top">

            <p><b>Session Id</b></p>
            <p><input type="text" class="text"></p>
            <p><button type="button" class="btn btn-secondary">Search</button></p>

          </div>

          <div class="map"/>

        </div>
  
        <div class="right-bottom">
   
          <div class="session_info">

            <table id="sessions_table" class="table table-striped">
              <tbody>

                <tr>
                  <td class="header">Datacenter</td>
                  <td> {{ this.datacenter_name }} </td>
                </tr>

                <tr>
                  <td class="header">ISP</td>
                  <td> {{ this.isp }} </td>
                </tr>

                <tr>
                  <td class="header">Platform</td>
                  <td> {{ this.platform }} </td>
                </tr>

                <tr>
                  <td class="header">Connection</td>
                  <td> {{ this.connection }} </td>
                </tr>

                <tr>
                  <td class="header">User Hash</td>
                  <td class="fixed"> {{ this.user_hash }} </td>
                </tr>

                <tr>
                  <td class="header">Buyer</td>
                  <td> {{ this.buyer_name }} </td>
                </tr>

              </tbody>
            </table>

          </div>

        </div>

        <div class="route_info">

          <table id="route_table" class="table">

            <p class="header">Current Route</p>
     
            <tbody>

              <tr>
                <td class="left_align header">Client</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
              </tr>

              <tr>
                <td class="left_align header relay_name">akamai.newyork.1.a</td>
                <td class="left_align"> 135.122.10.3:40000 </td>
              </tr>

              <tr>
                <td class="left_align header relay_name">i3d.chicago</td>
                <td class="left_align"> 122.61.5.10:40000 </td>
              </tr>

              <tr>
                <td class="left_align header relay_name">google.iowa.1.a</td>
                <td class="left_align"> 35.22.54.10:40000 </td>
              </tr>

              <tr>
                <td class="left_align header">Server</td>
                <td class="left_align"> 127.0.0.1:50000 </td>
              </tr>

            </tbody>

          </table>

        </div>

        <div class="near_relay_info">

          <p class="header">Near Relays</p>
   
          <table class="table">

            <tbody>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

              <tr>
                <td class="left_align header">i3d.newyork</td>
                <td class="left_align"> 127.0.0.1:30000 </td>
                <td class="left_align"> 5ms </td>
                <td class="left_align"> 2ms </td>
                <td class="left_align"> 0% </td>
              </tr>

            </tbody>

          </table>

        </div>

      </div>

      <div class="padding"/>

    </div>

  </div>

</template>

<script>

import axios from "axios";
import utils from '@/utils.js'
import update from '@/update.js'

import uPlot from "uplot";

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
  width: 2000,
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

let jitter_opts = {
  title: "Jitter",
  width: 2000,
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
  width: 2000,
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
  width: 2000,
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

export default {

  name: "App",

  mixins: [utils, update],

  mounted: function () {
    this.latency = new uPlot(latency_opts, data, document.getElementById('latency'))
    this.jitter = new uPlot(jitter_opts, data, document.getElementById('jitter'))
    this.packet_loss = new uPlot(packet_loss_opts, data, document.getElementById('packet_loss'))
    this.bandwidth = new uPlot(bandwidth_opts, data, document.getElementById('bandwidth'))
  },

  beforeUnmount() {
    this.latency.destroy()
    this.jitter.destroy()
    this.packet_loss.destroy()
    this.bandwidth.destroy()
    this.latency = null
    this.jitter = null
    this.packet_loss = null
    this.bandwidth = null
  },

  methods: {

    async update() {

      console.log('update')

      try {
        const session_id = this.$route.params.id
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/session/' + session_id)
        if (res.data.slice_data !== null) {
          this.datacenter_name = res.data.session_data.datacenter_name
          this.isp = res.data.session_data.isp
          this.buyer_name = res.data.session_data.buyer_name
          this.user_hash = this.parse_uint64(res.data.session_data.user_hash)
          this.platform = 'Linux' // todo
          this.connection = 'Wired' // todo
        }
      } catch (error) {
        console.log(error)
      }
    }
  },

  data() {
    return {
      user_hash: 0,
      buyer_name: '',
      datacenter_name: '',
    };
  },

};

</script>

<style scoped>

.fixed {
  font-family: monospace;
}

.parent {
  background-color: white;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 25px;
}

.top {
  background-color: white;
  width: 100%;
  display: flex;
  flex-direction: row;
  justify-content: left;
  align-content: center;
  align-items: center;
  gap: 15px;
  font-weight: 1;
  font-size: 18px;
  padding: 0px;
}

.bottom {
  background-color: white;
  height: 100%;  
  display: flex;
  flex-direction: row;
}

.left {
  width: 80%;
  height: 100%;
}

.right {
  width: 20%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.right-top {
  display: flex;
  flex-direction: column;
}

.right-bottom {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.padding {
  background-color: white;
  width: 15px;
  height: 100%;
}

.graph {
  background-color: white;
  width: 100%;
  height: 500px;
}

.map {
  background-color: black;
  width: 100%;
  height: 500px;
  flex-shrink: 0;
}

.session_info {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
}

.route_info {
  width: 100%;
  flex-direction: column;
}

.left_align {
  text-align: left;
}

.near_relay_info {
  width: 100%;
}

.relay_name {
}

.text {
  width: 305px; 
  height: 35px;
  font-size: 15px;
  padding: 5px;
}

.header {
  font-weight: bold;
}

button {
  font-size: 15px;
}

</style>
