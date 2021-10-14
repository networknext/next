<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-body">
            <div class="card-title">
              <div class="row">
                <div class="col"></div>
                <img class="logo-sizing" src="https://storage.googleapis.com/network-next-press-kit/networknext_logo_colour_black_RGB.png" />
                <div class="col"></div>
              </div>
              <div class="row">
                <div class="col"></div>
                <h2 class="header">Log in</h2>
                <div class="col"></div>
              </div>
            </div>
            <div class="form-group">
              <input
                type="text"
                class="form-control"
                id="email-input"
                placeholder="Email"
                v-model="email"
              />
              <br />
              <input
                type="password"
                class="form-control"
                id="password-input"
                placeholder="Password"
                v-model="password"
              />
              <small v-if="passwordError !== ''" class="text-danger">
                {{ passwordError }}
                <br/>
              </small>
            </div>
            <div style="padding: 1rem 0 1rem 0;"><strong>Forgot Password?</strong></div>
            <button class="btn btn-primary btn-block" @click="login()">
              Log in
            </button>
            <div style="padding: 1rem 0 1rem 0;">Don't have an account? <router-link to="get-access"><strong>Get Access</strong></router-link></div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component opens up a login form modal
 */

@Component
export default class LoginModal extends Vue {
  private email: string
  private password: string
  private passwordError: string
  private stepOne: boolean

  constructor () {
    super()
    this.email = ''
    this.password = ''
    this.passwordError = ''
    this.stepOne = true
  }

  private login (): void {
    this.$authService.login(this.email, this.password).catch((err: Error) => {
      this.password = ''
      this.passwordError = err.message
      setTimeout(() => {
        this.passwordError = ''
      }, 3000)
    })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .logo-sizing {
    width: 320px;
    height: 37px;
  }
  .header {
    padding-top: 1rem;
  }
  .modal-mask {
    position: fixed;
    z-index: 9998;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgb(0, 0, 0);
    display: table;
  }

  .modal-wrapper {
    display: table-cell;
    vertical-align: middle;
  }

  .modal-container {
    max-width: 400px;
    max-height: 600px;
    margin: 0px auto;
    background-color: #fff;
    border-radius: 5px;
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
