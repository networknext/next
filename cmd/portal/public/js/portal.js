import Vue from 'https://unpkg.com/vue@2.6.11/dist/vue.esm.browser.min.js';
import { PortalService } from './oto.gen.js'

const portalService = new PortalService()

portalService.relays({})
  .then((response) => {
    new Vue({
      el: '#relays',
      data: {
        relays: response.relays
      }
    })
  })
  .catch((e) => console.log(e))