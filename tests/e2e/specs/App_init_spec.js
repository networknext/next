// https://docs.cypress.io/api/introduction/api.html

function overrideEndpointsEmpty () {
  cy.window().then((win) => {
    cy.stub(win.app.$apiService, 'fetchTotalSessionCounts').resolves({ result: { direct: 2, next: 2 } })
    cy.stub(win.app.$apiService, 'fetchMapSessions').resolves({ result: { map_points: [] } })
    cy.stub(win.app.$apiService, 'fetchTopSessions').resolves({ result: { sessions: [] } })
  })
}

function overrideEndpointsFull () {
  cy.window().then((win) => {
    cy.stub(win.app.$apiService, 'fetchTotalSessionCounts').resolves({ result: { direct: 499, next: 143 } })
    cy.stub(win.app.$apiService, 'fetchMapSessions').resolves({ result: { map_points: [] } })
    cy.stub(win.app.$apiService, 'fetchTopSessions').resolves({ result: { sessions: [] } })
  })
}

// TODO: Figure out how to test number of map points
// TODO: Figure out how to stub out endpoints earlier. Page loads using non-overridden functions

describe('App Initialization Tests', () => {
  it('checks to see if the portal loaded correctly', () => {
    // Load the page
    cy.visit('/')

    // Override all of the endpoints that will be used
    overrideEndpointsEmpty()

    // check if we are on the right page (router is routing correctly)
    cy.url().should('eq', 'http://127.0.0.1:8080/#/')
    cy.contains('h1', 'Map').should('exist')

    // Check if all the links / buttons are available
    cy.contains('.nav-link', 'Map').should('exist')
    cy.contains('.nav-link', 'Sessions').should('exist')
    cy.contains('.nav-link', 'Session Tool').should('exist')
    cy.get('[data-test="loginButton"').should('have.text', ' Log in ')
    cy.get('[data-test="signUpButton"').should('have.text', ' Sign up ')

    // Check for main features of page
    cy.contains('span', ' Total Sessions').should('exist')
    cy.contains('span', ' on Network Next').should('exist')

    // Check that endpoints are populating features correctly
    cy.get('[data-test="totalSessions"').should('have.text', ' 4 Total Sessions ')
    cy.get('[data-test="nnSessions"').should('have.text', ' 2 on Network Next ')

    // Switch to sessions page and see what that looks like
    cy.get('[data-test="sessionsLink"').click()
    cy.url().should('eq', 'http://127.0.0.1:8080/#/sessions')

    // Check for main features of page
    cy.contains('h1', 'Sessions').should('exist')
    cy.contains('span', ' Total Sessions').should('exist')
    cy.contains('span', ' on Network Next').should('exist')

    // Check that endpoints are populating features correctly
    cy.get('[data-test="totalSessions"').should('have.text', ' 4 Total Sessions ')
    cy.get('[data-test="nnSessions"').should('have.text', ' 2 on Network Next ')

    // Check out the table
    cy.get('.table').contains('span', 'Session ID').should('be.visible')
    cy.get('.table').contains('span', 'User Hash').should('not.exist')
    cy.get('.table').contains('span', 'ISP').should('be.visible')
    cy.get('.table').contains('span', 'Datacenter').should('be.visible')
    cy.get('.table').contains('span', 'Direct RTT').should('be.visible')
    cy.get('.table').contains('span', 'Next RTT').should('be.visible')
    cy.get('.table').contains('span', 'Improvement').should('be.visible')

    // Switch to session tool page and see what that looks like
    cy.get('[data-test="sessionToolLink"').click()
    cy.url().should('eq', 'http://127.0.0.1:8080/#/session-tool')

    // Check for main features of the page
    cy.contains('h1', 'Session Tool').should('exist')
    cy.contains('label', 'Session ID').should('exist')
    cy.contains('button', 'View Stats').should('exist')
    cy.get('[data-test="searchInput"').should('exist').should('have.attr', 'placeholder', 'Enter a Session ID to view statistics')

    // Return back to the map
    cy.get('[data-test="mapLink"').click()

    // check if we are on the right page (router is routing correctly)
    cy.url().should('eq', 'http://127.0.0.1:8080/#/')
    cy.contains('h1', 'Map').should('exist')
  })
})
