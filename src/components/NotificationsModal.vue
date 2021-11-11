<template>
  <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="card modal-container">
          <div class="card-header">
            <div class="banner d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center">
              <div class="banner-message">
                Account Notifications
              </div>
            </div>
          </div>
          <div class="card-body" id="notifications-page" style="overflow-y: scroll; padding: 0 1.25rem 1.25rem 1.25rem">
            <div class="card" v-for="(notification, index) in releaseNotesNotifications" :key="`release-notes-notification-${index}`" style="margin-top: 20px;text-align: center;">
              <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  {{ notification.title }}
                </div>
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-left"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#release-notes-notification-${index}`"
                  />
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-down"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#release-notes-notification-${index}`"
                  />
                </div>
              </div>
              <div class="collapse collapse-release-notes" :id="`release-notes-notification-${index}`">
                <div class="card-body" style="padding-bottom: calc(1.25rem - 16px);">
                  <GistEmbed :cssURL="notification.css_url" :embedHTML="notification.embed_html"/>
                </div>
              </div>
            </div>
            <!-- TODO: It may be a good idea to break out these collapsable cards to be used elsewhere potentially -->
            <!-- TODO TODO: Yes please do this for notifications. Use prop to determine type -->
            <div class="card" v-for="(notification, index) in analyticsNotifications" :key="`analytics-notification-${index}`" style="margin-top: 20px;text-align: center;">
              <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  {{ notification.title }}
                </div>
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-left"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#analytics-notification-${index}`"
                  />
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-down"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#analytics-notification-${index}`"
                  />
                </div>
              </div>
              <div class="collapse collapse-analytics" :id="`analytics-notification-${index}`">
                <div class="card-body">
                  <div class="row">
                    <div class="col-2">
                      Super cool analytics look :)
                    </div>
                    <div class="col">
                      {{ notification.message }}
                    </div>
                    <div class="col-2">
                      <button id="analytics-demo-signup" class="btn btn-success btn-sm" @click="startAnalyticsTrial()">
                        Start Free Trial
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="card" v-for="(notification, index) in systemNotifications" :key="index" style="margin-top: 20px;text-align: center;">
              <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  {{ notification.title }}
                </div>
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-left"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#system-notification-${index}`"
                  />
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-down"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#system-notification-${index}`"
                  />
                </div>
              </div>
              <div class="collapse collapse-system" :id="`system-notification-${index}`">
                <div class="card-body">
                  <div class="row">
                    <div class="col" style="max-width:500px;height:300px;">
                    </div>
                    <div class="col">
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="card" v-for="(notification, index) in invoiceNotifications" :key="index" style="margin-top: 20px;text-align: center;">
              <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  {{ notification.title }}
                </div>
                <div class="mb-2 mb-md-0 flex-grow-1"></div>
                <div class="pr-5">
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-left"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#invoice-notification-${index}`"
                  />
                  <font-awesome-icon
                    aria-expanded="false"
                    id="status"
                    icon="chevron-down"
                    class="fa-w-16 fa-fw"
                    type="button"
                    data-toggle="collapse"
                    :data-target="`#invoice-notification-${index}`"
                  />
                </div>
              </div>
              <div class="collapse collapse-invoice" :id="`invoice-notification-${index}`">
                <div class="card-body">
                  <div class="row">
                    <div class="col" style="max-width:500px;height:300px;">
                    </div>
                    <div class="col">
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="card-footer">
            <button class="btn btn-xs btn-primary modal-default-button" style="max-height: 40px; width: 100px;" @click="$root.$emit('hideNotificationsModal')">
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
import GistEmbed from '@/components/GistEmbed.vue'

/**
 * This component opens up a modal for picking from nested map points
 */

/**
 * TODO: Clean up template
 */

@Component({
  components: {
    GistEmbed
  }
})
export default class NotificationsModal extends Vue {
  private analyticsNotifications: Array<any>
  private invoiceNotifications: Array<any>
  private systemNotifications: Array<any>
  private releaseNotesNotifications: Array<any>

  constructor () {
    super()
    this.analyticsNotifications = []
    this.invoiceNotifications = []
    this.systemNotifications = []
    this.releaseNotesNotifications = []
  }

  private created () {
    if ((this.$store.getters.isAdmin || this.$store.getters.isOwner) && this.$store.getters.isBuyer) {
      this.fetchNotifications()
    }
  }

  private fetchNotifications () {
    this.$apiService
      .fetchNotifications()
      .then((response: any) => {
        this.releaseNotesNotifications = response.release_notes_notifications || []
        this.analyticsNotifications = response.analytics_notifications || []
        this.invoiceNotifications = response.invoice_notifications || []
        this.systemNotifications = response.system_notifications || []
      })
      .catch((error: Error) => {
        console.log('Something went wrong fetching notifications')
        console.log(error)
      })
  }

  private startAnalyticsTrial () {
    this.$apiService
      .startAnalyticsTrial()
      .then(() => {
        return this.$authService.refreshToken()
      })
      .catch((error: Error) => {
        console.log('Something went wrong setting up analytics trial')
        console.log(error)
      })
      .then(() => {
        this.$root.$emit('showAnalyticsTrialResponse')
        // TODO: Start a tour for analytics here?
        this.fetchNotifications()
      })
  }
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
    max-height: 600px;
    margin: 0px auto 10%;
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
  [aria-expanded=true].fa-chevron-left {
    display: none;
  }
  [aria-expanded=false].fa-chevron-down {
    display: none;
  }
  .fa-clipboard {
    font-size: 24px;
  }
  .fa-bell{
    font-size: 24px;
  }
  .fa-file-alt {
    font-size: 24px;
  }
  .fa-credit-card{
    font-size: 24px;
  }
  .custom-badge-all {
    background-color: white;
    border-style: solid;
    border-color: #009FDF;
    border-width: 2px;
    cursor: default;
  }
  .banner {
    width: 100%;
    height: 40px;
    font-size: 24px;
  }
  .banner-message {
    min-width: 300px;
    padding-left: 1rem;
  }
  .center-text {
    margin-top: .5rem;
  }
  .fixed-width {
    font-family: monospace;
    font-size: 120%;
  }
  div.table-no-top-line th {
    border-top: none !important;
  }
</style>
