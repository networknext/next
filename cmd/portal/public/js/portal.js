import { PortalService } from './oto.gen.js'

const portalService = new PortalService()

portalService.relays({})
  .then((response) => console.log(response.relays))
  .catch((e) => console.log(e))