<template>
  <div class="card-body" id="config-page">
    <h5 class="card-title">
      Game Configuration
    </h5>
    <p class="card-text">
      Manage how your game connects to Network Next.
    </p>
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
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class GameConfiguration extends Vue {
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
