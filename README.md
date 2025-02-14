<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

Network Next is a **network accelerator** for multiplayer games. 

It works by monitoring player connections and sending game traffic down an optimized route across a relay network when the default internet route has high latency, jitter or packet loss.

You control the settings to decide when to accelerate your players. Set acceptable latency, jitter and packet loss values for your game and only accelerate players who are above your acceptable values. Also, set set how much latency reduction is required in milliseconds before accelerating a player. 

For example, accelerate players only if their packet loss is above above 1%, their jitter is above 10ms, _or_ their latency is above 50 milliseconds round trip time _and_ you can reduce their latency by at least 20ms.

This way you can target network acceleration to players who need it the most, while not wasting money accelerating players who already have good network performance!

[![Build Status](https://networknext.semaphoreci.com/badges/next/branches/master.svg?style=shields&key=b74eb8a5-60a2-4044-a0db-cfeab84798dc)](https://networknext.semaphoreci.com/projects/next)

# Installation

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
11. [Tear down staging and production](docs/tear_down_staging_and_production.md)

# Integrating Network Next

1. [Run a test client](docs/run_a_test_client.md)
2. [Create your own buyer](docs/create_your_own_buyer.md)
3. [Run your own client and server](docs/run_your_own_client_and_server.md)
4. [Integrate with your game](docs/integrate_with_your_game.md)
5. [Unreal engine plugin](docs/unreal_engine_plugin.md)

# User Guides

* [Portal user guide](docs/portal_user_guide.md)
* [Next tool user guide](docs/next_tool_user_guide.md)
  
# Operating Network Next

* [Getting production ready for your game](docs/getting_production_ready_for_your_game.md)
* [Planning your production relay fleet](docs/planning_your_production_relay_fleet.md)
* [Operator guide to Google Cloud relays](docs/operator_guide_to_google_cloud_relays.md)
* [Operator guide to AWS relays](docs/operator_guide_to_aws_relays.md)
* [Operator guide to Akamai relays](docs/operator_guide_to_akamai_relays.md)
* [Operator guide to bare metal relays](docs/operator_guide_to_bare_metal_relays.md)
* [Network Next Terraform provider](docs/network_next_terraform_provider.md)
* [Datacenter and relay naming conventions](docs/datacenter_and_relay_naming_conventions.md)

# Miscellaneous

* [BigQuery table schemas](docs/bigquery_table_schemas.md)
* [Glossary of Common Terms](docs/glossary_of_common_terms.md)
