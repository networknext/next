import { createRouter, createWebHistory } from 'vue-router'
import MapView from '../views/MapView.vue'
import SessionsView from '../views/SessionsView.vue'
import SessionView from '../views/SessionView.vue'
import UserView from '../views/UserView.vue'
import ServersView from '../views/ServersView.vue'
import ServerView from '../views/ServerView.vue'
import RelaysView from '../views/RelaysView.vue'
import RelayView from '../views/RelayView.vue'
import DatacentersView from '../views/DatacentersView.vue'
import DatacenterView from '../views/DatacenterView.vue'
import BuyersView from '../views/BuyersView.vue'
import BuyerView from '../views/BuyerView.vue'
import SellersView from '../views/SellersView.vue'
import SellerView from '../views/SellerView.vue'
import AdminView from '../views/AdminView.vue'

const routes = [
  {
    path: '/',
    name: 'index',
    component: MapView
  },
  {
    path: '/map',
    name: 'map',
    component: MapView
  },
  {
    path: '/sessions/:page?',
    name: 'sessions',
    component: SessionsView,
  },
  {
    path: '/session/:id', 
    name: 'session',
    component: SessionView 
  },    
  {
    path: '/user/:id/:page?', 
    name: 'user',
    component: UserView 
  },    
  {
    path: '/servers/:page?',
    name: 'servers',
    component: ServersView
  },
  {
    path: '/server/:id', 
    name: 'server',
    component: ServerView
  },    
  {
    path: '/relays/:page?',
    name: 'relays',
    component: RelaysView
  },
  {
    path: '/relay/:id',
    name: 'relay',
    component: RelayView
  },
  {
    path: '/datacenters/:page?',
    name: 'datacenters',
    component: DatacentersView
  },
  {
    path: '/datacenter/:id', 
    name: 'datacenter',
    component: DatacenterView
  },    
  {
    path: '/buyers/:page?',
    name: 'buyers',
    component: BuyersView
  },
  {
    path: '/buyer/:id', 
    name: 'buyer',
    component: BuyerView
  },    
  {
    path: '/sellers/:page?',
    name: 'sellers',
    component: SellersView
  },
  {
    path: '/seller/:id', 
    name: 'seller',
    component: SellerView
  },    
  {
    path: '/admin', 
    name: 'admin',
    component: AdminView
  },    
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
