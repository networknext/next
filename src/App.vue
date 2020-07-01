<template>
  <div id="app">
    <nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark p-0 shadow">
      <a class="navbar-brand col-sm-3 col-md-2 mr-0" href="#">
          <div class="logo-container">
              <img class="logo-fit" src="./assets/logo.png">
          </div>
      </a>
      <ul class="navbar-nav px-3 w-100 mr-auto">
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Map
              </a>
          </li>
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Sessions
              </a>
          </li>
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Session Tool
              </a>
          </li>
          <li class="nav-item text-nowrap">
              <a class="nav-link">
                  User Tool
              </a>
          </li>
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Downloads
              </a>
          </li>
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Settings
              </a>
          </li>
      </ul>
      <ul class="navbar-nav px-3">
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Log in
              </a>
          </li>
      </ul>
      <ul class="navbar-nav px-3">
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Sign up
              </a>
          </li>
      </ul>
      <ul class="navbar-nav px-3" v-if="false">
          <li class="nav-item text-nowrap">
              <a class="nav-link" href="#">
                  Logout
              </a>
          </li>
      </ul>
    </nav>
    <div style="padding-top: 48px;">
      <div class="alert alert-primary"
            style="text-align: center;"
            role="alert"
            v-if="false"
      >
        Please confirm your email address: EMAIL ADDRESS
        <a href="#">
            Resend email
        </a>
      </div>
      <div class="alert alert-success"
          style="text-align: center;"
          role="alert"
          v-if="false"
      >
        Verification email was sent successfully. Please check your email for futher instructions.
      </div>
      <div class="alert alert-danger"
          style="text-align: center;"
          role="alert"
          v-if="false"
      >
        Something went wrong sending the verification email. Please try again later.
      </div>
    </div>
    <div class="container-fluid">
      <div class="row">
        <main role="main" class="col-md-12 col-lg-12 px-4" v-show="true">
          <div class="
                    d-flex
                    justify-content-between
                    flex-wrap flex-md-nowrap
                    align-items-center
                    pt-3
                    pb-2
                    mb-3
                    border-bottom
                  "
              id="map-sessions-count"
          >
            <div id="map-spinner" v-if="false">
              <div class="spinner-border" role="status">
                <span class="sr-only">Loading...</span>
              </div>
            </div>
            <h1 class="h2" id="workspace-title" v-if="true">
              Map&nbsp;
              <span class="badge badge-dark">
                MAPSESSIONS Total Sessions
              </span>&nbsp;
              <span class="badge badge-success">
                ONNN on Network Next
              </span>
            </h1>
            <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1">
              <div class="mr-auto"></div>
              <div class="px-2 hidden">
                <select class="form-control">
                  <option>
                    World
                  </option>
                  <option>
                    United States
                  </option>
                </select>
              </div>
              <div class="px-2 hidden">
                <select class="form-control">
                  <option value="everything">
                    Everything
                  </option>
                  <option value="direct">
                    Direct Only
                  </option>
                  <option value="next">
                    Network Next Only
                  </option>
                </select>
              </div>
              <div class="px-2 hidden">
                <select class="form-control">
                  <option value="all">
                    All
                  </option>
                  <option>
                    BUYERNAME ADMIN LOOP
                  </option>
                  <option>
                    BUYERNAME SINGLE BUYER
                  </option>
                </select>
              </div>
            </div>
          </div>
          <div id="map-container"
              class="map-container-no-offset"
              v-bind:class="{
                  'map-container-no-offset': true,
                  'map-container-offset': false,
              }"
          ></div>
        </main>
        <main role="main" class="col-md-12 col-lg-12 px-4" v-if="false">
          <div id="sessions">
            <div class="
                    d-flex
                    justify-content-between
                    flex-wrap flex-md-nowrap
                    align-items-center
                    pt-3
                    pb-2
                    mb-3
                    border-bottom"
            >
              <h1 class="h2" id="workspace-title" v-if="true">
                Sessions&nbsp;
                <span class="badge badge-dark">
                  MAPSESSIONS Total Sessions
                </span>&nbsp;
                <span class="badge badge-success">
                  ONNN on Network Next
                </span>
              </h1>
              <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1">
                <div class="mr-auto"></div>
                <div class="px-2 hidden">
                  <select class="form-control">
                    <option value="rtt">
                      Sort by RTT
                    </option>
                  </select>
                </div>
                <div class="px-2 hidden">
                  <select class="form-control">
                    <option value="everything">
                      Everything
                    </option>
                    <option value="direct">
                      Direct Only
                    </option>
                    <option value="next">
                      Network Next Only
                    </option>
                  </select>
                </div>
                <div class="px-2 hidden">
                  <select class="form-control">
                    <option value="all">
                      All
                    </option>
                    <option>
                      BUYERNAME ADMIN LOOP
                    </option>
                    <option>
                      BUYERNAME SINGLE BUYER
                    </option>
                  </select>
                </div>
              </div>
            </div>
            <div class="spinner-border" role="status" id="sessions-spinner" v-if="false">
              <span class="sr-only">Loading...</span>
            </div>
              <div class="table-responsive table-no-top-line" v-if="true">
                  <table class="table table-sm table-striped table-hover">
                      <thead>
                          <tr>
                              <th>
                                  <span>
                                      <!-- No Title -->
                                  </span>
                              </th>
                              <th>
                                  <span>
                                      Session ID
                                  </span>
                              </th>
                              <th v-if="false">
                                  <span>
                                      User Hash
                                  </span>
                              </th>
                              <th>
                                  <span>
                                      ISP
                                  </span>
                              </th>
                              <th>
                                  <span>
                                      Datacenter
                                  </span>
                              </th>
                              <th class="text-right">
                                  <span>
                                      Direct RTT
                                  </span>
                              </th>
                              <th class="text-right">
                                  <span>
                                      Next RTT
                                  </span>
                              </th>
                              <th class="text-right">
                                  <span>
                                      Improvement
                                  </span>
                              </th>
                          </tr>
                      </thead>
                      <tbody>
                          <!-- <tr v-for="(session, index) in pages.sessions.sessions"> -->
                            <tr>
                              <td>
                                  <font-awesome-icon icon="circle"
                                                     class="fa-w-16 fa-fw text-success"
                                                     v-if="true"
                                  />
                                  <font-awesome-icon icon="circle"
                                                     class="fa-w-16 fa-fw text-primary"
                                                     v-if="false"
                                  />
                              </td>
                              <td>
                                  <a class="text-dark fixed-width" href="#">
                                      SESSION ID
                                  </a>
                              </td>
                              <td v-if="false">
                                  <a class="text-dark fixed-width" href="#">
                                      SESSION USER HASH
                                  </a>
                              </td>
                              <td>
                                  SESSION LOCATION ISP
                              </td>
                              <td>
                                  <span class="text-dark">
                                      SESSION DATACENTER
                                  </span>
                              </td>
                              <td class="text-right ">
                                  SESSION DIRECT RTT
                              </td>
                              <td class="text-right ">
                                  SESSION NEXT RTT
                              </td>
                              <td class="text-right ">
<!-- <span v-if="session.delta_rtt > 0 && session.on_network_next"
        v-bind:class="{
            'text-success': session.delta_rtt >= 5,
            'text-warning': session.delta_rtt >= 2 && session.delta_rtt < 5,
            'text-danger': session.delta_rtt < 2 && session.delta_rtt > 0
        }"
> -->
                                  <span>
                                      <b>
                                          SESSION DELTA RTT
                                      </b>
                                  </span>
<!-- <span v-if="session.delta_rtt < 0 || !session.on_network_next">
    <b>
        -
    </b>
</span> -->
                              </td>
                          </tr>
                      </tbody>
                  </table>
              </div>
          </div>
        </main>
        <main role="main" class="col-md-12 col-lg-12 px-4" v-if="false">
          <div class="
                  d-flex
                  justify-content-between
                  flex-wrap
                  flex-md-nowrap
                  align-items-center
                  pt-3
                  pb-2
                  mb-3
                  border-bottom
                "
          >
            <h1 class="h2">
              Session Tool
            </h1>
            <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
              <div class="mr-auto"></div>
            </div>
          </div>
          <form class="flow-stats-form">
            <div class="form-group">
              <label for="session-id-input">
                  Session ID
              </label>
              <div class="row">
                <div class="col">
                  <input class="form-control"
                        type="text"
                        placeholder="Enter a Session ID to view statistics"
                  >
                </div>
                <div class="col-auto">
                  <button class="btn btn-primary" type="submit">
                    View Stats
                  </button>
                </div>
              </div>
            </div>
          </form>
          <div class="alert alert-info" role="alert" id="session-tool-alert" v-if="false">
            Please enter a valid Session ID to view its statistics.
            It should be a hexadecimal number (with leading zeros), or a decimal number.
          </div>
          <div class="alert alert-danger" role="alert" id="session-tool-danger" v-if="false">
            Failed to fetch session details
          </div>
          <div id="session-details">
            <div v-if="true">
              <div class="row">
                <div class="col-12 col-lg-8">
                  <div class="card mb-2">
                    <div class="card-header">
                      <strong>
                        Latency
                      </strong>
                      <div class="float-right">
                        <span class="mr-2"
                              style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
                        ></span>
                        <span style="color: rgb(0, 109, 44);">
                          — Network Next
                        </span>
                        <span style="color: rgb(49, 130, 189);">
                          — Direct
                        </span>
                        <span></span>
                      </div>
                    </div>
                    <div class="card-body">
                        <div id="latency-chart-1"></div>
                    </div>
                  </div>
                  <div class="card mb-2">
                    <div class="card-header">
                      <strong>
                        Jitter
                      </strong>
                      <div class="float-right">
                        <span class="mr-2"
                              style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
                        ></span>
                        <span style="color: rgb(0, 109, 44);">
                          — Network Next
                        </span>
                        <span style="color: rgb(49, 130, 189);">
                          — Direct
                        </span>
                        <span></span>
                      </div>
                    </div>
                    <div class="card-body">
                      <div id="jitter-chart-1"></div>
                    </div>
                  </div>
                  <div class="card mb-2">
                    <div class="card-header">
                      <strong>
                        Packet Loss
                      </strong>
                      <div class="float-right">
                        <span class="mr-2"
                              style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
                        ></span>
                        <span style="color: rgb(0, 109, 44);">
                            — Network Next
                        </span>
                        <span style="color: rgb(49, 130, 189);">
                            — Direct
                        </span>
                        <span></span>
                      </div>
                    </div>
                    <div class="card-body">
                      <div id="packet-loss-chart-1"></div>
                    </div>
                  </div>
                  <div class="card mb-2">
                    <div class="card-header">
                      <strong>
                        Bandwidth
                      </strong>
                      <div class="float-right">
                        <span class="mr-2"
                              style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
                        ></span>
                        <span style="color: orange;">
                          — Up
                        </span>
                        <span style="color: blue;">
                          — Down
                        </span>
                        <span></span>
                      </div>
                      <div class="float-right">
                        <span id="traffic-up-legend"></span>
                        <span id="traffic-down-legend"></span>
                      </div>
                    </div>
                    <div class="card-body">
                      <div id="bandwidth-chart-1"></div>
                    </div>
                  </div>
                </div>
                <div class="col-12 col-lg-4">
                  <div class="card">
                    <div class="card-img-top">
                      <div id="session-tool-map"
                          style="
                                width: 100%;
                                height: 40vh;
                                margin: 0px;
                                padding: 0px;
                                position: relative;
                          "
                      ></div>
                    </div>
                    <div class="card-body">
                      <div class="card-text">
                        <dl>
                          <dt>
                            ISP
                          </dt>
                          <dd>
                            <em>
                              META LOCATION ISP
                            </em>
                          </dd>
                          <div v-if="false">
                            <dt>
                              User Hash
                            </dt>
                            <dd>
                              <a class="text-dark">
                                META USER HASH
                              </a>
                            </dd>
                          </div>
                          <dt>
                              User IP Address
                          </dt>
                          <dd>
                              META CLIENT ADDR
                          </dd>
                          <dt>
                              Platform
                          </dt>
                          <dd>
                              META PLATFORM
                          </dd>
                          <dt v-if="false">
                              Customer
                          </dt>
                          <dd v-if="false">
                              BUYER NAME
                          </dd>
                          <dt>
                            SDK Version
                          </dt>
                          <dd>
                            META SDK
                          </dd>
                          <dt>
                            Connection Type
                          </dt>
                          <dd>
                            META CONNECTION
                          </dd>
                          <dt v-if="false">
                              Nearby Relays
                          </dt>
                          <dd v-if="false">
                              No Nearby Relays
                          </dd>
                          <table class="table table-sm mt-1" v-if="false">
                            <thead>
                              <tr>
                                <th style="width: 50%;">
                                  Name
                                </th>
                                <th style="width: 16.66%;">
                                  RTT
                                </th>
                                <th style="width: 16.66%;">
                                  Jitter
                                </th>
                                <th style="width: 16.66%;">
                                  Packet Loss
                                </th>
                              </tr>
                            </thead>
                            <tbody>
                                <!-- <tr v-for="relay in pages.sessionTool.meta.nearby_relays"> -->
                                <tr>
                                  <td>
                                    <a class="text-dark">NAME</a>&nbsp;
                                  </td>
                                  <td>
                                    CLIENT STATS RTT
                                  </td>
                                  <td>
                                    CLIENT STATS JITTER
                                  </td>
                                  <td>
                                    CLIENT STATS PACKETLOSS%
                                  </td>
                                </tr>
                            </tbody>
                          </table>
                          <dt  v-if="false">
                              Route
                          </dt>
                          <table class="table table-sm mt-1" v-if="false">
                            <thead>
                              <tr>
                                <th style="width: 50%;">
                                  Name
                                </th>
                                <th style="width: 50%;">
                                  Role
                                </th>
                              </tr>
                            </thead>
                            <tbody>
                              <tr>
                                  <td>
                                      META CLIENT ADDR
                                  </td>
                                  <td>
                                      <em>
                                          User (Player)
                                      </em>
                                  </td>
                              </tr>
          <!-- <tr v-for="(hop, index) in pages.sessionTool.meta.hops" scope="row"> -->
                              <tr>
                                <td>
                                    NAME
                                </td>
                                <td>
                                    Hop INDEX + 1
                                </td>
                              </tr>
                              <tr>
                                <td>
                                  META SERVER ADDR
                                </td>
                                <td>
                                    <em>Destination Server</em>
                                </td>
                              </tr>
                            </tbody>
                          </table>
                        </dl>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </main>
        <main role="main" class="col-md-12 col-lg-12 px-4" v-if="false">
          <div class="
                    d-flex
                    justify-content-between
                    flex-wrap
                    flex-md-nowrap
                    align-items-center
                    pt-3
                    pb-2
                    mb-3
                    border-bottom
                  "
          >
            <h1 class="h2">
              User Tool
            </h1>
            <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
              <div class="mr-auto"></div>
            </div>
          </div>
          <form class="flow-stats-form">
            <div class="form-group">
              <label for="user-hash-input">
                  User Hash
              </label>
              <div class="row">
                <div class="col">
                  <input class="form-control"
                         type="text"
                         placeholder="Enter a User Hash to view statistics"
                  >
                </div>
                <div class="col-auto">
                  <button class="btn btn-primary" type="submit">
                    View Sessions
                  </button>
                </div>
              </div>
            </div>
          </form>
          <div class="alert alert-info" role="alert" id="user-tool-alert" v-if="false">
              Please enter a User ID or Hash to view their sessions.
          </div>
          <div class="alert alert-danger" role="alert" id="user-tool-danger" v-if="false">
              Failed to fetch user sessions
          </div>
          <div id="user-sessions">
            <div class="table-responsive table-no-top-line">
              <table class="table table-sm" v-if="true">
                <thead>
                  <tr>
                    <th>
                      <span>
                        Session ID
                      </span>
                    </th>
                    <th>
                      <span>
                        Platform
                      </span>
                    </th>
                    <th>
                      <span>
                        Connection Type
                      </span>
                    </th>
                    <th>
                      <span>
                        ISP
                      </span>
                    </th>
                    <th>
                      <span>
                        Datacenter
                      </span>
                    </th>
                    <th>
                      <span>
                        Server Address
                      </span>
                    </th>
                  </tr>
                </thead>
                <tbody v-if="true">
                  <!-- <tr v-for="session in pages.userTool.sessions"> -->
                  <tr>
                    <td>
                      <a class="text-dark fixed-width" href="#">
                          SESSION ID
                      </a>
                    </td>
                    <td>
                      SESSION PLATFORM
                    </td>
                    <td>
                      SESSION CONNECTION
                    </td>
                    <td>
                      SESSION LOCATION ISP
                    </td>
                    <td>
                      SESSION DATACENTER
                    </td>
                    <td>
                      SESSION SERVER ADDR
                    </td>
                  </tr>
                </tbody>
                <tbody v-if="false">
                  <tr>
                    <td colspan="7" class="text-muted">
                        There are no sessions belonging to this user.
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </main>
        <main role="main" class="col-md-12 col-lg-12 px-4" v-show="false">
          <div class="
                    d-flex
                    justify-content-between
                    flex-wrap
                    flex-md-nowrap
                    align-items-center
                    pt-3
                    pb-2
                    mb-3
                    border-bottom
                  "
          >
            <h1 class="h2">
              Downloads
            </h1>
            <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
              <div class="mr-auto"></div>
            </div>
          </div>
          <div class="card mb-2">
            <div class="card-body">
              <div class="btn-group-vertical btn-group-sm float-right">
                <a
                  href="#"
                  onclick="window.open('https://storage.googleapis.com/portal_sdk_download_storage_prod/next-3.4.6.zip')"
                  class="btn btn-primary m-1"
                  title="SDK v3.4.6"
                  style='width: "160px"'
                >
                  <font-awesome-icon icon="download"
                                     class="fa-w-16 fa-fw"
                  />
                  SDK v3.4.6
                </a>
                <a
                  href="#"
                  onclick="window.open('https://network-next-sdk.readthedocs-hosted.com/en/latest/')"
                  class="btn btn-primary m-1"
                  title="v3.4.6 Documentation"
                  style='width: "160px"'
                >
                  <font-awesome-icon icon="download"
                                      class="fa-w-16 fa-fw"
                  />
                  v3.4.6 Documentation
                </a>
            </div>
            <h5 class="card-title">
                Network Next SDK
            </h5>
            </div>
          </div>
        </main>
        <main role="main" class="col-md-12 col-lg-12 px-4" v-if="false">
          <div class="
                    d-flex
                    justify-content-between
                    flex-wrap
                    flex-md-nowrap
                    align-items-center
                    pt-3
                    pb-2
                    mb-3
                    border-bottom
          ">
            <h1 class="h2">
              Settings
            </h1>
            <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
              <div class="mr-auto"></div>
            </div>
          </div>
          <div class="card" style="margin-bottom: 250px;">
            <div class="card-header">
              <ul class="nav nav-tabs card-header-tabs">
                <li class="nav-item">
                  <a class="nav-link" id="accounts-link" href="#" v-bind:class="{ active: true }">
                    Users
                  </a>
                </li>
                <li class="nav-item">
                  <a class="nav-link" id="config-link" href="#" v-bind:class="{ active: false }">
                    Game Configuration
                  </a>
                </li>
              </ul>
            </div>
            <div id="settings-page">
              <div class="card-body" v-if="true">
                <div id="auto-signin" class="hidden">
                  <h5 class="card-title">
                    Automatic sign-in
                  </h5>
                  <p class="card-text">
                    Save time by allowing users with verified email
                    addresses automatic access to your Network Next account.
                  </p>
                  <div id="auto-sign-in-spinner" v-show="false">
                    <div class="spinner-border" role="status">
                      <span class="sr-only">Loading...</span>
                    </div>
                  </div>
                  <form v-show="true">
                    <div class="form-group">
                      <label for="customerId">
                        Automatic Sign-in Domain
                      </label>
                      <input type="text"
                             class="form-control form-control-sm"
                             value="networknext.com"
                             id="auto-sign-in-domain"
                      >
                      <small class="form-text text-muted">
                        If you set this to a domain, such as "example.com",
                        then anyone with a verified email address with
                        this domain will automatically
                        join your account when logging in.
                        You can set this so that you don't need to
                        manually onboard everyone in your organization.
                      </small>
                    </div>
                    <div class="form-group">
                      <label for="customerId">
                        Permission Level
                      </label>
                      <select
                        class="form-control"
                        id="auto-signin-permissions"
                        multiple
                      >
                      </select>
                      <small class="form-text text-muted">
                        The permission level to grant accounts that
                        join your account via automatic sign-in.
                      </small>
                    </div>
                    <button type="submit" class="btn btn-primary btn-sm">
                      Save Automatic Sign-in
                    </button>
                    <p class="text-muted text-small mt-2"></p>
                  </form>
                  <hr class="mt-4 mb-4">
                </div>
                <h5 class="card-title">
                  Add new users
                </h5>
                <p class="card-text">
                  Provide other people with access to your account.
                </p>
                <div id="add-user-spinner" v-show="false">
                  <div class="spinner-border" role="status">
                    <span class="sr-only">Loading...</span>
                  </div>
                </div>
                <div class="alert alert-success"
                     role="alert"
                     id="session-tool-alert"
                     v-show="false">
                  NEWUSER SUCCESS
                </div>
                <div class="alert alert-danger"
                     role="alert"
                     id="session-tool-alert"
                     v-show="false">
                  NEWUSER FAILURE
                </div>
                <form v-show="true">
                  <div class="form-group">
                    <label for="customerId">
                      Add users by email address
                    </label>
                    <textarea class="form-control form-control-sm" id="new-user-emails"></textarea>
                    <small class="form-text text-muted">
                      Enter a newline or comma-delimited list of email
                      addresses to add users to your account.
                    </small>
                  </div>
                  <div class="form-group">
                    <label for="customerId">
                      Permission Level
                    </label>
                    <!-- Replace these with a new dropdown component -->
                    <select
                      class="form-control"
                      id="add-user-permissions"
                      placeholder="Select..."
                      multiple
                    >
                    </select>
                    <!-- Replace these with a new dropdown component -->
                    <small class="form-text text-muted">
                      The permission level to grant the added user accounts.
                    </small>
                  </div>
                  <button type="submit" class="btn btn-primary btn-sm">
                    Add Users
                  </button>
                  <p class="text-muted text-small mt-2"></p>
                </form>
                <hr class="mt-4 mb-4">
                <h5 class="card-title">
                  Manage existing users
                </h5>
                <p class="card-text">
                  Manage the list of users that currently have access to your Network Next account.
                </p>
                <div id="account-table-spinner" v-show="false">
                  <div class="spinner-border" role="status">
                    <span class="sr-only">Loading...</span>
                  </div>
                </div>
                <table class="table table-sm mt-4" v-show="true">
                  <thead class="thead-light">
                      <tr>
                          <th style="width: 20%;">
                              Email Address
                          </th>
                          <th style="width: 70%;">
                              Permissions
                          </th>
                          <th style="width: 10%;">
                              Actions
                          </th>
                      </tr>
                  </thead>
                  <tbody>
                    <!-- <tr v-for="(account, index) in pages.settings.accounts"> -->
                    <tr>
                      <div
                        class="alert alert-success"
                        role="alert"
                        v-show="false"
                      >
                        UPDATEUSER SUCCESS
                      </div>
                      <div
                          class="alert alert-danger"
                          role="alert"
                          v-show="false"
                      >
                        UPDATEUSER FAILURE
                      </div>
                      <td>
                        EMAIL
                      </td>
                      <td>
                      <!-- Replace these with a new dropdown component -->
                        <!-- <select
                          class="form-control"
                          :id="`edit-user-permissions-${account.user_id}`"
                          multiple
                        >
                        </select> -->
                      <!-- Replace these with a new dropdown component -->
                      </td>
                      <td class="td-btn" v-show="true">
                        <button
                          class="btn btn-xs btn-primary"
                          data-toggle="tooltip"
                          data-placement="bottom"
                          title="Change this user's permissions"
                        >
                          <font-awesome-icon icon="pen"
                                             class="fa-w-16 fa-fw"
                          />
                        </button>&nbsp;
                        <button
                          class="btn btn-xs btn-danger"
                          data-toggle="tooltip"
                          data-placement="bottom"
                          title="Remove this user"
                        >
                          <font-awesome-icon icon="trash"
                                             class="fa-w-16 fa-fw"
                          />
                        </button>&nbsp;
                      </td>
                      <td class="td-btn" v-show="false">
                        <button
                          class="btn btn-xs btn-success"
                          data-toggle="tooltip"
                          data-placement="bottom"
                          title="Save Changes"
                        >
                          <font-awesome-icon icon="check"
                                             class="fa-w-16 fa-fw"
                          />
                        </button>&nbsp;
                        <button
                          class="btn btn-xs btn-secondary"
                          data-toggle="tooltip"
                          data-placement="bottom"
                          title="Cancel Changes"
                        >
                          <font-awesome-icon icon="times"
                                             class="fa-w-16 fa-fw"
                          />
                        </button>&nbsp;
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="card-body" id="config-page" v-show="false">
                <h5 class="card-title">
                  Game Configuration
                </h5>
                <p class="card-text">
                  Manage how your game connects to Network Next.
                </p>
                <div
                  class="alert alert-success"
                  role="alert"
                  v-show="false"
                >
                  UPDATEKEY SUCCESS
                </div>
                <div
                  class="alert alert-danger"
                  role="alert"
                  v-show="false"
                >
                  UPDATEKEY FAILURE
                </div>
                <!-- NOT SURE THESE ARE USED -->
                <div
                  class="alert alert-success"
                  role="alert"
                  v-show="false"
                >
                  UPGRADE SUCCESS
                </div>
                <div
                  class="alert alert-danger"
                  role="alert"
                  v-show="false"
                >
                  UPGRADE FAILURE
                </div>
                <!-- NOT SURE THESE ARE USED -->
                <form v-on:submit.prevent="updatePubKey()">
                  <div class="form-group" id="pubKey">
                    <label>
                      Company Name
                    </label>
                    <input type="text"
                           class="form-control"
                           placeholder="Enter your company name"
                           id="company-input"
                           :disabled="true"
                    >
                    <br>
                    <label>
                      Public Key
                    </label>
                    <textarea class="form-control"
                              placeholder="Enter your base64-encoded public key"
                              id="pubkey-input"
                              :disabled="true">
                    </textarea>
                  </div>
                  <button type="submit"
                          class="btn btn-primary btn-sm"
                          v-if="false"
                  >
                    Save game configuration
                  </button>
                  <p class="text-muted text-small mt-2"></p>
                </form>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
  @import url('https://fonts.googleapis.com/css?family=Montserrat:400,700');

  .map-container-offset {
    width: 100%;
    height: calc(-160px + 90vh);
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
  }
  .map-container-no-offset {
    width: 100%;
    height: calc(-160px + 100vh);
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
  }
  .uplot .legend.inline th::after {
      content: "";
      vertical-align: middle;
  }
  .choices__list--multiple .choices__item {
      background-color: lightgray;
      border: 1px solid darkgray;
      color: black;
      border-radius: 5px;
  }
  .choices[data-type*="select-multiple"]
  .choices__button, .choices[data-type*="text"] .choices__button {
      border-left: 0px;
      color: black;
  }
  .choices__item .is-highlighted {
      background-color: lightgray;
      border: 1px solid darkgray;
      color: black;
  }
  .fill {
      width: 100%;
      height: 100%;
  }
  .dashboard-title {
      width: 95%;
  }
  .workspace {
      width: 100%;
      justify-content: center;
  }
  .map {
      height: 95%;
      max-height: 1000px;
  }
  .mapboxgl-canvas {
      height: 100%;
      width: 100%;
  }
  .mapboxgl-map {
      height: 100%;
      width: 100%;
  }
  #session-tool-map { position: absolute; top: 0; bottom: 0; width: 100%; }

  .feather {
      width: 16px;
      height: 16px;
      vertical-align: text-bottom;
  }

  .logo-container {
      width: 100%;
      height: 24px;
      margin: 0;
      padding: 0;
      display: flex;
      align-items: center;
      justify-content: left;
  }

  .logo-fit {
      max-height: 24px;
      max-width: 100%;
  }

  .fixed-width {
      font-family: monospace;
      font-size: 120%;
  }

  .hidden {
      display: none;
  }

  div.table-no-top-line th {
      border-top: none !important;
  }

  .mapboxgl-ctrl-bottom-left {
      display: none;
  }

  .mapboxgl-ctrl-bottom-right {
      display: none;
  }

  .h2 {
    font-size: 2rem;
  }

  /*
  * Sidebar
  */

  .sidebar {
      position: fixed;
      top: 0;
      bottom: 0;
      left: 0;
      z-index: 100; /* Behind the navbar */
      padding: 48px 0 0; /* Height of navbar */
      box-shadow: inset -1px 0 0 rgba(0, 0, 0, 0.1);
  }

  .sidebar-sticky {
      position: relative;
      top: 0;
      height: calc(100vh - 48px);
      padding-top: 0.5rem;
      overflow-x: hidden;
      overflow-y: auto; /* Scrollable contents if viewport is shorter than content. */
  }

  @supports ((position: -webkit-sticky) or (position: sticky)) {
      .sidebar-sticky {
          position: -webkit-sticky;
          position: sticky;
      }
  }

  .sidebar .nav-link {
      font-weight: 500;
      color: #333;
  }

  .sidebar .nav-link .feather {
      margin-right: 4px;
      color: #999;
  }

  .sidebar .nav-link.active {
      color: #007bff;
  }

  .sidebar .nav-link:hover .feather,
  .sidebar .nav-link.active .feather {
      color: inherit;
  }

  .sidebar-heading {
      font-size: 0.75rem;
      text-transform: uppercase;
  }

  .navbar .form-control {
      padding: 0.75rem 1rem;
      border-width: 0;
      border-radius: 0;
  }

  .form-control-dark {
      color: #fff;
      background-color: rgba(255, 255, 255, 0.1);
      border-color: rgba(255, 255, 255, 0.1);
  }

  .form-control-dark:focus {
      border-color: transparent;
      box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.25);
  }

  .btn-group-xs > .btn,
  .btn-xs {
      padding: 0.25rem 0.4rem;
      font-size: 0.875rem;
      line-height: 0.5;
      border-radius: 0.2rem;
  }

  td.td-btn {
      padding-top: 0.15rem;
      padding-bottom: 0;
  }

  td.td-btn > a.btn-xs {
      padding-bottom: calc(0.25rem + 1px);
  }
</style>
<script lang="ts">
import Vue from 'vue';

import {
  faCheck,
  faCircle,
  faDownload,
  faPen,
  faTimes,
  faTrash,
} from '@fortawesome/free-solid-svg-icons';
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';

const ICONS = [
  faCheck,
  faCircle,
  faDownload,
  faPen,
  faTimes,
  faTrash,
];

library.add(...ICONS);

Vue.component('font-awesome-icon', FontAwesomeIcon);
export default Vue.extend({
  beforeCreate: () => {
    Vue.prototype.$apiService.call('BuyersService.TopSessions', {}).then((response: any) => {
      console.log(response);
    });
  },
});

</script>
