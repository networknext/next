<template>
  <BaseModal ref="baseModal">
    <template v-slot:body>
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
                  @click.native="toggleHideModal()"
                  :to="`/session-tool/${point.source[3]}`"
                  class="text-dark fixed-width"
                >{{ point.source[3] }}</router-link>
              </td>
              <td v-if="!point.source[3] || point.source[3] === ''">
                Unavailable
              </td>
              <td>{{ point.source[1] }}</td>
              <td>{{ point.source[0] }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
    <template v-slot:footer>
      <button class="btn btn-xs btn-primary modal-default-button" style="max-height: 40px; width: 100px;" @click="toggleHideModal()">
        Close
      </button>
    </template>
  </BaseModal>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator'
import { NavigationGuardNext, Route } from 'vue-router'
import BaseModal from './BaseModal.vue'

/**
 * This component opens up a modal for picking from nested map points
 */

/**
 * TODO: Clean up template
 */

@Component({
  components: {
    BaseModal
  }
})
export default class MapPointsModal extends Vue {
  @Prop({ default: () => { return [] } }) readonly points!: Array<any>

  $refs!: {
    baseModal: BaseModal;
  }

  // These are weird helper methods that act like inherited methods.
  // TODO: There may be a better way of doing inheritance with slots
  //       Would be better if component inherited from BaseModal and
  //       was still able to use slots and access parent functions
  public toggleShowModal () {
    this.$refs.baseModal.toggleShowModal()
  }

  public toggleHideModal () {
    this.$refs.baseModal.toggleHideModal()
  }

  public isVisible () {
    return this.$refs.baseModal.isVisible()
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .modal-default-button {
    float: right;
    border-color: #009FDF;
    background-color: #009FDF;
  }

  .modal-default-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
</style>
