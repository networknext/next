// https://docs.cypress.io/api/introduction/api.html

describe('App Load Test', () => {
  it('Mounts the Vue app and loads the session map workspace', () => {
    cy.server()
    cy.visit('/')
  })
})
