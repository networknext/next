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

const routes = [
  {
    path: '/',
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
    path: '/user/:id', 
    name: 'user',
    component: UserView 
  },    
  {
    path: '/servers',
    name: 'servers',
    component: ServersView
  },
  {
    path: '/server/:id', 
    name: 'server',
    component: ServerView
  },    
  {
    path: '/relays',
    name: 'relays',
    component: RelaysView
  },
  {
    path: '/relay/:id',
    name: 'relay',
    component: RelayView
  },
  {
    path: '/datacenters',
    name: 'datacenters',
    component: DatacentersView
  },
  {
    path: '/datacenter/:id', 
    name: 'datacenter',
    component: DatacenterView
  },    
  {
    path: '/buyers',
    name: 'buyers',
    component: BuyersView
  },
  {
    path: '/buyer/:id', 
    name: 'buyer',
    component: BuyerView
  },    
  {
    path: '/sellers',
    name: 'sellers',
    component: SellersView
  },
  {
    path: '/seller/:id', 
    name: 'seller',
    component: SellerView
  },    
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
