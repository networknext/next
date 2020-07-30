// https://docs.cypress.io/api/introduction/api.html

describe('App Initialization Tests', () => {
  it('checks to see if the portal loaded correctly', () => {
    cy.visit('/')
    cy.window().then((win) => {
      console.log(win.app.$apiService)
      cy.stub(win.app.$apiService, 'fetchTotalSessionCounts').resolves({ result: { direct: 2, next: 2 } })
      cy.stub(win.app.$apiService, 'fetchMapSessions').resolves({ result: { map_points: [] } })
    })
    cy.url().should('eq', 'http://127.0.0.1:8080/#/')
    cy.contains('h1', 'Map')

    // TODO: Figure out how to test number of map points

    cy.contains('span', ' Total Sessions')
    cy.contains('span', ' on Network Next')
  })
})
