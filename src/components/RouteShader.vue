<template>
  <div class="card-body">
    <h5 class="card-title">
        Route Shader
    </h5>
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
        <button type="submit" class="btn btn-primary btn-sm" v-if="$store.getters.isOwner() || $store.getters.isAdmin()">
          Save route shader
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Multiselect from 'vue-multiselect'
import APIService from '../services/api.service'

@Component({
  components: {
    Multiselect
  }
})
export default class UserManagement extends Vue {
  // TODO: Fix weird issue with dropdown library change events (select/delete) handler
  private apiService: APIService
  private routeShader: any

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
  }

  private mounted () {
    this.routeShader = this.$store.getters.userProfile.routeShader
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
