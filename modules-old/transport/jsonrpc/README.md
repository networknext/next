#### Structure

`auth.go`:
* Authmiddleware
* Role and user management endpoints
* Endpoint authentification system

`buyer.go`:
* Buyer related endpoints used by the Portal

`config.go`:
* Feature flags to be accessed from the storer

`ops.go`:
* Endpoints setup to be used specifically by the Admin or Next tools

`relay_fleet.go`:
* Endpoints for the Admin tool to allow internal access to endpoints across services (i.e. `/status`, `/relay_dashboard`)
* Endpoints to commit database.bin file to Google Cloud Storage 

`servers.go`:
* Endpoint to access live servers connected to the Server Backends

`*_test.go`:
* Unit tests for the corresponding service (auth, buyer, config, ops)

## Roles
We have 6 different roles that are assigned via Auth0 to each user:
- `Anonymous`: Portal Map, Sessions Tab, Session Tool
- `Anonymous Plus`: `Anonymous` + User Tool
- `Verified`: `Anonymous Plus` + Downloads Tab, Settings Tab
- `Explorer`: `Verified` + Explore Tab
- `Owner`: `Explorer` + Game Configuration & User Management
- `Admin`: All of the above, including Buyer Filter and Admin Tool access

