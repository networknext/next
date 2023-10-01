<template>

  <div class="parent">
    
    <div class="left">

      <div id="latency" class="graph"/>
      
      <div id="jitter" class="graph"/>
      
      <div id="packet_loss" class="graph"/>

    </div>

    <div class="padding"/>

    <div class="right">

      <div class="map"/>

      <div class="session_info">
  
        <p class="session">session id = {{ $route.params.id }}</p>

        <p class="user">user hash = {{ this.user_hash }}</p>

        <p class="slices">slices = {{ this.num_slices }}</p>

        <p class="near_relays">near relays = {{ this.num_near_relays }}</p>

        <p class="buyer_code">buyer code = {{ this.buyer_code }}</p>

        <p class="datacenter_name">datacenter name = {{ this.datacenter_name }}</p>

      </div>

      <div class="near_relay_info">

        <ul>
          <li>near relay 1</li>
          <li>near relay 2</li>
          <li>near relay 3</li>
          <li>near relay 4</li>
          <li>near relay 5</li>
          <li>near relay 6</li>
          <li>near relay 7</li>
          <li>near relay 8</li>
          <li>near relay 9</li>
          <li>near relay 10</li>
          <li>near relay 11</li>
          <li>near relay 12</li>
          <li>near relay 13</li>
          <li>near relay 14</li>
          <li>near relay 15</li>
        </ul>

      </div>

    </div>

    <div class="padding"/>

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

const data = arr;

export default {

  name: "App",

  mixins: [utils, update],

  mounted: function () {
    this.latency = new uPlot(latency_opts, data, document.getElementById('latency'))
    this.jitter = new uPlot(jitter_opts, data, document.getElementById('jitter'))
    this.packet_loss = new uPlot(packet_loss_opts, data, document.getElementById('jitter'))
  },

  methods: {

    async update() {

      console.log('update')

      try {
        const session_id = this.$route.params.id
        const res = await axios.get(process.env.VUE_APP_API_URL + '/portal/session/' + session_id)
        if (res.data.slice_data !== null) {
          this.user_hash = this.parse_uint64(res.data.session_data.user_hash)
          this.buyer_code = res.data.session_data.buyer_code
          this.datacenter_name = res.data.session_data.datacenter_name
          this.num_slices = res.data.slice_data.length
          this.num_near_relays = res.data.near_relay_data[0].num_near_relays
          this.updated = true
        }
      } catch (error) {
        console.log(error)
      }
    }
  },

  data() {
    return {
      fields: ["Session ID", "ISP", "Buyer", "Datacenter", "Server Address", "Direct RTT", "Next RTT", "Improvement"],
      data: [],
      user_hash: 0,
      num_slices: 0,
      num_near_relays: 0,
      buyer_code: '',
      relay_name: '',
      datacenter_name: '',
    };
  },

};

</script>

<style scoped>

session {
  font-family: fixed-width;
}

.parent {
  background-color: #00AAFF;
  width: 100%;
  height: 100%;
  display: flex;
}

.left {
  background-color: rgb(116, 255, 116);
  width: 80%;
  height: 100%;
}

.right {
  background-color: rgb(248, 117, 117);
  width: 20%;
  height: 100%;
}

.padding {
  background-color: white;
  width: 25px;
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
}

.session_info {
  background-color: grey;
  width: 100%;
  height: 500px;
}

.near_relay_info {
  background-color: lightgrey;
  width: 100%;
  height: 500px;
}

</style>
