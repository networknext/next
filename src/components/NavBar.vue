<template>
  <nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark p-0 shadow">
    <a class="navbar-brand col-sm-3 col-md-2 mr-0" href="#">
        <div class="logo-container">
            <img class="logo-fit" src="../assets/logo.png">
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
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

@Component
export default class NavBar extends Vue {
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .navbar .form-control {
      padding: 0.75rem 1rem;
      border-width: 0;
      border-radius: 0;
  }
</style>
