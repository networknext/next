<template>
  <div class="card-body" id="notifications-page">
    <!-- TODO: Add a container here for the title information and floated right button to refresh notifications -->
    <div class="card-title">
      <!-- TODO: Add banner with banner message | card count                    filters -->
      <div class="banner d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center">
        <div class="banner-message">
          Account Notifications
        </div>
        <div class="pl-4">
          <span class="badge rounded-pill custom-badge-all"
                style="font-size: 1rem;margin-right: 1rem;"
                data-toggle="tooltip"
                data-placement="right"
                title="All Notifications">{{ analyticsNotifications.length + systemNotifications.length + invoiceNotifications.length + 1 }}</span>
          <span class="badge rounded-pill custom-badge-info"
                style="font-size: 1rem;margin-right: 1rem;"
                data-toggle="tooltip"
                data-placement="right"
                title="Info Notifications">{{ infoNotifications.length }}</span>
          <span class="badge rounded-pill custom-badge-urgent"
                style="font-size: 1rem;margin-right: 1rem;"
                data-toggle="tooltip"
                data-placement="right"
                title="Urgent Notifications">{{ urgentNotifications.length }}</span>
        </div>
        <div class="pr-5">
          <font-awesome-icon
            aria-expanded="false"
            id="status"
            icon="minus-square"
            class="fa-w-16 fa-fw"
            type="button"
            data-toggle="collapse"
            data-target=".collapse-all"
          />
          <font-awesome-icon
            aria-expanded="false"
            id="status"
            icon="plus-square"
            class="fa-w-16 fa-fw"
            type="button"
            data-toggle="collapse"
            data-target=".collapse-all"
          />
        </div>
      </div>
    </div>
    <!-- TODO: It may be a good idea to break out these collapsable cards to be used elsewhere potentially -->
    <!-- TODO TODO: Yes please do this for notifications. Use prop to determine type -->
    <div class="card" v-for="(notification, index) in analyticsNotifications" v-bind:key="index" style="margin-top: 20px;text-align: center;">
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
            v-bind:data-target="`#notification-${index}`"
          />
          <font-awesome-icon
            aria-expanded="false"
            id="status"
            icon="chevron-down"
            class="fa-w-16 fa-fw"
            type="button"
            data-toggle="collapse"
            v-bind:data-target="`#notification-${index}`"
          />
        </div>
      </div>
      <div class="collapse collapse-all" :id="`analytics-notification-${index}`">
        <div class="card-body">
          <div class="row">
            <div class="col" style="max-width:500px;height:300px;">
              <iframe
                style="padding-top:10px;padding-bottom:10px;"
                v-if="notification.analytics_url !== ''"
                v-bind:src="notification.analytics_url"
                width="300"
                height="300"
                frameborder="0">
              </iframe>
            </div>
            <div class="col">
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card" style="margin-top: 20px;text-align: center;" v-if="showReleaseNotes">
      <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
        <div class="mb-2 mb-md-0 flex-grow-1"></div>
        <div class="pr-5">
          {{ releaseNotes.title }}
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
            data-target="#release-notes-notification"
          />
          <font-awesome-icon
            aria-expanded="false"
            id="status"
            icon="chevron-down"
            class="fa-w-16 fa-fw"
            type="button"
            data-toggle="collapse"
            data-target="#release-notes-notification"
          />
        </div>
      </div>
      <div class="collapse collapse-all" id="release-notes-notification">
        <div class="card-body" style="padding-bottom: calc(1.25rem - 16px);">
          <GistEmbed :cssURL="releaseNotes.css_url" :embedHTML="releaseNotes.embed_html"/>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import GistEmbed from '@/components/GistEmbed.vue'

@Component({
  components: {
    GistEmbed
  }
})
export default class Notifications extends Vue {
  private analyticsNotifications: Array<any>
  private invoiceNotifications: Array<any>
  private systemNotifications: Array<any>

  private releaseNotes: any
  private showReleaseNotes: boolean

  private urgentNotifications: Array<any>
  private infoNotifications: Array<any>

  private unwatchNotifications: any
  private notificationLoop: any

  constructor () {
    super()
    this.analyticsNotifications = []
    this.invoiceNotifications = []
    this.systemNotifications = []
    this.urgentNotifications = []
    this.infoNotifications = []
    this.showReleaseNotes = false
  }

  private created () {
    this.fetchNotifications()
  }

  private fetchNotifications () {
    this.$apiService
      .fetchNotifications()
      .then((response: any) => {
        this.releaseNotes = response.release_notes_notification
        this.showReleaseNotes = true
      })
  }

  // TODO: Figure out if this is necessary or not
  private restartLoop () {
    if (this.notificationLoop) {
      clearInterval(this.notificationLoop)
    }
    this.fetchNotifications()
    this.notificationLoop = setInterval(() => {
      this.fetchNotifications()
    }, 60000000)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  [aria-expanded=true].fa-chevron-left {
    display: none;
  }
  [aria-expanded=false].fa-chevron-down {
    display: none;
  }
  [aria-expanded=true].fa-plus-square {
    display: none;
  }
  [aria-expanded=false].fa-minus-square {
    display: none;
  }
  .fa-plus-square {
    font-size: 24px;
  }
  .fa-minus-square {
    font-size: 24px;
  }
  .custom-badge-all {
    background-color: white;
    border-style: solid;
    border-color: #009FDF;
    border-width: 2px;
    cursor: default;
  }
  .custom-badge-info {
    background-color: white;
    border-style: solid;
    border-color: blue;
    border-width: 2px;
    cursor: default;
  }
  .custom-badge-urgent {
    background-color: white;
    border-style: solid;
    border-color: red;
    border-width: 2px;
    cursor: default;
  }
  .banner {
    width: 100%;
    height: 40px;
    border-color: #dfdfdf;
    border-width: 1px;
    border-style: solid;
    border-radius: .25rem;
  }
  .banner-message {
    min-width: 300px;
    border-color: #dfdfdf;
    border-width: 1px;
    border-style: none solid none none;
    text-align: center;
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
