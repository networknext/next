<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run a local instance with Docker Compose

You can run a Network Next test environment locally with docker compose on any OS.

1. Install Docker from http://docker.com

2. Change into the 'next' directory at the command line where you cloned the source

3. Build the system

   `docker compose build`

4. Bring the system up

   `docker compose up`

After about one minute, you should see output like:

```console
next-client-1           | 0.557106: info: client upgraded to session 3fd88eb577d0826
next-client-1           | 3.449071: info: client pinging 5 near relays
next-client-1           | 3.449525: info: client direct route
next-client-1           | 13.464149: info: client near relay pings completed
next-client-1           | 13.464641: info: client direct route
next-client-1           | 23.483219: info: client direct route
next-client-1           | 33.480910: info: client direct route
next-client-1           | 43.494686: info: client direct route
next-client-1           | 53.525247: info: client direct route
next-client-1           | 63.549006: info: client next route
next-client-1           | 63.549072: info: client multipath enabled
next-client-1           | 73.517053: info: client continues route
next-client-1           | 83.499405: info: client continues route
```

An entire network next backend, relays and test client and server are now running inside docker compose on your local machine.

5. View the portal

_todo: this portion is not ready yet_

Navigate to the network next portal at http://127.0.0.1:40001

You should see (todo: image)

_todo: this portion is not ready yet_

6. Take the system down

   `docker compose down`

_You are now ready to [setup for development](setup_for_development.md)._
