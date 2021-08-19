<template>
  <div>
    <v-tour name="getAccessTour" :steps="getAccessTourSteps" :options="getAccessTourOptions" :callbacks="getAccessTourCallbacks"></v-tour>
    <v-tour v-show="$store.getters.currentPage !== 'downloads'" name="downloadLinkTour" :steps="downloadLinkTourSteps" :options="downloadLinkTourOptions" :callbacks="downloadLinkTourCallbacks"></v-tour>
    <nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark p-0 shadow">
      <a class="navbar-brand col-sm-3 col-md-2 mr-0" href="https://networknext.com">
        <div class="logo-container">
          <img class="logo-fit" src="../assets/logo.png" />
        </div>
      </a>
      <ul class="navbar-nav px-3 w-100 mr-auto">
        <li class="nav-item text-nowrap">
          <router-link
            to="/map"
            class="nav-link"
            v-bind:class="{ active: $store.getters.currentPage == 'map' }"
            data-test="mapLink"
          >Map</router-link>
        </li>
        <li class="nav-item text-nowrap">
          <router-link
            to="/sessions"
            class="nav-link"
            data-intercom="sessions"
            v-bind:class="{ active: $store.getters.currentPage == 'sessions' }"
            data-test="sessionsLink"
            data-tour="sessionsLink"
          >Sessions</router-link>
        </li>
        <li class="nav-item text-nowrap">
          <router-link
            to="/session-tool"
            class="nav-link"
            v-bind:class="{
              active:
                $store.getters.currentPage == 'session-tool' ||
                $store.getters.currentPage == 'session-details'
            }"
            data-test="sessionToolLink"
          >Session Tool</router-link>
        </li>
        <li class="nav-item text-nowrap">
          <router-link
            to="/user-tool"
            class="nav-link"
            v-bind:class="{
              active:
                $store.getters.currentPage == 'user-tool' ||
                $store.getters.currentPage == 'user-sessions'
            }"
            v-if="!$store.getters.isAnonymous"
          >User Tool</router-link>
        </li>
        <li class="nav-item text-nowrap">
          <router-link
            to="/downloads"
            class="nav-link"
            data-intercom="downloads"
            data-tour="downloadsLink"
            v-bind:class="{ active: $store.getters.currentPage == 'downloads' }"
            v-if="!$store.getters.isAnonymous && !$store.getters.isAnonymousPlus"
          >Downloads</router-link>
        </li>
        <li class="nav-item text-nowrap">
          <router-link
            to="/explore"
            class="nav-link"
            v-bind:class="{
              active:
                $store.getters.currentPage == 'notifications' ||
                $store.getters.currentPage == 'analytics' ||
                $store.getters.currentPage == 'invoicing'
            }"
            v-if="$store.getters.isOwner || $store.getters.isAdmin"
          >Explore</router-link>
        </li>
        <li class="nav-item text-nowrap">
          <router-link
            to="/settings"
            class="nav-link"
            v-bind:class="{
              active:
                $store.getters.currentPage == 'account-settings' ||
                $store.getters.currentPage == 'config' ||
                $store.getters.currentPage == 'users' ||
                $store.getters.currentPage == 'route-shader'
            }"
            v-if="!$store.getters.isAnonymous && !$store.getters.isAnonymousPlus"
          >Settings</router-link>
        </li>
      </ul>
      <ul class="navbar-nav px-3 w-100">
        <li class="nav-item text-nowrap" style="color: #9a9da0;">{{ portalVersion }}</li>
      </ul>
      <ul class="navbar-nav px-1" v-if="!$store.getters.isAnonymous">
        <li class="nav-item text-nowrap" style="color: white;">
          {{ $store.getters.userProfile.email || "" }}
        </li>
      </ul>
      <ul class="navbar-nav px-3" v-if="$store.getters.isAnonymous">
        <li class="nav-item text-nowrap">
          <a data-test="loginButton" class="login btn-sm btn-primary" href="#" @click="login()">Log in</a>
        </li>
      </ul>
      <ul class="navbar-nav px-3" v-if="$flagService.isEnabled(FeatureEnum.FEATURE_IMPERSONATION) && $store.getters.isAdmin">
        <li class="nav-item text-nowrap">
          <select v-on:change="impersonate($event.target.value)">
            <option :value="''">Impersonate</option>
            <option
              :value="buyer.company_code"
              v-for="buyer in allBuyers"
              v-bind:key="buyer.company_code"
              :selected="buyer.company_code === companyCode"
            >{{ buyer.company_name }}</option>
          </select>
        </li>
      </ul>
      <ul class="navbar-nav px-3" v-if="$store.getters.isAnonymous">
        <li class="nav-item text-nowrap" data-tour="signUpButton">
          <a
            data-test="signUpButton"
            data-intercom="signUpButton"
            class="signup btn-sm btn-primary"
            href="#"
            @click="signUp()"
          >Get Access</a>
        </li>
      </ul>
      <ul class="navbar-nav px-3" v-if="!$store.getters.isAnonymous">
        <li class="nav-item text-nowrap">
          <a class="logout btn-sm btn-primary" href="#" @click="logout()">Logout</a>
        </li>
      </ul>
    </nav>
  </div>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash'
import { Component, Vue } from 'vue-property-decorator'
import { FeatureEnum } from './types/FeatureTypes'

/**
 * This component opens up the main Vue router handlers to user interaction in the form of a navigation bar
 */

/**
 * TODO: Clean up template
 */

@Component
export default class NavBar extends Vue {
  get allBuyers () {
    if (!this.$store.getters.isAdmin) {
      return []
    }
    return this.$store.getters.allBuyers
  }

  private companyCode: string
  private portalVersion: string

  private unwatch: any
  private getAccessTourSteps: Array<any>
  private getAccessTourOptions: any
  private getAccessTourCallbacks: any

  private downloadLinkTourSteps: Array<any>
  private downloadLinkTourOptions: any
  private downloadLinkTourCallbacks: any

  private FeatureEnum: any

  constructor () {
    super()
    this.portalVersion = ''
    this.companyCode = ''

    this.getAccessTourSteps = [
      {
        target: '[data-tour="signUpButton"]',
        header: {
          title: 'Get Access'
        },
        content: '<strong>Try it for your game for FREE!</strong><br><br> Just create an account and log in to try Network Next: <ul><li>Download the open source SDK and documentation.</li><li>Integrate the SDK into your game.</li></ul> Now you\'re in control of the network. Please contact us in <strong>chat</strong> (lower right) if you have any questions.'
      }
    ]

    this.getAccessTourOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'OK'
      }
    }

    this.getAccessTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_TOURS', 'get-access')
        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          Vue.prototype.$gtag.event('Get access tour finished', {
            event_category: 'Tours'
          })
        }
      }
    }

    this.downloadLinkTourSteps = [
      {
        target: '[data-tour="downloadsLink"]',
        header: {
          title: 'Downloads'
        },
        content: 'You\'re now logged in! You can now integrate the Network Next SDK into your game to start accelerating your traffic.<br><br>The SDK is in the <strong>Downloads</strong> section.',
        params: {
          placement: 'bottom',
          enableScrolling: false
        }
      }
    ]

    this.downloadLinkTourOptions = this.getAccessTourOptions
    this.downloadLinkTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_SIGN_UP_TOURS', 'downloadLink')

        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          Vue.prototype.$gtag.event('Download link tour finished', {
            event_category: 'Tours'
          })
        }
      }
    }
  }

  private created () {
    this.fetchPortalVersion()
    this.FeatureEnum = FeatureEnum
  }

  private mounted () {
    if (this.companyCode === '') {
      const userProfile = cloneDeep(this.$store.getters.userProfile)
      this.companyCode = userProfile.companyCode || ''
    }

    if (this.$store.getters.isSignUpTour && this.$tours.downloadLinkTour && !this.$tours.downloadLinkTour.isRunning && !this.$store.getters.finishedSignUpTours.includes('downloadLink') && this.$route.name !== 'downloads') {
      setTimeout(() => {
        this.$tours.downloadLinkTour.start()
      }, 3000)
    }
    this.unwatch = this.$store.watch(
      (_, getters: any) => {
        return getters.finishedTours
      },
      (finishedTours: Array<string>) => {
        if (this.$tours.getAccessTour && !this.$tours.getAccessTour.isRunning && !finishedTours.includes('get-access') && finishedTours.length > 0) {
          this.$tours.getAccessTour.start()
        }
      }
    )
  }

  private login (): void {
    this.$authService.login()
  }

  private logout (): void {
    this.$authService.logout()
  }

  private signUp (): void {
    if (process.env.VUE_APP_MODE === 'prod') {
      this.$gtag.event('clicked sign up', {
        event_category: 'Account Creation',
        event_label: 'Sign up'
      })
    }
    this.$authService.signUp()
  }

  private impersonate (companyCode: string): void {
    this.$apiService.impersonate({ company_code: companyCode })
      .then((response: any) => {
        this.$authService.refreshToken()
      })
      .catch((error: Error) => {
        console.log('something went wrong with impersonating')
        console.log(error)
      })
  }

  private fetchPortalVersion (): void {
    let url = ''

    if (process.env.VUE_APP_MODE === 'local') {
      url = `${process.env.VUE_APP_API_URL}`
    }

    if (process.env.VUE_APP_MODE === 'dev' || process.env.VUE_APP_MODE === 'local') {
      fetch(`${url}/version`, {
        headers: {
          Accept: 'application/json',
          'Accept-Encoding': 'gzip',
          'Content-Type': 'application/json'
        },
        method: 'POST'
      }).then((response: any) => {
        response.json().then((json: any) => {
          if (json.error) {
            throw new Error(json.error)
          }
          this.portalVersion = `Git Hash: ${json.sha}`
          if (json.commit_message) {
            this.portalVersion = `${this.portalVersion} - Commit: ${json.commit_message}`
          }
        })
      }).catch((error: Error) => {
        console.log('Something went wrong fetching the software version')
        console.log(error)
      })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
.navbar-brand {
  display: inline-block;
  padding-top: 0.3125rem;
  padding-bottom: 0.3125rem;
  margin-right: 1rem;
  font-size: 1.25rem;
  line-height: inherit;
  white-space: nowrap;
}
.navbar .form-control {
  padding: 0.75rem 1rem;
  border-width: 0;
  border-radius: 0;
}
@media not all and (min-resolution:.001dpcm) {
  @supports (-webkit-appearance:none) {
    .signup {
      width: 6rem;
      height: 1.7rem;
      display: block;
      text-align: center;
      border-color: #343a40;
      border-style: solid;
      border-width: 1px;
      border-radius: 9999px;
      background-color: #FF6700;
      font-weight: bold;
      line-height: 1.1rem;
    }
  }
}

.signup {
  width: 6rem;
  height: 1.7rem;
  display: block;
  text-align: center;
  border-radius: 9999px;
  background-color: #FF6700;
  font-weight: bold;
  line-height: 1.1rem;
}
.signup:hover {
  text-decoration: none;
  background-color: rgba(255, 103, 0, 0.9);
}
.signup:not(:disabled):not(.disabled):active {
  background-color: #007bff;
  text-decoration: none;
  border-color: #007bff;
  box-shadow: none;
}
.signup:focus {
  background-color: #007bff;
  text-decoration: none;
  border-color: #007bff;
  box-shadow: none;
}
.login {
  width: 6rem;
  height: 1.7rem;
  display: block;
  border-width: 1px;
  border-color: #FF6700;
  border-radius: 9999px;
  text-align: center;
  background-color: #343a40;
  border-style: solid;
  font-weight: bold;
  line-height: 1rem;
}
.login:hover {
  background-color: rgba(255, 103, 0, 0.1);
  text-decoration: none;
  border-color: #FF6700;
}
.login:not(:disabled):not(.disabled):active {
  background-color: #343a40;
  text-decoration: none;
  border-color: #007bff;
  box-shadow: none;
}
.login:focus {
  background-color: rgba(0, 123, 255, 0.1);
  text-decoration: none;
  border-color: #007bff;
  box-shadow: none;
}
.logout {
  width: 6rem;
  height: 1.7rem;
  display: block;
  border-width: 1px;
  border-color: #FF6700;
  border-radius: 9999px;
  text-align: center;
  background-color: #343a40;
  border-style: solid;
  font-weight: bold;
  line-height: 1rem;
}
.logout:hover {
  background-color: rgba(255, 103, 0, 0.1);
  text-decoration: none;
  border-color: #FF6700;
}
.logout:not(:disabled):not(.disabled):active {
  background-color: #343a40;
  text-decoration: none;
  border-color: #007bff;
  box-shadow: none;
}
.logout:focus {
  background-color: rgba(0, 123, 255, 0.1);
  text-decoration: none;
  border-color: #007bff;
  box-shadow: none;
}
</style>
