import { PortalService } from './oto.gen.js'

const portalService = new PortalService()

portalService.relays({})
  .then((response) => {
    document.getElementById('relays').innerText = JSON.stringify(response.relays)
  })
  .catch((e) => console.log(e))