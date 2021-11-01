<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-body">
            <h5 class="card-title">
              Session IDs by location
            </h5>
            <div class="table-responsive table-no-top-line table-wrapper-scroll-y my-custom-scrollbar">
              <table class="table table-sm">
                <thead>
                  <tr>
                    <th>
                      <span>
                        <!-- No Title -->
                      </span>
                    </th>
                    <th>
                      <span>
                        Session ID
                      </span>
                    </th>
                    <th>
                      <span>
                        Latitude
                      </span>
                    </th>
                    <th>
                      <span>
                        Longitude
                      </span>
                    </th>
                  </tr>
                </thead>
                <tbody v-if="points.length === 0">
                  <tr>
                    <td colspan="7" class="text-muted">
                        No points could be found at this zoom level. Please zoom in further for better results
                    </td>
                  </tr>
                </tbody>
                <tbody>
                  <tr v-for="(point, index) in points" :key="index">
                    <td>
                      <font-awesome-icon
                        id="status"
                        icon="circle"
                        class="fa-w-16 fa-fw"
                        :class="{
                          'text-success': point.source[2],
                          'text-primary': !point.source[2]
                        }"
                      />
                    </td>
                    <td v-if="point.source[3]">
                      <router-link
                        :to="`/session-tool/${point.source[3]}`"
                        class="text-dark fixed-width"
                      >{{ point.source[3] }}</router-link>
                    </td>
                    <td v-if="!point.source[3]">
                      Unavailable
                    </td>
                    <td>{{ point.source[1] }}</td>
                    <td>{{ point.source[0] }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="card-footer">
            <button class="btn btn-xs btn-primary modal-default-button" style="max-height: 40px; width: 100px;" @click="$root.$emit('hideMapPointsModal')">
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'

/**
 * This component opens up a modal for picking from nested map points
 */

/**
 * TODO: Clean up template
 */

@Component
export default class MapPointsModal extends Vue {
  @Prop({ default: () => { return [] } }) readonly points!: Array<any>
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .modal-mask {
    position: fixed;
    z-index: 9998;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: table;
    transition: opacity 0.3s ease;
  }

  .modal-wrapper {
    display: table-cell;
    vertical-align: middle;
  }

  .modal-container {
    max-width: 800px;
    margin: 0px auto;
    background-color: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
    transition: all 0.3s ease;
    font-family: Helvetica, Arial, sans-serif;
  }

  .modal-header h3 {
    margin-top: 0;
    color: #42b983;
  }

  .modal-body {
    margin: 20px 0;
  }

  .modal-default-button {
    float: right;
    border-color: #009FDF;
    background-color: #009FDF;
  }

  .modal-default-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }

  /*
  * The following styles are auto-applied to elements with
  * transition="modal" when their visibility is toggled
  * by Vue.js.
  *
  * You can easily play with the modal transition by editing
  * these styles.
  */

  .modal-enter {
    opacity: 0;
  }

  .modal-leave-active {
    opacity: 0;
  }

  .modal-enter .modal-container,
  .modal-leave-active .modal-container {
    -webkit-transform: scale(1.1);
    transform: scale(1.1);
  }

  .my-custom-scrollbar {
    position: relative;
    height: 300px;
    overflow: auto;
  }
  .table-wrapper-scroll-y {
    display: block;
  }
</style>
