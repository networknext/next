<template>
  <div class="card-body" id="notifications-page">
    <!-- TODO: Add a container here for the title information and floated right button to refresh notifications -->
    <div class="card-title">
      <!-- TODO: Add banner with banner message | card count                    filters -->
      <div class="banner d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center">
        <div class="banner-message">
          Account Notifications
        </div>
        <div style="min-width: 30%;">
          <span class="badge rounded-pill custom-badge-all"
                style="font-size: 1rem;min-width:300px;"
                data-toggle="tooltip"
                data-placement="right"
                v-if="analyticsNotifications.length + systemNotifications.length + invoiceNotifications.length > 0"
                title="All Notifications">New Notifications: {{ analyticsNotifications.length + systemNotifications.length + invoiceNotifications.length }}</span>
        </div>
        <div class="pr-5">
          <span data-toggle="tooltip"
                data-placement="right"
                v-if="releaseNotesNotifications.length > 0"
                style="padding-right:5px;"
                title="Toggle Release Notes Notifications">
            <font-awesome-icon
              aria-expanded="false"
              id="status"
              icon="clipboard"
              class="fa-w-16 fa-fw"
              type="button"
              data-toggle="collapse"
              data-target=".collapse-release-notes"
            />
          </span>
          <span data-toggle="tooltip"
                data-placement="right"
                v-if="systemNotifications.length > 0"
                style="padding-right:5px;"
                title="Toggle System Notifications">
            <font-awesome-icon
              aria-expanded="false"
              id="status"
              icon="bell"
              class="fa-w-16 fa-fw"
              type="button"
              data-toggle="collapse"
              data-target=".collapse-system"
            />
          </span>
          <span data-toggle="tooltip"
                data-placement="right"
                v-if="analyticsNotifications.length > 0"
                style="padding-right:5px;"
                title="Toggle Analysis Notifications">
            <font-awesome-icon
              aria-expanded="false"
              id="status"
              icon="file-alt"
              class="fa-w-16 fa-fw"
              type="button"
              data-toggle="collapse"
              data-target=".collapse-analysis"
            />
          </span>
          <span data-toggle="tooltip"
                data-placement="right"
                v-if="invoiceNotifications.length > 0"
                title="Toggle Invoice Notifications">
            <font-awesome-icon
              aria-expanded="false"
              id="status"
              icon="credit-card"
              class="fa-w-16 fa-fw"
              type="button"
              data-toggle="collapse"
              data-target=".collapse-invoice"
            />
          </span>
        </div>
      </div>
    </div>
    <div class="card" v-for="(notification, index) in releaseNotesNotifications" :key="index" style="margin-top: 20px;text-align: center;">
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
    <div class="card" v-for="(notification, index) in analyticsNotifications" :key="index" style="margin-top: 20px;text-align: center;">
      <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center" >
        <div class="mb-2 mb-md-0 flex-grow-1"></div>
        <div>
          {{ notification.title }}
        </div>
        <div class="mb-2 mb-md-0 flex-grow-1"></div>
        <div>
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
  private releaseNotesNotifications: Array<any>

  constructor () {
    super()
    this.analyticsNotifications = []
    this.invoiceNotifications = []
    this.systemNotifications = []
    this.releaseNotesNotifications = []
  }

  private created () {
    this.fetchNotifications()
  }

  private fetchNotifications () {
    this.$apiService
      .fetchNotifications()
      .then((response: any) => {
        this.releaseNotesNotifications = response.release_notes_notifications
      })
      .catch((error: Error) => {
        console.log('Something went wrong fetching notifications')
        console.log(error)
      })
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
