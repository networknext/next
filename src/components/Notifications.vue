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
                title="All Notifications">{{ notifications.length }}</span>
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
        <div class="mb-2 mb-md-0 flex-grow-1">
          <button v-if="false" class="btn btn-primary" type="button" data-toggle="collapse" data-target="#notification-0" aria-expanded="false" aria-controls="collapseExample">
            Test Collapse
          </button>
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
    <div class="card" v-for="(notification, index) in notifications" v-bind:key="index" style="margin-top: 20px;text-align: center;">
      <div class="card-header d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center">
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
      <div class="collapse collapse-all" v-bind:id="`notification-${index}`">
        <div class="card-body" v-if="notification.notification_type === 0">
          <div class="table-responsive table-no-top-line">
            <table class="table table-sm" :class="{'table-striped': notification.data.headers.length > 0, 'table-hover': notification.data.headers.length > 0}">
              <thead>
                <tr>
                  <th v-for="(header, index) in notification.data.headers" v-bind:key="index">
                    <span>{{ header }}</span>
                  </th>
                </tr>
              </thead>
              <tbody v-if="notification.data.rows.length === 0">
                <tr>
                  <td colspan="7" class="text-muted">
                      There is no data at this time.
                  </td>
                </tr>
              </tbody>
              <tbody>
                <tr v-for="(row, index) in notification.data.rows" v-bind:key="index">
                  <td v-for="(header, index) in notification.data.headers" v-bind:key="index">
                    {{ row }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="card-body" v-if="notification.notification_type === 1">
          <iframe
            v-bind:src="notification.data.url"
            width="600"
            height="300"
            frameborder="0">
          </iframe>
        </div>
    </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Notifications extends Vue {
  private notifications: Array<any>
  private urgentNotifications: Array<any>
  private infoNotifications: Array<any>
  private unwatchNotifications: any
  private notificationLoop: any

  constructor () {
    super()
    this.notifications = []
    this.urgentNotifications = []
    this.infoNotifications = []
  }

  private mounted () {
    this.restartLoop()
    this.unwatchNotifications = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        clearInterval(this.notificationLoop)
        this.restartLoop()
      }
    )
  }

  private beforeDestroy () {
    clearInterval(this.notificationLoop)
    this.unwatchNotifications()
  }

  private fetchNotifications () {
    this.$apiService
      .fetchNotifications()
      .then((response: any) => {
        console.log(response.notifications)
        this.notifications = response.notifications || []
        this.infoNotifications = this.notifications.filter((notification: any) => {
          return notification.priority === 1
        })
        this.urgentNotifications = this.notifications.filter((notification: any) => {
          return notification.priority === 3
        })
      })
  }

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
