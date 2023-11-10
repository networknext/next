<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run a local instance with Docker Compose

You can run a Network Next test environment locally with docker compose on any OS. This lets you quickly get a test instance of the whole system up and running locally, which you can interact with via a web browser.

1. Install Docker from http://docker.com

2. Change into the 'next' directory at the command line where you cloned the source

3. Build the system

   `docker compose build`

4. Bring the system up

   `docker compose up`

5. View the portal

_todo: this portion is not ready yet_

Navigate to the network next portal at http://127.0.0.1:8080

You should see (todo: image)

_todo: this portion is not ready yet_

6. Take the system down

   `docker compose down`

_You are now ready to [setup for development](setup_for_development.md)._
