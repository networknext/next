<template>
  <div class="card-body">
    <h5 class="card-title">
        Route Shader
    </h5>
    <Alert classType="alertType" message="message" v-if="message !== ''"/>
    <form v-on:submit.prevent="updateRouteShader()">
        <div class="form-group row">
            <div class="col-sm-2">Enable Network Next</div>
            <div class="col-sm-10">
                <div class="form-check">
                    <label class="switch">
                        <input type="checkbox" checked v-model="routeShader.enable_nn">
                        <span class="slider round"></span>
                    </label>
                    <span style="padding-left: 2rem;">
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
                <span style="padding-left: 2rem;">
                  Reduce latency when it's higher than an acceptable amount
                </span>
              </div>
            </div>
          </div>
          <div class="form-group row" style="padding-left: 2rem;" v-show="routeShader.enable_rtt">
            <div class="col-sm-2">Acceptable Latency:</div>
            <div class="col-sm-10">
              <input type="text" v-model="routeShader.acceptable_latency" style="width: 3.1rem;"> ms
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
              <span style="padding-left: 2rem;">
                Reduce packet loss when it's higher than an acceptable amount
              </span>
            </div>
          </div>
        </div>
        <div class="form-group row" style="padding-left: 2rem;" v-show="routeShader.enable_pl">
          <div class="col-sm-2">Acceptable Packet Loss:</div>
          <div class="col-sm-10">
            <input type="text" v-model="routeShader.pl_threshold" style="width: 3.1rem;"> %
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
              <span style="padding-left: 2rem;">
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
              <span style="padding-left: 2rem;">
                Even session ids have Network Next enabled, odd do not
              </span>
            </div>
          </div>
        </div>
        <button type="submit" class="btn btn-primary btn-sm" v-show="$store.getters.isOwner || $store.getters.isAdmin">
          Save route shader
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import APIService from '../services/api.service'
import Alert from '@/components/Alert.vue'

@Component({
  components: {
    Alert
  }
})
export default class RouteShader extends Vue {
  // TODO: Fix weird issue with dropdown library change events (select/delete) handler
  private apiService: APIService
  private routeShader: any
  private message: string
  private classType: string

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
    this.routeShader = this.$store.getters.userProfile.routeShader
    this.message = ''
    this.classType = ''
  }

  public updateRouteShader () {
    this.apiService
      .updateRouteShader(this.routeShader)
      .then((response: any) => {
        const userProfile = this.$store.getters.userProfile
        userProfile.routeShader = this.routeShader
        this.$store.commit('UPDATE_USER_PROFILE', userProfile)
        this.classType = 'alert-success'
        this.message = 'Updated route shader successfully'
        setTimeout(() => {
          this.message = ''
          this.classType = ''
        }, 5000)
      })
      .catch((error: Error) => {
        console.log('Something went wrong updating the route shader')
        console.log(error)
        this.classType = 'alert-danger'
        this.message = 'Updating route shader failed'
        setTimeout(() => {
          this.message = ''
          this.classType = ''
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
</style>
