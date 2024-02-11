<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Glossary of Common Terms

* **Acceleration** - _Sending traffic for a session across the relay fleet to reduce latency, jitter and packet loss for a session._

* **API** - _Application Programming Interface. In this context, the Network Next API is a service in the Network Next backend that lets you call CRUD (create-read-update-delete) operations on the Postgres SQL database, as well as supporting the portal with REST APIs that give it access to the data it needs to display. You can get an API token and call the Network Next API yourself from your own website to get data from your Network Next environment._

* **BigQuery** - _BigQuery is a Google Cloud database technology that stores large volumes of row oriented data, and allows querying this data with an SQL-like language. It is used to store data about the operation of the Network Next system so you can perform data science and analytics queries on your players network performance, the performance of your relay fleet and any other part of your Network Next environment._

* **Buyer** - _A buyer is effectively a game or application in your Network Next environment. Buyers purchase traffic from sellers, such as "google", "amazon" and any other company that provides relays, these are "sellers" who sell traffic to your buyers. Forgive, me, Network Next was originally a two-sided marketplace prior to becoming the turnkey product that you are using today._

* **CCU** - _Concurrent Users. Or alternatively, the total number of sessions running at any time._

* **Continue Token** - _When the Network Next route does not change from one slice to another, we send a continue token, which simply extends the route across relays for 10 more seconds, without the bandwidth expense of fully redescribing the route in terms of the linked list of IP addresses and ports between the client and server._

* **Cost Matrix** - _The first step of route optimization in Network Next is for all relay pings to be turned into a triangular cost matrix, which provides a constant time lookup of the cost (in terms of latency, or round trip time (rtt) - in milliseconds), between any two relays._

* **Database** - _Generally when we refer to the "database" in Network Next, it is the Postgres SQL database that stores all the configuration of your Network Next environment. For example, the database stores the set of datacenters, relays, buyers, and sellers. It does not store any runtime data, nor does the relay backend or server backend directly talk to the Postgres SQL database. Instead, the runtime components in the Network Next backend use a database.bin file, which is an extracted binary format of the Network Next configuration, so it can continue to run even if Postgres SQL is down._

* **database.bin** - _The Postgres SQL database is not directly used by the runtime components of the Network Next backend, instead a database.bin file, which is a Golang "GOB" file format datastructure saved to a file, of all data in the Postgres SQL database. This database.bin is then uploaded or "committed" to the backend, and it becomes active for the relay backend and server backend to get their configuration from. This way we make the backend resilient so that it can continue to run even if Postgres SQL is down._

* **Datacenter** - _A datacenter in the Network Next system is a physical location, owned by a seller that relays and servers can logically reside within._

* **Destination Relay** - _Network Next requires a relay in the same physical datacenter as your server, so that we can accelerate to the correct place, assuming that the cost between any server in the datacenter and the destination relay is <= 1ms. For example, if you had your server in google.virginia.1, then traffic would only be accelerated to that server if there was also a destination relay in the datacenter google.virginia.1, and that datacenter was enabled for the buyer._

* **Direct Route** - _The default route that traffic takes across the internet when packets are sent between client and server directly, eg. packets sent from the client IP address and port, to the server IP address and port, and vice versa in the other direction._

* **Environment** - _An instance of Network Next. For example, your "development environment" and your "production environment" are two separate environments, each with their own relay fleet._

* **Fallback to Direct** - _When the SDK is unable to continue the Network Next route for any reason, for example if a relay suddenly crasher, or the server backend suddenly went away, or the relay backend was unable to provide a route matrix. The SDK "falls back to direct" automatically, so that the session continues in a non-accelerated mode. This way even if the Network Next backend is completely down, sessions can still play and connect from client to server._

* **Google Cloud Storage** - _It's a standard feature of Google Cloud. You can upload files to google cloud storage buckets, and they will be stored there, and made available to the deployment process in Google Cloud, or made publicly available for download as necessary._

* **Google PubSub** - _Google PubSub is a persistent message system. Parts of the Network Next backend publish Google PubSub messages to Google Cloud, which processes these messages and inserts the data contained in those messages into BigQuery. This is a standard feature of Google Cloud, and is the best and cheapest way to insert data into BigQuery in close to real-time._

* **ip2location** - _Network Next looks up an approximate location as (latitude, longitude) pair using the ip2location service maxmind.com_

* **Jitter** - _The amount of time variance in packet delivery time. When jitter is happening some packets arrive late. Jitter is often expressed in milliseconds units, representing some measurement of the amount of jitter in milliseconds._ 

* **Near Relays** - _Network Next gets the approximate latitude and longitude of the player and looks up (ideally) 16 relays that are physically close to the player. For example, a player in Helsinki, Finland would ideally find 16 different near relays, all in Helsinki, Finland, or relatively close to it, perhaps in Stockholm, Sweden as well. These near relays are then pinged for 10 seconds at the start of each session, to determine the initial hop costs onto the relays. This initial hop cost is then added to the total cost for all routes in the route matrix the destination relay, to determine the total route cost from client to server. If 16 near relays cannot be found within some threshold distance, the system falls back to searching for 16 relays close to the destination relay._

* **Network Next Backend** - _Refers to the entire backend of Network Next that runs in Google Cloud. It is written in Golang._

* **Next Route** - _The route that traffic takes between the client and server across relays in your relay fleet._

* **next tool** - _There is a command line tool that will help you maintain and operate your Network Next environment at the root of the source code repository. To run it and get a list of commands just go `cd ~/next && next`._

* **Node** - _In the Network Next system, a node is any point along a route, including a client, a relay or a server._

* **Portal** - _The portal is a website where you can view your Network Next environment in real-time._

* **Redis** - _Redis is an open source in-memory database that is used internally by the Network Next backend for inter-service communication, and especially by the portal to store and query data about sessions, servers, and so on._

* **Redis Cluster** - _When a single Redis database instance is not sufficient in terms of memory, throughput or CPU, you can split up your data across a Redis Cluster and scale horizontally. Staging and Production environments in Network Next use Redis cluster so they are able to scale to 10 million CCU._

* **Relay** - _A linux machine running the Network Next relay software. It acts as a software router for packets sent between the client and server, when a session is accelerated by Network Next._

* **Relay Backend** - _The relay backend is the component that relays in your relay fleet talks to, to upload their pings to all other relays in your fleet. It then takes these ping every second and creates a cost matrix, then runs route optimization to turn this into a route matrix. The route matrix is then passed to the server backend, so it can perform routing for sessions._

* **Relay Fleet** - _The set of all relays in your Network Next environment_.

* **Relay Pings** - _All relays are continuously pinging every other relay in your system 10 times per-second. This is how we know the cost in milliseconds to travel from one relay to another. These relay pings are the very first input into the route optimization process._

* **Route Diversity** - _The idea that you can get a better connection if you have many different options, or routes to get from the client to server to choose from._

* **Route Matrix** - _A route matrix is the output of the route optimization process, providing a constant time lookup of the set of routes available from one relay to another, sorted in least to highest RTT, such that all routes have a lower RTT than a packet sent directly from one relay to another. It is used by the server backend to look up high quality routes from one relay to another in constant time._

* **Route Token** - _Network Next cryptographically stores and describes the accelerated route between client and server as a series of "route tokens". Each route token has a link to the next address and port, and the previous address and port, and is encrypted and signed in such a way that it can only be read by the node that it belongs to. This way the Network Next backend is able to perform routing and tell the SDK which relays to send a client across securely, and in such a way that the client only knows the IP address of the first hop in the route._

* **RTT** - _Round trip time. This is time it takes for a packet to travel from the client to the server and then back down to the client. It's usually expressed in milliseconds._

* **SDK** - _Software Development Kit. The Network Next software that integrates with your game client and game server and takes over sending and receiving UDP packets on behalf of your game._

* **Seller** - _A seller is any company that provides relays in your Network Next environment. For example: "google", "amazon" and "akamai" are sellers._

* **Semaphore CI** - _We use the Semaphore CI (continuous integration) build tool to build and upload artifacts and run tests for Network Next, and perform deployments to the backend. You can learn more at https://semaphoreci.com_

* **Server Backend** - _The server backend refers to the service in the Network Next backend that the game server talks to (via the SDK), in order to upload near relay ping results from the client for the session, and get the current route that the session should take (if any), every 10 seconds. It's the part that creates routes for sessions across your relay fleet._

* **Session** - _A session is a connection between a client and server across the Network Next environment_.

* **Slice** - _A 10 second period of a session. We consider sessions to be made up of 10 second long "slices". Each slice corresponds to one data point in the portal, and one row written to bigquery for a session._

* **Terraform** - _Terraform is a really cool system for querying and mutating CRUD (create-read-update-delete) REST APIs. You can configure Network Next Postgres SQL database, and configure and spin up your relay fleet using Terraform. To learn more go to https://terraform.io_

* **UDP** - _User datagram protocol. A low level way of sending data across the internet that is suitable for real-time applications like games._

* **Upgrade** - _A server upgrades a client connection, so that it will be monitored and potentially accelerated in the Network Next system. For example, the server upgraded the client after it connected and completed authentication_.

* **User Hash** - _A 64 bit The user hash is the FNV1a 64bit hash of the user id string passed in to the SDK on the server when the client session is upgraded_.

* **User Id** - _A string that uniquely identifies the user of a game. For example, a Playstation Id, Steam Id or your own unique identifier. User Ids are passed in when the server upgrades a session, but they are never sent up to the Network Next backend, due to GDPR risk, only the hash of the user id is ever sent and stored by your Network Next environment._

* **Veto** - _Veto is the idea that if for any reason a Network Next session experiences some error and leaves Network Next, it won't try to return to acceleration until the player plays a new game, and starts a new session. We use veto to ensure that players never get caught in a loop of being accelerated, dropping off acceleration for some reason, and then going back on to acceleration, and dropping off again... over and over._

[Back to main documentation](../README.md)
