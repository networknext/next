// https://docs.cypress.io/api/introduction/api.html

describe('App Load Test', () => {
  it('Mounts the Vue app and loads the session map workspace', () => {
    cy.visit('/', {
      onBeforeLoad (win) {
        cy.stub(win, 'fetch').withArgs(`${process.env.VUE_APP_BASE_URL}/rpc`, {
          method: 'POST',
          headers: {
            Accept: 'application/json',
            'Accept-Encoding': 'gzip',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'BuyersService.TotalSessions',
            params: {},
            id: 'id'
          })
        }).resolves(
          new Cypress.Promise((resolve) => {
            resolve({
              direct: 4,
              next: 4
            })
          }).delay(2000)
        )
      }
    })
  })
})
