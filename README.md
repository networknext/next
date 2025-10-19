<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

[![Build Status](https://networknext.semaphoreci.com/badges/next/branches/main.svg?style=shields&key=b74eb8a5-60a2-4044-a0db-cfeab84798dc)](https://networknext.semaphoreci.com/projects/next)

Network Next is a **network accelerator** for multiplayer games. 

It players by sending game traffic across relay network when the default internet route has high latency or packet loss.

It's currently live worldwide with the football game [REMATCH](https://www.playrematch.com). More details [https://mas-bandwidth.com/rematch-accelerated-by-network-next/](here).

It's source available and free to use when you host the software for your own game with less than 10k peak CCU. 

For games with 10k-100k peak CCU, the license fee is $25k USD per-game, per-year.

For games with more than 100k peak CCU, the license fee is $100k USD per-game, per-year.

We also offer a hosted solution with commercial support. Please [contact us](mailto:glenn@networknext.com) for details.

# Install Network Next

1. [Fork the network next repository on github](docs/fork_next_repository.md)
2. [Run a local instance with docker compose](docs/run_local_instance_with_docker_compose.md)
3. [Setup your local machine for development](docs/setup_your_local_machine_for_development.md)
4. [Setup prerequisites](docs/setup_prerequisites.md)
5. [Configure network next](docs/configure_network_next.md)
6. [Create google cloud projects with terraform](docs/create_google_cloud_projects_with_terraform.md)
7. [Setup Semaphore CI to build and deploy artifacts](docs/setup_semaphore_ci_to_build_and_deploy_artifacts.md)
8. [Deploy to Development](docs/deploy_to_development.md)
9. [Deploy to Staging](docs/deploy_to_staging.md)
10. [Deploy to Production](docs/deploy_to_production.md)
11. [Tear down dev, staging and production](docs/tear_down_dev_staging_and_production.md)

# Customize your Installation

1. [Spin dev back up](docs/spin_dev_back_up.md)
2. [Disable the raspberry clients](docs/disable_the_raspberry_clients.md)
3. [Connect a client to the test server](docs/connect_a_client_to_the_test_server.md)
4. [Modify route shader for test buyer](docs/modify_route_shader_for_test_buyer.md)
5. [Move test server to Sao Paulo](docs/move_test_server_to_sao_paulo.md)
6. [Enable acceleration to Sao Paulo](docs/enable_acceleration_to_sao_paulo.md)
7. [Modify set of google relays](docs/modify_set_of_google_relays.md)
8. [Modify set of amazon relays](docs/modify_set_of_amazon_relays.md)
9. [Modify set of akamai relays](docs/modify_set_of_akamai_relays.md)
10. [Spin up other relays in Sao Paulo](docs/spin_up_other_relays_in_sao_paulo.md)
11. [Spin up relays near you](docs/spin_up_relays_near_you.md)
12. [Test acceleration to Sao Paulo](docs/test_acceleration_to_sao_paolo.md)

# Integrate with your Game

1. [Create your own buyer](docs/create_your_own_buyer.md)
2. [Customize your route shader](docs/customize_your_route_shader.md)
3. [Run your own client and server](docs/run_your_own_client_and_server.md)
4. [Integrate with your game](docs/integrate_with_your_game.md)

# Set up your Production Environment

* [Getting production ready for your game](docs/getting_production_ready_for_your_game.md)
* [Planning your production relay fleet](docs/planning_your_production_relay_fleet.md)

# Reference Documentation

* [Portal user guide](docs/portal_user_guide.md)
* [Next tool user guide](docs/next_tool_user_guide.md)
* [Network Next Terraform provider](docs/network_next_terraform_provider.md)
* [Datacenter and relay naming conventions](docs/datacenter_and_relay_naming_conventions.md)
* [BigQuery table schemas](docs/bigquery_table_schemas.md)
* [Glossary of Common Terms](docs/glossary_of_common_terms.md)
