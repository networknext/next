// https://docs.cypress.io/api/introduction/api.html

describe('My First Test', () => {
  it('Visits the app root url', () => {
    cy.visit('/')
    cy.contains('h1', 'Welcome to Your Vue.js + TypeScript App')
  })
})

describe('App Load Test', () => {
  it('Mounts the Vue app and loads the session map workspace', () => {
    cy.server()
    cy.visit('/')
  })
})
