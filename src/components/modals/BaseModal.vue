<template>
  <transition name="modal" v-if="showModal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-body ml-2 mr-2" style="overflow-y: scroll;">
            <slot name="body"></slot>
          </div>
          <div class="card-footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

// TODO: extend this to allow for different sized modals via props

/**
 * This component is a base class for all modals. It should be
 * used to control modal state, as well as basic layout of the modal
 */

@Component
export default class BaseModal extends Vue {
  private showModal: boolean

  constructor () {
    super()
    this.showModal = false
  }

  public toggleShowModal () {
    this.showModal = true
  }

  public toggleHideModal () {
    this.showModal = false
  }

  public isVisible () {
    return this.showModal
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  li {
    padding-bottom: 1rem;
  }
  .paragraph-header {
    font-size: 16px;
    font-weight: bold;
    padding-bottom: 1rem;
  }
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
    max-height: 600px;
    margin: 0px auto 8%;
    background-color: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
    transition: all 0.3s ease;
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
