<template>
  <div class="card-body">
    <h5 class="card-title">
        Route Shader
    </h5>
    <Alert :alertType="alertType" :message="message" v-if="message !== ''"/>
    <form v-on:submit.prevent="updateRouteShader()">
        <div class="form-group row">
            <div class="col-sm-2">Enable Network Next</div>
            <div class="col-sm-10">
                <div class="form-check">
                    <label class="switch">
                        <input type="checkbox" checked v-model="routeShader.enable_nn">
                        <span class="slider round"></span>
                    </label>
                    <span class="text-pad-left">
                        Accelerate your sessions with Network Next
                    </span>
                </div>
            </div>
        </div>
        <div v-show="routeShader.enable_nn">
          <div class="form-group row">
            <div class="col-sm-2">Reduce Latency</div>
            <div class="col-sm-10">
              <div class="form-check">
                <label class="switch">
                  <input type="checkbox" checked v-model="routeShader.enable_rtt">
                  <span class="slider round"></span>
                </label>
                <span class="text-pad-left">
                  Reduce latency when it's higher than an acceptable amount
                </span>
              </div>
            </div>
          </div>
          <div class="form-group row text-pad-left" v-show="routeShader.enable_rtt">
            <div class="col-sm-2">Acceptable Latency:</div>
            <div class="col-sm-10">
              <input type="text" v-model="routeShader.acceptable_latency" class="text-input-width"> ms
            </div>
          </div>
        <div class="form-group row">
          <div class="col-sm-2">Reduce Packet Loss</div>
          <div class="col-sm-10">
            <div class="form-check">
              <label class="switch">
                <input type="checkbox" v-model="routeShader.enable_pl">
                <span class="slider round"></span>
              </label>
              <span class="text-pad-left">
                Reduce packet loss when it's higher than an acceptable amount
              </span>
            </div>
          </div>
        </div>
        <div class="form-group row text-pad-left" v-show="routeShader.enable_pl">
          <div class="col-sm-2">Acceptable Packet Loss:</div>
          <div class="col-sm-10">
            <input type="text" v-model="routeShader.pl_threshold" class="text-input-width"> %
          </div>
        </div>
        <div class="form-group row">
          <div class="col-sm-2">Multipath</div>
          <div class="col-sm-10">
            <div class="form-check">
              <label class="switch">
                <input type="checkbox" v-model="routeShader.enable_mp">
                <span class="slider round"></span>
              </label>
              <span class="text-pad-left">
                Send packets across Network Next <u>and</u> the public internet at the same time
              </span>
            </div>
          </div>
        </div>
        <div class="form-group row">
          <div class="col-sm-2">A/B Test</div>
          <div class="col-sm-10">
            <div class="form-check">
              <label class="switch">
                  <input type="checkbox" v-model="routeShader.enable_ab">
                  <span class="slider round"></span>
              </label>
              <span class="text-pad-left">
                Even session ids have Network Next enabled, odd do not
              </span>
            </div>
          </div>
        </div>
        <button type="submit" class="btn btn-primary btn-sm" v-if="$store.getters.isOwner || $store.getters.isAdmin">
          Save route shader
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import _ from 'lodash'
import Alert from '@/components/Alert.vue'
import { AlertTypes } from './types/AlertTypes'
import { UserProfile } from '@/components/types/AuthTypes.ts'

/**
 * This component displays all of the necessary information for the route shader tab
 *  within the settings page of the Portal and houses all the associated logic and api calls
 */

/**
 * TODO: Clean up template
 * TODO: Pretty sure the card-body can be taken out into a wrapper component - same with game config and user management...
 */

@Component({
  components: {
    Alert
  }
})
export default class RouteShader extends Vue {
  private routeShader: any
  private message: string
  private alertType: string
  private userProfile: UserProfile

  constructor () {
    super()
    this.userProfile = _.cloneDeep(this.$store.getters.userProfile)
    this.routeShader = this.userProfile.routeShader
    this.message = ''
    this.alertType = ''
  }

  public updateRouteShader () {
    this.$apiService
      .updateRouteShader(this.routeShader)
      .then((response: any) => {
        this.userProfile.routeShader = this.routeShader
        this.$store.commit('UPDATE_USER_PROFILE', this.userProfile)
        this.alertType = AlertTypes.SUCCESS
        this.message = 'Updated route shader successfully'
        setTimeout(() => {
          this.message = ''
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the route shader')
        console.log(error)
        this.alertType = AlertTypes.ERROR
        this.message = 'Failed to update router shader'
        setTimeout(() => {
          this.message = ''
        }, 5000)
      })
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .switch {
    position: relative;
    display: inline-block;
    width: 60px;
    height: 24px;
  }
  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #ccc;
    -webkit-transition: .4s;
    transition: .4s;
  }
  .slider:before {
    position: absolute;
    content: "";
    height: 16px;
    width: 16px;
    left: 4px;
    bottom: 4px;
    background-color: white;
    -webkit-transition: .4s;
    transition: .4s;
  }
  input:checked + .slider {
    background-color: #007bff;
  }
  input:focus + .slider {
    box-shadow: 0 0 1px #007bff;
  }
  input:checked + .slider:before {
    -webkit-transform: translateX(36px);
    -ms-transform: translateX(36px);
    transform: translateX(36px);
  }
  .slider.round {
    border-radius: 34px;
  }
  .slider.round:before {
    border-radius: 50%;
  }
  .text-pad-left {
    padding-left: 2rem;
  }
  .text-input-width {
    width: 3.1rem;
  }
</style>
